package proc

// #include "windows.h"
// #include <stdlib.h>
import "C"
import (
	"debug/pe"
	"debug/gosym"
	"errors"
	"fmt"
	"os"
	"sync"
	sys "golang.org/x/sys/windows"
	"path/filepath"
	"os/exec"
	"syscall"
	"unsafe"
	"runtime"
	
	"github.com/derekparker/delve/dwarf/line"
	"github.com/derekparker/delve/dwarf/frame"
)

const (
	// DEBUGONLYTHISPROCESS tracks https://msdn.microsoft.com/en-us/library/windows/desktop/ms684863(v=vs.85).aspx
	DEBUGONLYTHISPROCESS = 0x00000002
)

// OSProcessDetails holds Windows specific information.
type OSProcessDetails struct {
	hProcess	syscall.Handle
	breakThread int
}

// ErrContinueThread is the error returned when a thread could not
// be continued.
var ErrContinueThread = fmt.Errorf("could not continue thread")

// Launch creates and begins debugging a new process.
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
    
    attr := new(syscall.ProcAttr)
    attr.Files = []uintptr{uintptr(syscall.Stdin), uintptr(syscall.Stdout), uintptr(syscall.Stderr)}
    attr.Sys = new(syscall.SysProcAttr)
    attr.Sys.CreationFlags |= DEBUGONLYTHISPROCESS
    pid, phandle, err := syscall.StartProcess(argv0Go, []string{}, attr)
	if err != nil {
		return nil, err
	}
	defer sys.CloseHandle(sys.Handle(phandle))
	
	dbp := New(pid)

	switch runtime.GOARCH {
	case "amd64":
		dbp.arch = AMD64Arch()
	}
	
	dbp.execPtraceFunc(func() {
		// TODO - We're ignoring the results because we assume we'll immediately hit
		// the default breakpoint that Windows sets at process creation.
		// Should perhaps be testing that we're not overlooking an exit event or similar?
		// Should return ProcessExitedError if the process exits (if that is even possible?).
		_, _, err = dbp.waitForDebugEvent()
	})
	
	return initializeDebugProcess(dbp, argv0Go, false)
}

// Attach to an existing process with the given PID.
func Attach(pid int) (*Process, error) {
	fmt.Println("Attach")
	return nil, fmt.Errorf("Not implemented: Attach")
}

// Kill kills the process.
func (dbp *Process) Kill() (err error) {
	if dbp.exited {
		return nil
	}
	if !dbp.Threads[dbp.Pid].Stopped() {
		return errors.New("process must be stopped in order to kill it")
	}
	// TODO: Should not have to ignore failures here,
	// but some tests appear to Kill twice causing 
	// this to fail on second attempt.
	_ = C.TerminateProcess(C.HANDLE(unsafe.Pointer(dbp.os.hProcess)), 1)
	dbp.exited = true
	return
}

func (dbp *Process) requestManualStop() (err error) {
	res := C.DebugBreakProcess(C.HANDLE(unsafe.Pointer(dbp.os.hProcess)))
	if res == 0 {
		return fmt.Errorf("Failed to break process %d", dbp.Pid)	
	}
	return nil
}

func (dbp *Process) updateThreadList() error {
	// TODO: Currently we are ignoring this request since we assume that
	// threads are being tracked as they are created/killed. 
	
	return nil
}

func (dbp *Process) addThread(hThread syscall.Handle, threadID int, attach bool) (*Thread, error) {
	if thread, ok := dbp.Threads[threadID]; ok {
		return thread, nil
	}
	thread := &Thread{
		ID:  threadID,
		dbp: dbp,
		os:  new(OSSpecificDetails),
	}
	thread.os.hThread = hThread
	dbp.Threads[threadID] = thread
	if dbp.CurrentThread == nil {
		dbp.SwitchThread(thread.ID)
	}
	return thread, nil
}

func (dbp *Process) parseDebugFrame(exe *pe.File, wg *sync.WaitGroup) {
	defer wg.Done()

	if sec := exe.Section(".debug_frame"); sec != nil {
		debugFrame, err := sec.Data()
		if err != nil && uint32(len(debugFrame)) < sec.Size {
			fmt.Println("could not get .debug_frame section", err)
			os.Exit(1)
		}
		if 0 < sec.VirtualSize && sec.VirtualSize < sec.Size {
			debugFrame = debugFrame[:sec.VirtualSize]
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
		debugLine, err := sec.Data()
		if err != nil && uint32(len(debugLine)) < sec.Size {
			fmt.Println("could not get .debug_line section", err)
			os.Exit(1)
		}
		if 0 < sec.VirtualSize && sec.VirtualSize < sec.Size {
			debugLine = debugLine[:sec.VirtualSize]
		}
		dbp.lineInfo = line.Parse(debugLine)
	} else {
		fmt.Println("could not find .debug_line section in binary")
		os.Exit(1)
	}
}

func (dbp *Process) findExecutable(path string) (*pe.File, error) {
	if path == "" {
		// TODO: Find executable path from PID/handle on Windows:
		// https://msdn.microsoft.com/en-us/library/aa366789(VS.85).aspx
		return nil, fmt.Errorf("Not yet implemented")
	}
	f, err := os.OpenFile(path, 0, os.ModePerm)
	if err != nil {
		return nil, err
	}
	peFile, err := pe.NewFile(f)
	if err != nil {
		return nil, err
	}
	data, err := peFile.DWARF()
	if err != nil {
		return nil, err
	}
	dbp.dwarf = data
	return peFile, nil
}

func (dbp *Process) waitForDebugEvent() (threadID, exitCode int, err error) {
	var debugEvent C.DEBUG_EVENT
	for {
		if C.WaitForDebugEvent(&debugEvent, C.INFINITE) == C.WINBOOL(0) { 
			return 0, 0, fmt.Errorf("Could not WaitForDebugEvent")
		}
		unionPtr := unsafe.Pointer(&debugEvent.u[0])
		switch debugEvent.dwDebugEventCode {
		case C.CREATE_PROCESS_DEBUG_EVENT:
			debugInfo := (*C.CREATE_PROCESS_DEBUG_INFO)(unionPtr)
			dbp.os.hProcess = syscall.Handle(unsafe.Pointer(debugInfo.hProcess))
			_, err = dbp.addThread(syscall.Handle(unsafe.Pointer(debugInfo.hThread)), int(debugEvent.dwThreadId), false)
			if err != nil {
				return 0, 0, err
			}
			
			C.ContinueDebugEvent(debugEvent.dwProcessId, debugEvent.dwThreadId, C.DBG_CONTINUE)
			continue
		case C.CREATE_THREAD_DEBUG_EVENT:
			debugInfo := (*C.CREATE_THREAD_DEBUG_INFO)(unionPtr)

			_, err = dbp.addThread(syscall.Handle(unsafe.Pointer(debugInfo.hThread)), int(debugEvent.dwThreadId), false)
			if err != nil {
				return 0, 0, err
			}

			C.ContinueDebugEvent(debugEvent.dwProcessId, debugEvent.dwThreadId, C.DBG_CONTINUE)
			continue
		case C.LOAD_DLL_DEBUG_EVENT, C.UNLOAD_DLL_DEBUG_EVENT, C.EXIT_THREAD_DEBUG_EVENT, C.OUTPUT_DEBUG_STRING_EVENT, C.RIP_EVENT:
			C.ContinueDebugEvent(debugEvent.dwProcessId, debugEvent.dwThreadId, C.DBG_CONTINUE)
			continue
		case C.EXCEPTION_DEBUG_EVENT:
			dbp.os.breakThread = int(debugEvent.dwThreadId)
			return int(debugEvent.dwThreadId), 0, nil
		case C.EXIT_PROCESS_DEBUG_EVENT:
			debugInfo := (*C.EXIT_PROCESS_DEBUG_INFO)(unionPtr)
			return 0, int(debugInfo.dwExitCode), nil
		default:
			return 0, 0, fmt.Errorf("Unknown event code: %d", debugEvent.dwDebugEventCode)
		}
	}
}

func (dbp *Process) trapWait(pid int) (*Thread, error) {	
	for {
		var err error
		var tid int
		var exitCode int
		dbp.execPtraceFunc(func() {
			tid, exitCode, err = dbp.waitForDebugEvent()
		})
		if err != nil {
			return nil, err	
		}
		if tid == 0 {
			dbp.postExit()
			return nil, ProcessExitedError{Pid: dbp.Pid, Status: exitCode}
		}
		
		th, ok := dbp.Threads[tid]
        if !ok {
			if dbp.halt {
				dbp.halt = false
				return th, nil
			}
			if dbp.firstStart || th.singleStepping {
				dbp.firstStart = false
				return th, nil
			}
			if err := th.Continue(); err != nil {
				return nil, err
			}
			continue
		}
		return th, nil
	}
}

func (dbp *Process) loadProcessInformation(wg *sync.WaitGroup) {
	wg.Done()
}

func (dbp *Process) wait(pid, options int) (int, *sys.WaitStatus, error) {
	fmt.Println("wait")
	return 0, nil, fmt.Errorf("Not implemented: wait")
}

func (dbp *Process) setCurrentBreakpoints(trapthread *Thread) error {
	for _, th := range dbp.Threads {
		if th.CurrentBreakpoint == nil {
			err := th.SetCurrentBreakpoint()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (dbp *Process) exitGuard(err error) error {
	if err != ErrContinueThread {
		return err
	}
	return ProcessExitedError{Pid: dbp.Pid, Status: 0}
}

func (dbp *Process) resume() error {
	// all threads stopped over a breakpoint are made to step over it
	for _, thread := range dbp.Threads {
		if thread.CurrentBreakpoint != nil {
			if err := thread.Step(); err != nil {
				return err
			}
			thread.CurrentBreakpoint = nil
		}
	}
	// everything is resumed
	var res C.WINBOOL
	dbp.execPtraceFunc(func() {	
		res = C.ContinueDebugEvent(C.DWORD(dbp.Pid), C.DWORD(dbp.os.breakThread), C.DBG_CONTINUE)
	})
	if res == 0 {
		return fmt.Errorf("Could not ContinueDebugEvent.")	
	}
	return nil
}

func killProcess(pid int) error {
	fmt.Println("killProcess")
	return fmt.Errorf("Not implemented: killProcess")
}
