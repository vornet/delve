package proc

import (
	"fmt"
)

func PtraceAttach(pid int) error {
	return fmt.Errorf("Not implemented")
}

func PtraceDetach(tid, sig int) error {
	return fmt.Errorf("Not implemented")
}

func PtraceCont(tid, sig int) error {
	return fmt.Errorf("Not implemented")
}

func PtraceSingleStep(tid int) error {
	return fmt.Errorf("Not implemented")
}
