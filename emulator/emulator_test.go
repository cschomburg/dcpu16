package emulator

import (
	"testing"
)

var notchMem  = []uint16{
	0x7c01, 0x0030, 0x7de1, 0x1000, 0x0020, 0x7803, 0x1000, 0xc00d,
	0x7dc1, 0x001a, 0xa861, 0x7c01, 0x2000, 0x2161, 0x2000, 0x8463,
	0x806d, 0x7dc1, 0x000d, 0x9031, 0x7c10, 0x0018, 0x7dc1, 0x001a,
	0x9037, 0x61c1, 0x7dc1, 0x001a, 0x0000, 0x0000, 0x0000, 0x0000,
}

func TestNotch(t *testing.T) {
	dcpu := NewDCPU()

	dcpu.Load(notchMem)

	for dcpu.PC != 0x001a {
		dcpu.Step()
	}

	if dcpu.R[3] != 0x40 {
		t.Errorf("Register X: got 0x%04x, want 0x0040\n", dcpu.R[3])
	}
}

func TestInvalidCode(t *testing.T) {
	dcpu := NewDCPU()
	err := dcpu.Exec()
	if err == nil {
		t.Errorf("No error, but expected UnknownOpError\n")
	} else if e, ok := err.(*UnknownOpError); !ok {
		t.Errorf("Expected UnknownOpError, but got: %s\n", e)
	}
}
