package proc

// #include "threads_darwin.h"
// #include "proc_darwin.h"
import "C"
import (
	"fmt"
	"unsafe"
	sys "golang.org/x/sys/unix"
)

// WaitStatus is a synonym for the platform-specific WaitStatus
type WaitStatus sys.WaitStatus

// OSSpecificDetails holds information specific to the OSX/Darwin
// operating system / kernel.
type OSSpecificDetails struct {
	threadAct C.thread_act_t
	registers C.x86_thread_state64_t
}

// ErrContinueThread is the error returned when a thread could not
// be continued.
var ErrContinueThread = fmt.Errorf("could not continue thread")

func (t *Thread) halt() (err error) {
	kret := C.thread_suspend(t.os.threadAct)
	if kret != C.KERN_SUCCESS {
		errStr := C.GoString(C.mach_error_string(C.mach_error_t(kret)))
		err = fmt.Errorf("could not suspend thread %d %s", t.ID, errStr)
		return
	}
	return
}

func (t *Thread) singleStep() error {
	kret := C.single_step(t.os.threadAct)
	if kret != C.KERN_SUCCESS {
		return fmt.Errorf("could not single step")
	}
	for {
		port := C.mach_port_wait(t.dbp.os.portSet, C.int(0))
		if port == C.mach_port_t(t.ID) {
			break
		}
	}

	kret = C.clear_trap_flag(t.os.threadAct)
	if kret != C.KERN_SUCCESS {
		return fmt.Errorf("could not clear CPU trap flag")
	}
	return nil
}

func (t *Thread) resume() error {
	t.running = true
	// TODO(dp) set flag for ptrace stops
	var err error
	t.dbp.execPtraceFunc(func() { err = PtraceCont(t.dbp.Pid, 0) })
	if err == nil {
		return nil
	}
	kret := C.resume_thread(t.os.threadAct)
	if kret != C.KERN_SUCCESS {
		return ErrContinueThread
	}
	return nil
}

func (t *Thread) blocked() bool {
	// TODO(dp) cache the func pc to remove this lookup
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
	return C.thread_blocked(t.os.threadAct) > C.int(0)
}

func (t *Thread) writeMemory(addr uintptr, data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}
	var (
		vmData = unsafe.Pointer(&data[0])
		vmAddr = C.mach_vm_address_t(addr)
		length = C.mach_msg_type_number_t(len(data))
	)
	if ret := C.write_memory(t.dbp.os.task, vmAddr, vmData, length); ret < 0 {
		return 0, fmt.Errorf("could not write memory")
	}
	return len(data), nil
}

func (t *Thread) readMemory(addr uintptr, size int) ([]byte, error) {
	if size == 0 {
		return nil, nil
	}
	var (
		buf    = make([]byte, size)
		vmData = unsafe.Pointer(&buf[0])
		vmAddr = C.mach_vm_address_t(addr)
		length = C.mach_msg_type_number_t(size)
	)

	ret := C.read_memory(t.dbp.os.task, vmAddr, vmData, length)
	if ret < 0 {
		return nil, fmt.Errorf("could not read memory")
	}
	return buf, nil
}
