package debugger

import (
	"github.com/xconstruct/dcpu16/disassembler"
	"github.com/xconstruct/dcpu16/emulator"
	"fmt"
)

// RDump outputs the current state of the registers.
func RDump(d *emulator.DCPU) {
	for i, word := range(d.R) {
		fmt.Printf(" %s: %#04x ", disassembler.Registers[i], word)
	}
	fmt.Println()
}

// PrintInstruction prints the instruction at PC.
func PrintInstruction(d *emulator.DCPU) {
	str, _ := disassembler.InstructionString(d.RAM[d.PC:])
	fmt.Println(str)
}

// StepLoop executes the program until an single instruction loop is encountered.
func StepLoop(d *emulator.DCPU) error {
	lastPC := d.PC
	for {
		err := d.Step()
		if d.PC == lastPC || err != nil {
			return err
		}
		lastPC = d.PC
	}
	return nil
}

// StepJmp executes the program until a "SET PC, ..." is encountered.
func StepJmp(d *emulator.DCPU) error {
	for {
		word := d.RAM[d.PC]
		if (word & 0x3ff) == 0x1c1 {
			return nil
		}
		err := d.Step()
		if err != nil {
			return err
		}
	}
	return nil
}
