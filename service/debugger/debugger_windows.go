package debugger

import (
	"fmt"
)

func attachErrorMessage(pid int, err error) error {
	return fmt.Errorf("could not attach to pid %d: %s", pid, err)
}

func stopProcess(pid int) error {
	// TODO: We're assuming that process is always stopped 
	// when we are in interactive debugger.
	return nil
}
