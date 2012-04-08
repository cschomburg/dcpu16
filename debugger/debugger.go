package debugger

import (
	"dcpu/disassembler"
	"dcpu/emulator"
	"fmt"
)

// Memdump outputs the current state of the RAM.
// If part is set to true, only non-null rows are printed.
func Memdump(d *emulator.DCPU, part bool) {
	for l := 0; l < (len(d.RAM) / 8); l++ {
		if part {
			lineNull := true
			for c := 0; c < 8; c++ {
				if d.RAM[l*8+c] != 0 {
					lineNull = false
					break
				}
			}
			if lineNull {
				continue
			}
		}

		fmt.Printf("0x%04x:    ", l*8)
		for c := 0; c < 8; c++ {
			fmt.Printf("0x%04x ", d.RAM[l*8+c])
		}
		print("\n")
	}
}

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
