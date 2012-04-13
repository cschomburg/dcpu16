package disassembler

import (
	"dcpu16/emulator"
	"fmt"
)

var BasicOp = []string{
	"UNKNOWN",
	"SET",
	"ADD",
	"SUB",
	"MUL",
	"DIV",
	"MOD",
	"SHL",
	"SHR",
	"AND",
	"BOR",
	"XOR",
	"IFE",
	"IFN",
	"IFG",
	"IFB",
}

var Registers = []string{"A", "B", "C", "X", "Y", "Z", "I", "J"}

var NonBasicOp = []string{
	"UNKNOWN",
	"JSR",
}

func OpString(level int, op byte) string {
	var lookup []string
	switch level {
	case 0: lookup = BasicOp
	case 1: lookup = NonBasicOp
	default: return "UNKNOWN"
	}
	return lookup[op]
}

func InstructionString(mem []uint16) (str string, wordsRead int) {
	level, op, args := emulator.GetOp(mem[0])
	str = OpString(level, op)

	wordsRead = 1
	for i, v := range(args) {
		if i > 0 {
			str += ","
		}
		vStr, vWordsRead := ValueString(v, mem[wordsRead:])
		str += " " + vStr
		wordsRead += vWordsRead
	}

	return str, wordsRead
}

func ValueString(v byte, mem []uint16) (str string, wordsRead int) {
	switch {
	case v <= 0x07: return Registers[v], 0// register
	case v <= 0x0f: return "["+Registers[v-0x08]+"]", 0 // [register]
	case v <= 0x17:  // [next word + register]
		return fmt.Sprintf("[%#04x+%s]", mem[1], Registers[v-0x10]), 1
	case v == 0x18: return "POP", 0 // POP [SP++]
	case v == 0x19: return "PEEK", 0 // PEEK [SP]
	case v == 0x1a: return "PUSH", 0 // PUSH [--SP]
	case v == 0x1c: return "PC", 0 // PC
	case v == 0x1d: return "O", 0 // O
	case v == 0x1e: return fmt.Sprintf("[%#04x]", mem[0]), 1 // [next word]
	case v == 0x1f: return fmt.Sprintf("%#04x", mem[0]), 1 // next word (literal)
	}
	return fmt.Sprintf("%#02x", v-0x20), 0 // literal value 0x00-0x1f (literal)
}

func Disassemble(mem []uint16) string {
	str := ""
	offset := 0
	for offset <= len(mem) {
		iStr, wordsRead := InstructionString(mem[offset:offset+3])
		str += iStr + "\n"
		offset += wordsRead
	}
	return str
}
