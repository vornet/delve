package proc

// #include "proc_windows.h"
// #include "windows.h"
// #include <stdlib.h>
import "C"
import (
	"debug/pe"
	"debug/gosym"
	"fmt"
	"os"
	"sync"
	sys "golang.org/x/sys/windows"
	"path/filepath"
	"os/exec"
	"syscall"
	
	"github.com/derekparker/delve/dwarf/line"
	"github.com/derekparker/delve/dwarf/frame"
)

// Windows specific information.
type OSProcessDetails struct {
	hProcess	C.HANDLE
}

// Create and begin debugging a new process.
func Launch(cmd []string) (*Process, error) {
	argv0Go, err := filepath.Abs(cmd[0])
	if err != nil {
		return nil, err
	}
	// Make sure the binary exists.
	if filepath.Base(cmd[0]) == cmd[0] {
		if _, err := exec.LookPath(cmd[0]); err != nil {
			return nil, err
		}
	}
	if _, err := os.Stat(argv0Go); err != nil {
		return nil, err
	}

	argv0, _ := syscall.UTF16PtrFromString(argv0Go)

	pi := new(sys.ProcessInformation)
	si := new(sys.StartupInfo)
	err = sys.CreateProcess(argv0, nil, nil, nil, true, 2, nil, nil, si, pi)

	if err != nil {
		return nil, err
	}
	defer sys.CloseHandle(sys.Handle(pi.Thread))
	
	dbp := New(int(pi.ProcessId))
	
	var hProcess C.HANDLE
	var hThread C.HANDLE
	var threadID C.int
	
	var res C.int
	dbp.execPtraceFunc(func() {
		res, err = C.waitForCreateProcessEvent(&hProcess, &hThread, &threadID)
	})
	if res != 0 {
		return nil, err 	
	}
	dbp.os.hProcess = hProcess
	_, err = dbp.addThread(hThread, int(threadID), false)
	if err != nil {
		return nil, err
	}
	
	return initializeDebugProcess(dbp, argv0Go, false)
}

// Attach to an existing process with the given PID.
func Attach(pid int) (*Process, error) {
	fmt.Println("Attach")
	return nil, fmt.Errorf("Not implemented: Attach")
}

func (dbp *Process) Kill() (err error) {
	fmt.Println("Kill")
	return fmt.Errorf("Not implemented: Kill")
}

func (dbp *Process) requestManualStop() (err error) {
	fmt.Println("requestManualStop")
	return fmt.Errorf("Not implemented: requestManualStop")
}

func (dbp *Process) updateThreadList() error {
	fmt.Println("Did not update the thread list - I think this will not be necessary?")
	
	for threadID := range dbp.Threads {
		if dbp.CurrentThread == nil {
			dbp.SwitchThread(threadID)
		}
		return nil
	}
	return nil
}

func (dbp *Process) addThread(hThread C.HANDLE, threadID int, attach bool) (*Thread, error) {
	if thread, ok := dbp.Threads[threadID]; ok {
		return thread, nil
	}
	thread := &Thread{
		Id:  threadID,
		dbp: dbp,
		os:  new(OSSpecificDetails),
	}
	thread.os.hThread = hThread
	dbp.Threads[threadID] = thread
	
	return thread, nil
}

func (dbp *Process) parseDebugFrame(exe *pe.File, wg *sync.WaitGroup) {
	defer wg.Done()

	if sec := exe.Section(".debug_frame"); sec != nil {
		debugFrame, err := exe.Section(".debug_frame").Data()
		if err != nil {
			fmt.Println("could not get .debug_frame section", err)
			os.Exit(1)
		}
		dbp.frameEntries = frame.Parse(debugFrame)
	} else {
		fmt.Println("could not find .debug_frame section in binary")
		os.Exit(1)
	}
}

// Borrowed from https://golang.org/src/cmd/internal/objfile/pe.go
func findPESymbol(f *pe.File, name string) (*pe.Symbol, error) {
	for _, s := range f.Symbols {
		if s.Name != name {
			continue
		}
		if s.SectionNumber <= 0 {
			return nil, fmt.Errorf("symbol %s: invalid section number %d", name, s.SectionNumber)
		}
		if len(f.Sections) < int(s.SectionNumber) {
			return nil, fmt.Errorf("symbol %s: section number %d is larger than max %d", name, s.SectionNumber, len(f.Sections))
		}
		return s, nil
	}
	return nil, fmt.Errorf("no %s symbol found", name)
}

// Borrowed from https://golang.org/src/cmd/internal/objfile/pe.go
func loadPETable(f *pe.File, sname, ename string) ([]byte, error) {
	ssym, err := findPESymbol(f, sname)
	if err != nil {
		return nil, err
	}
	esym, err := findPESymbol(f, ename)
	if err != nil {
		return nil, err
	}
	if ssym.SectionNumber != esym.SectionNumber {
		return nil, fmt.Errorf("%s and %s symbols must be in the same section", sname, ename)
	}
	sect := f.Sections[ssym.SectionNumber-1]
	data, err := sect.Data()
	if err != nil {
		return nil, err
	}
	return data[ssym.Value:esym.Value], nil
}

// Borrowed from https://golang.org/src/cmd/internal/objfile/pe.go
func pcln(exe *pe.File) (textStart uint64, symtab, pclntab []byte, err error) {
	var imageBase uint64
	switch oh := exe.OptionalHeader.(type) {
	case *pe.OptionalHeader32:
		imageBase = uint64(oh.ImageBase)
	case *pe.OptionalHeader64:
		imageBase = oh.ImageBase
	default:
		return 0, nil, nil, fmt.Errorf("pe file format not recognized")
	}
	if sect := exe.Section(".text"); sect != nil {
		textStart = imageBase + uint64(sect.VirtualAddress)
	}
	if pclntab, err = loadPETable(exe, "runtime.pclntab", "runtime.epclntab"); err != nil {
		// We didn't find the symbols, so look for the names used in 1.3 and earlier.
		// TODO: Remove code looking for the old symbols when we no longer care about 1.3.
		var err2 error
		if pclntab, err2 = loadPETable(exe, "pclntab", "epclntab"); err2 != nil {
			return 0, nil, nil, err
		}
	}
	if symtab, err = loadPETable(exe, "runtime.symtab", "runtime.esymtab"); err != nil {
		// Same as above.
		var err2 error
		if symtab, err2 = loadPETable(exe, "symtab", "esymtab"); err2 != nil {
			return 0, nil, nil, err
		}
	}
	return textStart, symtab, pclntab, nil
}

func (dbp *Process) obtainGoSymbols(exe *pe.File, wg *sync.WaitGroup) {
	defer wg.Done()
	
	_, symdat, pclndat, err := pcln(exe)
	if err != nil {
		fmt.Println("could not get Go symbols", err)
		os.Exit(1)
	}
	
	pcln := gosym.NewLineTable(pclndat, uint64(exe.Section(".text").Offset))
	tab, err := gosym.NewTable(symdat, pcln)
	if err != nil {
		fmt.Println("could not get initialize line table", err)
		os.Exit(1)
	}

	dbp.goSymTable = tab
}

func (dbp *Process) parseDebugLineInfo(exe *pe.File, wg *sync.WaitGroup) {
	defer wg.Done()
		
	if sec := exe.Section(".debug_line"); sec != nil {
		debugLine, err := exe.Section(".debug_line").Data()
		if err != nil {
			fmt.Println("could not get .debug_line section", err)
			os.Exit(1)
		}
		dbp.lineInfo = line.Parse(debugLine)
	} else {
		fmt.Println("could not find .debug_line section in binary")
		os.Exit(1)
	}
}

func (dbp *Process) findExecutable(path string) (*pe.File, error) {
	if path == "" {
		path = fmt.Sprintf("/proc/%d/exe", dbp.Pid)
	}
	f, err := os.OpenFile(path, 0, os.ModePerm)
	if err != nil {
		return nil, err
	}
	elfFile, err := pe.NewFile(f)
	if err != nil {
		return nil, err
	}
	data, err := elfFile.DWARF()
	if err != nil {
		return nil, err
	}
	dbp.dwarf = data
	return elfFile, nil
}

func (dbp *Process) trapWait(pid int) (*Thread, error) {
	var res C.BOOL
	var tid int
	dbp.execPtraceFunc(func() {
		var threadID C.DWORD
		res = C.wait(&threadID)
		tid = int(threadID)
	})
	if res < 0 {
		return nil, fmt.Errorf("Failed to continue debugging")	
	}
	thread, err := dbp.handleBreakpointOnThread(tid)
	if err != nil {
		return nil, err
	}
	return thread, nil
}

func wait(pid, tgid, options int) (int, *sys.WaitStatus, error) {
	fmt.Println("wait")
	return 0, nil, fmt.Errorf("Not implemented: wait")
}

func killProcess(pid int) (err error) {
	fmt.Println("killProcess")
	return fmt.Errorf("Not implemented: killProcess")
}
