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
	fmt.Println("halt")
	return fmt.Errorf("Not implemented: halt")
}

func (t *Thread) singleStep() error {
	fmt.Println("singleStep")
	return fmt.Errorf("Not implemented: singleStep")
}

func (t *Thread) resume() error {
	fmt.Println("resume")
	return fmt.Errorf("Not implemented: resume")
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
	fmt.Println("writeMemory")
	return 0, fmt.Errorf("Not implemented: writeMemory")
}

func (thread *Thread) readMemory(addr uintptr, size int) ([]byte, error) {
	fmt.Println("readMemory")
	if size == 0 {
		return nil, nil
	}
	var (
		buf     = make([]byte, size)
		vm_data = unsafe.Pointer(&buf[0])
		vm_addr = unsafe.Pointer(addr)
		length  = C.int(size)
	)

	fmt.Println(thread.dbp.os.hProcess)
	fmt.Println(addr)
	fmt.Println(size)
	fmt.Println(length)
	ret := C.read_memory(thread.dbp.os.hProcess, vm_addr, vm_data, length)
	fmt.Println(buf)
	if ret < 0 {
		return nil, fmt.Errorf("could not read memory")
	}
	return buf, nil
}
