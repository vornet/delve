package proc

import (
	"fmt"
)

func PtraceAttach(pid int) error {
	return fmt.Errorf("Not implemented: PtraceAttach")
}

func PtraceDetach(tid, sig int) error {
	return fmt.Errorf("Not implemented: PtraceDetach")
}

func PtraceCont(tid, sig int) error {
	return fmt.Errorf("Not implemented: PtraceCont")
}

func PtraceSingleStep(tid int) error {
	return fmt.Errorf("Not implemented: PtraceSingleStep")
}
