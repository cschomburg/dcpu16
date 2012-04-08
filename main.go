package main

import (
	"dcpu/debugger"
	"dcpu/emulator"
	"flag"
	"fmt"
	"os"
	"io"
	"bufio"
)

func assert(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {

	tool := ""
	if len(os.Args) > 1 {
		tool = os.Args[1]
	}
	switch tool {
	case "e": fallthrough
	case "emulate":
		runEmulator()
	case "d": fallthrough
	case "debug":
		runDebugger()
	case "dis": fallthrough
	case "disassemble":
		runDisassembler()
	default:
		printHelp("");
	}
}

func runEmulator() {
	flag.Parse()
	path := flag.Arg(1)
	if path == "" {
		fmt.Println("Usage: dcpu debug [ramfile]")
		return
	}

	file, err := os.Open(path)
	assert(err)
	dcpu := emulator.NewDCPU()
	_, err = io.Copy(dcpu, file)
	file.Close()
	assert(err)
	err = dcpu.Exec()
	assert(err)
}

func runDebugger() {
	flag.Parse()
	path := flag.Arg(1)
	if path == "" {
		fmt.Println("Usage: dcpu debug [ramfile]")
		return
	}

	file, err := os.Open(path)
	assert(err)
	dcpu := emulator.NewDCPU()
	_, err = io.Copy(dcpu, file)
	file.Close()
	assert(err)

	in := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("(d) ")
		buf, _, err := in.ReadLine()
		line := string(buf)
		if err == io.EOF {
			return
		}
		assert(err)

		switch (line) {
		case "quit": return
		case "step":
			err := dcpu.Step()
			if err != nil {
				fmt.Println("dcpu err: ", err)
			}
		case "steploop":
			err := debugger.StepLoop(dcpu)
			if err != nil {
				fmt.Println("dcpu err: ", err)
			}
		case "stepjmp":
			err := debugger.StepJmp(dcpu)
			if err != nil {
				fmt.Println("dcpu err: ", err)
			}
		case "mem": debugger.Memdump(dcpu, true)
		case "r":   debugger.RDump(dcpu)
		case "op":  debugger.PrintInstruction(dcpu)
		}
	}
}

func runDisassembler() {
}

func printHelp(topic string) {
	switch topic {
	default:
		fmt.Println(
`Dcpu is an assembler suite targeting the DCPU-16.

Usage:
	
	dcpu command [arguments]

The commands and their shorthands are:

	debug       d      debug a program in the emulator
	disassemble dis    converts machine code to assembler
	emulate     e      execute a program in the emulator

Use "dcpu help [command]" for more information about a command.`)
	}
}
