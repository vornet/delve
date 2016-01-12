package proc

// #include "threads_windows.h"
import "C"
import (
	"fmt"
	"unsafe"
	"syscall"
	sys "golang.org/x/sys/windows"
)

// WaitStatus is a synonym for the platform-specific WaitStatus
type WaitStatus sys.WaitStatus

// OSSpecificDetails holds information specific to the Windows
// operating system / kernel.
type OSSpecificDetails struct {
	hThread syscall.Handle
}

func (t *Thread) halt() (err error) {
	// TODO: We are currently ignoring the request to halt.
	// I believe this is guaranteed to already be halted. 
	return nil
}

func (t *Thread) singleStep() error {	
	var context C.CONTEXT
	context.ContextFlags = C.CONTEXT_ALL;
	
	res, err := C.GetThreadContext(C.HANDLE(unsafe.Pointer(t.os.hThread)), &context)
	if res == 0 {
		return err
	}
	
	//TODO: Is it really okay to not decrement to IP?
	//context.Rip--;
	context.EFlags |= 0x100;
	
	res, err = C.SetThreadContext(C.HANDLE(unsafe.Pointer(t.os.hThread)), &context)
	if res == 0 {
		return err
	}
	
	// TODO: It's possible that this resume/trapWait will 
	// land us on another thread. We need to guard for that 
	// and loop here till we sucessfully land back on our thread.
	err = t.resume()
	if err != nil {
		return err
	}
	_, err = t.dbp.trapWait(0)
	if err != nil {
		return err
	}
	
	res, err = C.GetThreadContext(C.HANDLE(unsafe.Pointer(t.os.hThread)), &context)
	if res == 0 {
		return err
	}
		
	context.EFlags ^= 0x100;
	
	res, err = C.SetThreadContext(C.HANDLE(unsafe.Pointer(t.os.hThread)), &context)
	if res == 0 {
		return err
	}
	
	return nil
}

func (t *Thread) resume() error {
	t.running = true
	var res C.WINBOOL
	t.dbp.execPtraceFunc(func() {
		//TODO: Note that we are ignoring the thread we were asked to continue and are continuing the 
		//thread that we last broke on.	
		res = C.ContinueDebugEvent(C.DWORD(t.dbp.Pid), C.DWORD(t.dbp.os.breakThread), C.DBG_CONTINUE)
	})
	if res == 0 {
		return fmt.Errorf("Could not ContinueDebugEvent.")	
	}
	return nil
}

func (t *Thread) blocked() bool {
	// TODO: Probably incorrect - what are teh runtime functions that
	// indicate blocking on Windows?
	pc, err := t.PC()
	if err != nil {
		return false
	}
	fn := t.dbp.goSymTable.PCToFunc(pc)
	if fn == nil {
		return false
	}
	switch fn.Name {
	case "runtime.kevent", "runtime.mach_semaphore_wait", "runtime.usleep":
		return true
	default:
		return false
	}
}

func (t *Thread) stopped() bool {
	// TODO: We are assuming that threads are always stopped
	// during command exection.
	return true 
}

func (t *Thread) canContinue() bool {
	return t.dbp.os.breakThread == t.ID
}

func (t *Thread) writeMemory(addr uintptr, data []byte) (int, error) {
	var (
		vmData = unsafe.Pointer(&data[0])
		vmAddr = unsafe.Pointer(addr)
		length  = C.int(len(data))
	)
	ret := C.write_memory(C.HANDLE(unsafe.Pointer(t.dbp.os.hProcess)), vmAddr, vmData, length)
	if ret < 0 {
		return int(ret), fmt.Errorf("could not write memory")
	}
	return int(ret), nil
}

func (t *Thread) readMemory(addr uintptr, size int) ([]byte, error) {
	if size == 0 {
		return nil, nil
	}
	var (
		buf     = make([]byte, size)
		vmData = unsafe.Pointer(&buf[0])
		vmAddr = unsafe.Pointer(addr)
		length  = C.int(size)
	)

	ret := C.read_memory(C.HANDLE(unsafe.Pointer(t.dbp.os.hProcess)), vmAddr, vmData, length)
	if ret < 0 {
		return nil, fmt.Errorf("could not read memory")
	}
	return buf, nil
}
