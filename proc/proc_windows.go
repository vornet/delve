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
	var hThread C.int
	
	res, err := C.waitForCreateProcessEvent(&hProcess, &hThread)
	if res != 0 {
		return nil, err 
	}
	dbp.os.hProcess = hProcess
	_, err = dbp.addThread(int(hThread), false)
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
	fmt.Println("updateThreadList")
	fmt.Println("Did not update the thread list - I think this will not be necessary?")
	return nil
}

func (dbp *Process) addThread(hThread int, attach bool) (*Thread, error) {
	fmt.Println("addThread")
	if thread, ok := dbp.Threads[hThread]; ok {
		return thread, nil
	}
	thread := &Thread{
		Id:  hThread,
		dbp: dbp,
		os:  new(OSSpecificDetails),
	}
	dbp.Threads[hThread] = thread
	//thread.os.thread_act = C.thread_act_t(hThread)
	if dbp.CurrentThread == nil {
		dbp.SwitchThread(thread.Id)
	}
	return thread, nil
}

func (dbp *Process) parseDebugFrame(exe *pe.File, wg *sync.WaitGroup) {
	fmt.Println("parseDebugFrame")
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
	fmt.Println("#success parseDebugFrame")
}

func (dbp *Process) obtainGoSymbols(exe *pe.File, wg *sync.WaitGroup) {
	fmt.Println("obtainGoSymbols")
	defer wg.Done()

	var (
		symdat  []byte
		pclndat []byte
		err     error
	)

	if sec := exe.Section(".gosymtab"); sec != nil {
		symdat, err = sec.Data()
		if err != nil {
			fmt.Println("could not get .gosymtab section", err)
			os.Exit(1)
		}
	}

	if sec := exe.Section(".gopclntab"); sec != nil {
		pclndat, err = sec.Data()
		if err != nil {
			fmt.Println("could not get .gopclntab section", err)
			os.Exit(1)
		}
	}
	
	pcln := gosym.NewLineTable(pclndat, uint64(exe.Section(".text").Offset))
	tab, err := gosym.NewTable(symdat, pcln)
	if err != nil {
		fmt.Println("could not get initialize line table", err)
		os.Exit(1)
	}

	dbp.goSymTable = tab
	fmt.Println("#success obtainGoSymbols")
}

func (dbp *Process) parseDebugLineInfo(exe *pe.File, wg *sync.WaitGroup) {
	fmt.Println("parseDebugLineInfo")
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
	fmt.Println("#success parseDebugLineInfo")
}

func (dbp *Process) findExecutable(path string) (*pe.File, error) {
	fmt.Println("findExecutable")
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
	fmt.Println("trapWait")
	sum := C.add(3,4)
	fmt.Println(sum)
	return nil, fmt.Errorf("Not implemented")
}

func wait(pid, tgid, options int) (int, *sys.WaitStatus, error) {
	fmt.Println("wait")
	return 0, nil, fmt.Errorf("Not implemented")
}

func killProcess(pid int) (err error) {
	fmt.Println("killProcess")
	return fmt.Errorf("Not implemented")
}

