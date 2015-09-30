package debugger

import (
	"fmt"
)

func attachErrorMessage(pid int, err error) error {
	//TODO: mention certificates?
	return fmt.Errorf("could not attach to pid %d: %s", pid, err)
}

func killProcess(pid int) error {
	return fmt.Errorf("Not implemented")
}
