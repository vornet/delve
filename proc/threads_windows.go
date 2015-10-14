package proc

// #include "threads_windows.h"
import "C"
import (
	"fmt"
	"unsafe"
	sys "golang.org/x/sys/windows"
)

type WaitStatus sys.WaitStatus

type OSSpecificDetails struct {
	hThread C.HANDLE
}

func (t *Thread) halt() (err error) {
	fmt.Println("Ignoring the request to halt - I believe this is guaranteed to already be halted...")
	return nil
}

func (t *Thread) singleStep() error {
	fmt.Println("singleStep")
	return fmt.Errorf("Not implemented: singleStep")
}

func (t *Thread) resume() error {
	var res C.BOOL
	t.dbp.execPtraceFunc(func() {
		res = C.continue_debugger(C.DWORD(t.dbp.Pid), C.DWORD(t.Id))
	})
	if res == 0 {
		errCode := int(C.GetLastError())
		return fmt.Errorf("Could not continue: %d", errCode)	
	}
	return nil
}

func (thread *Thread) blocked() bool {
	fmt.Println("blocked")
	return false
}

func (thread *Thread) stopped() bool {
	fmt.Println("stopped")
	return false
}

func (thread *Thread) writeMemory(addr uintptr, data []byte) (int, error) {
	var (
		vm_data = unsafe.Pointer(&data[0])
		vm_addr = unsafe.Pointer(addr)
		length  = C.int(len(data))
	)
	ret := C.write_memory(thread.dbp.os.hProcess, vm_addr, vm_data, length)
	if ret < 0 {
		return int(ret), fmt.Errorf("could not write memory")
	}
	return int(ret), nil
}

func (thread *Thread) readMemory(addr uintptr, size int) ([]byte, error) {
	if size == 0 {
		return nil, nil
	}
	var (
		buf     = make([]byte, size)
		vm_data = unsafe.Pointer(&buf[0])
		vm_addr = unsafe.Pointer(addr)
		length  = C.int(size)
	)

	ret := C.read_memory(thread.dbp.os.hProcess, vm_addr, vm_data, length)
	if ret < 0 {
		return nil, fmt.Errorf("could not read memory")
	}
	return buf, nil
}
