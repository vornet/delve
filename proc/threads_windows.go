package proc

// #include "threads_darwin.h"
//import "C"
import (
	"fmt"
	sys "golang.org/x/sys/windows"
)

// Thread represents a single thread in the traced process
// Id represents the thread id or port, Process holds a reference to the
// Process struct that contains info on the process as
// a whole, and Status represents the last result of a `wait` call
// on this thread.
type Thread struct {
	Id                int             // Thread ID or mach port
	Status            *sys.WaitStatus // Status returned from last wait call
	CurrentBreakpoint *Breakpoint     // Breakpoint thread is currently stopped at

	dbp            *Process
	singleStepping bool
	running        bool
	os             *OSSpecificDetails
}


type OSSpecificDetails struct {
}

func (t *Thread) halt() (err error) {
	return fmt.Errorf("Not implemented")
}

func (t *Thread) singleStep() error {
	return fmt.Errorf("Not implemented")
}

func (t *Thread) resume() error {
	return fmt.Errorf("Not implemented")
}

func (thread *Thread) blocked() bool {
	return false
}

func (thread *Thread) stopped() bool {
	return false
}

func (thread *Thread) writeMemory(addr uintptr, data []byte) (int, error) {
	return 0, fmt.Errorf("Not implemented")
}


func (thread *Thread) readMemory(addr uintptr, size int) ([]byte, error) {
	return nil, fmt.Errorf("Not implemented")
}
