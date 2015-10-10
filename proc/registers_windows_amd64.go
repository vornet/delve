package proc

// #include "threads_darwin.h"
//import "C"
import (
	"fmt"
)

type Regs struct {
	rax     uint64
	rbx     uint64
	rcx     uint64
	rdx     uint64
	rdi     uint64
	rsi     uint64
	rbp     uint64
	rsp     uint64
	r8      uint64
	r9      uint64
	r10     uint64
	r11     uint64
	r12     uint64
	r13     uint64
	r14     uint64
	r15     uint64
	rip     uint64
	rflags  uint64
	cs      uint64
	fs      uint64
	gs      uint64
	gs_base uint64
}

func (r *Regs) String() string {
	return ""
}

func (r *Regs) PC() uint64 {
	return 0
}

func (r *Regs) SP() uint64 {
	return 0
}

func (r *Regs) CX() uint64 {
	return 0
}

func (r *Regs) TLS() uint64 {
	return r.gs_base
}

func (r *Regs) SetPC(thread *Thread, pc uint64) error {
	return fmt.Errorf("Not implemented: SetPC")
}

func registers(thread *Thread) (Registers, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (thread *Thread) saveRegisters() (Registers, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (thread *Thread) restoreRegisters() error {
	return fmt.Errorf("Not implemented")
}
