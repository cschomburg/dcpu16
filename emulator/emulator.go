// Package dcpu/emulator provides a simple emulator type for the DCPU-16
// which can load code into memory and execute it.
//
// Here is a simple example:
//
// 		dcpu := emulator.NewDCPU()
// 		dcpu.Load(myWordArray)
// 		err := dcpu.Exec()
package emulator

import (
	"fmt"
)

// UnknownOpError records an error when a missing opcode is encountered.
type UnknownOpError struct {
	PC uint16
	Op byte
}

func (e *UnknownOpError) Error() string {
	return fmt.Sprintf("dcpu/emulator: Unknown op 0x%x at 0x%04x", e.Op, e.PC)
}

// DCPU is an emulator for the DCPU-16.
type DCPU struct {
	RAM []uint16
	R []uint16
	PC uint16
	SP uint16
	O uint16
	offset int
}

// NewDCPU creates a new DCPU instance.
func NewDCPU() (*DCPU) {
	return &DCPU{
		make([]uint16, 0x10000),
		make([]uint16, 8),
		0, 0, 0, 0,
	}
}

// Reset completely resets the program state except for the RAM.
func (d *DCPU) Reset() {
	d.R = make([]uint16, 8)
	d.PC = 0
	d.SP = 0
	d.O = 0
}

// Load copies the mem word-array into the RAM
func (d *DCPU) Load(mem []uint16) {
	copy(d.RAM, mem)
}

// Exec runs the program saved in RAM until an error is encountered.
func (d *DCPU) Exec() error {
	for 0x0000 <= d.PC && d.PC <= 0xffff {
		err := d.Step()
		if err != nil {
			return err
		}
	}
	return nil
}

// Step executes the next instruction in RAM.
func (d *DCPU) Step() error {
	word := d.nextWord()
	level, op, args := GetOp(word)

	if level == 0 { // basic opcodes
		aV, aP:= d.readValue(args[0])
		bV, _ := d.readValue(args[1])

		if aP == nil && op <= 0xc { // fail silently for setting literal a
			return nil
		}

		switch op {
		case 0x1: *aP = bV // SET
		case 0x2: *aP = aV + bV // ADD
		case 0x3: *aP = aV - bV // SUB
		case 0x4: *aP = aV * bV; d.O = uint16(((uint(aV)*uint(bV))>>16) & 0xffff)// MUL
		case 0x5: // DIV
			if bV == 0 {
				*aP = 0; d.O = 0
			} else {
				*aP = aV / bV; d.O = uint16(((uint(aV<<16)/uint(bV))) & 0xffff)
			}
		case 0x6: // MOD
			if bV == 0 {
				*aP = 0
			} else {
				*aP = aV % bV;
			}
		case 0x7: *aP = aV << bV; d.O = uint16(((uint(aV)<<bV)>>16)&0xffff)
		case 0x8: *aP = aV >> bV; d.O = uint16(((uint(aV)<<16)>>bV)&0xffff)
		case 0x9: *aP = aV & bV; // AND
		case 0xa: *aP = aV | bV; // BOR
		case 0xb: *aP = aV ^ bV; // XOR
		case 0xc: if aV != bV { d.stepIgnore() } // IFE
		case 0xd: if aV == bV { d.stepIgnore() } // IFN
		case 0xe: if aV <= bV { d.stepIgnore() } // IFG
		case 0xf: if (aV & bV) == 0 { d.stepIgnore() } // IFB
		}

		return nil
	}

	// non-basic opcodes
	if op == 0x01 { // JSR
		aV, _ := d.readValue(args[0])
		d.SP--
		d.RAM[d.SP] = d.PC
		d.PC = aV
		return nil
	}

	if op == 0x02 || op == 0x03 { // reserved
		return nil
	}

	return &UnknownOpError{d.PC, op}
}

// stepIgnore steps over the next instruction without executing it.
func (d *DCPU) stepIgnore() {
	_, _, args := GetOp(d.nextWord())
	for _, v := range(args) {
		d.readValueIgnore(v)
	}
}

// readValue parses a value code and returns the referenced value and,
// if applicable, a pointer to write to this location. May modify PC / SP.
func (d *DCPU) readValue(v byte) (word uint16, ptr *uint16) {
	switch {
	case v <= 0x07: ptr = &d.R[v] // register
	case v <= 0x0f: ptr = &d.RAM[d.R[v-0x08]] // [register]
	case v <= 0x17: ptr = &d.RAM[d.nextWord() + d.R[v-0x10]] // [next word + register]
	case v == 0x18: ptr = &d.RAM[d.SP]; d.SP++; // POP [SP++]
	case v == 0x19: ptr = &d.RAM[d.SP] // PEEK [SP]
	case v == 0x1a: d.SP--; ptr = &d.RAM[d.SP] // PUSH [--SP]
	case v == 0x1c: ptr = &d.PC // PC
	case v == 0x1d: ptr = &d.O // O
	case v == 0x1e: ptr = &d.RAM[d.nextWord()] // [next word]
	case v == 0x1f: word = d.nextWord() // next word (literal)
	default:        word = uint16(v-0x20) // literal value 0x00-0x1f (literal)
	}
	if ptr != nil {
		word = *ptr
	}
	return word, ptr
}

// readValueIgnore parses a value code and fetches the next word if needed
// without modifying SP.
func (d *DCPU) readValueIgnore(v byte) {
	switch {
	case v <= 0x0f: return
	case v <= 0x17: d.nextWord() // [next word + register]
	case v == 0x1e: d.nextWord() // [next word]
	case v == 0x1f: d.nextWord() // next word (literal)
	}
}

// nextWord returns the current word in memory and increments the program counter.
func (d *DCPU) nextWord() uint16 {
	word := d.RAM[d.PC]
	d.PC++
	return word
}

// GetOP splits a word into opcode and an array of arguments.
// Basic returns the level of the op (0 = basic, 1 = non-basic).
func GetOp(word uint16) (level int, op byte, args []byte) {
	level = 0
	op = byte(word & 0xf)
	word >>= 4
	if op == 0x0 {
		op = byte(word & 0x3f)
		word >>= 6
		level++
	}
	args = make([]byte, 2-level)
	for i := 0; i < (2-level); i++ {
		args[i] = byte(word & 0x3f)
		word >>= 6
	}
	return level, op, args
}
