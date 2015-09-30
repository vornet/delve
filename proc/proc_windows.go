package proc

// #include "proc_darwin.h"
// #include "exec_darwin.h"
// #include <stdlib.h>
//import "C"
import (
	"debug/macho"
	"fmt"
	"os"
	"sync"
	sys "golang.org/x/sys/windows"
)

// Darwin specific information.
type OSProcessDetails struct {
}

// Create and begin debugging a new process. Uses a
// custom fork/exec process in order to take advantage of
// PT_SIGEXC on Darwin which will turn Unix signals into
// Mach exceptions.
func Launch(cmd []string) (*Process, error) {
	return nil, fmt.Errorf("Not implemented")
}

// Attach to an existing process with the given PID.
func Attach(pid int) (*Process, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (dbp *Process) Kill() (err error) {
	return fmt.Errorf("Not implemented")
}

func (dbp *Process) requestManualStop() (err error) {
	return fmt.Errorf("Not implemented")
}

func (dbp *Process) updateThreadList() error {
	return fmt.Errorf("Not implemented")
}

func (dbp *Process) addThread(port int, attach bool) (*Thread, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (dbp *Process) parseDebugFrame(exe *macho.File, wg *sync.WaitGroup) {
	os.Exit(1)
}

func (dbp *Process) obtainGoSymbols(exe *macho.File, wg *sync.WaitGroup) {
	os.Exit(1)
}

func (dbp *Process) parseDebugLineInfo(exe *macho.File, wg *sync.WaitGroup) {
	os.Exit(1)
}

func (dbp *Process) findExecutable(path string) (*macho.File, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (dbp *Process) trapWait(pid int) (*Thread, error) {
	return nil, fmt.Errorf("Not implemented")
}

func wait(pid, tgid, options int) (int, *sys.WaitStatus, error) {
	return 0, nil, fmt.Errorf("Not implemented")
}

func killProcess(pid int) (err error) {
	return fmt.Errorf("Not implemented")
}

