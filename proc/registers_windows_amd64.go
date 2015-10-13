package proc

// #include "threads_windows.h"
import "C"
import (
	"fmt"
	"bytes"
	"encoding/binary"
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
	eflags  uint64
	cs      uint64
	fs      uint64
	gs      uint64
	tls		uint64
}

func (r *Regs) String() string {
	fmt.Println("registers.String")
	var buf bytes.Buffer
	var regs = []struct {
		k string
		v uint64
	}{
		{"Rip", r.rip},
		{"Rsp", r.rsp},
		{"Rax", r.rax},
		{"Rbx", r.rbx},
		{"Rcx", r.rcx},
		{"Rdx", r.rdx},
		{"Rdi", r.rdi},
		{"Rsi", r.rsi},
		{"Rbp", r.rbp},
		{"R8", r.r8},
		{"R9", r.r9},
		{"R10", r.r10},
		{"R11", r.r11},
		{"R12", r.r12},
		{"R13", r.r13},
		{"R14", r.r14},
		{"R15", r.r15},
		{"Eflags", r.eflags},
		{"Cs", r.cs},
		{"Fs", r.fs},
		{"Gs", r.gs},
		{"TLS", r.tls},
	}
	for _, reg := range regs {
		fmt.Fprintf(&buf, "%8s = %0#16x\n", reg.k, reg.v)
	}
	return buf.String()
}

func (r *Regs) PC() uint64 {
	fmt.Println("registers.PC")
	return r.rip
}

func (r *Regs) SP() uint64 {
	fmt.Println("registers.SP")
	return r.rsp
}

func (r *Regs) CX() uint64 {
	fmt.Println("registers.CX")
	return r.rcx
}

func (r *Regs) TLS() uint64 {
	fmt.Println("registers.TLS")
	return r.tls
}

func (r *Regs) SetPC(thread *Thread, pc uint64) error {
	fmt.Println("registers.SetPC")
	return fmt.Errorf("Not implemented: SetPC")
}

func registers(thread *Thread) (Registers, error) {
	fmt.Println("registers.registers")
	var context C.CONTEXT
	
	context.ContextFlags = C.CONTEXT_ALL;
	res, err := C.GetThreadContext(thread.os.hThread, &context)
	if res == 0 {
		return nil, err
	}
	
	var threadInfo C.THREAD_BASIC_INFORMATION
	status := C.thread_basic_information(thread.os.hThread, &threadInfo)
	if status != 0 {
		return nil, fmt.Errorf("Failed to get thread_basic_information")
	}
	bytes,err := thread.readMemory(uintptr(threadInfo.TebBaseAddress), 0x60)
	tls := binary.LittleEndian.Uint64(bytes[0x58:0x60])
	
	regs := &Regs{
		rax:     uint64(context.Rax),
		rbx:     uint64(context.Rbx),
		rcx:     uint64(context.Rcx),
		rdx:     uint64(context.Rdx),
		rdi:     uint64(context.Rdi),
		rsi:     uint64(context.Rsi),
		rbp:     uint64(context.Rbp),
		rsp:     uint64(context.Rsp),
		r8:      uint64(context.R8),
		r9:      uint64(context.R9),
		r10:     uint64(context.R10),
		r11:     uint64(context.R11),
		r12:     uint64(context.R12),
		r13:     uint64(context.R13),
		r14:     uint64(context.R14),
		r15:     uint64(context.R15),
		rip:     uint64(context.Rip),
		eflags:  uint64(context.EFlags),
		cs:      uint64(context.SegCs),
		fs:      uint64(context.SegFs),
		gs:      uint64(context.SegGs),
		tls:     tls,
	}
	return regs, nil
}

func (thread *Thread) saveRegisters() (Registers, error) {
	fmt.Println("registers.saveRegisters")
	return nil, fmt.Errorf("Not implemented")
}

func (thread *Thread) restoreRegisters() error {
	fmt.Println("registers.restoreRegisters")
	return fmt.Errorf("Not implemented")
}
