package main

import (
	"github.com/xconstruct/dcpu16/assembler"
	"github.com/xconstruct/dcpu16/debugger"
	"github.com/xconstruct/dcpu16/emulator"
	"github.com/xconstruct/dcpu16/words"
	"flag"
	"fmt"
	"os"
	"io"
	"io/ioutil"
	"bufio"
)

func assert(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {

	tool := ""
	if len(os.Args) > 1 {
		tool = os.Args[1]
	}
	switch tool {
	case "a": fallthrough
	case "assemble":
		runAssembler();
	case "d": fallthrough
	case "debug":
		runDebugger()
	case "dis": fallthrough
	case "disassemble":
		runDisassembler()
	case "e": fallthrough
	case "emulate":
		runEmulator()
	case "h": fallthrough
	case "hexdump":
		runHexdump()
	default:
		if len(os.Args) > 2 {
			printHelp(os.Args[2]);
		} else {
			printHelp("");
		}
	}
}

func runEmulator() {
	flag.Parse()
	path := flag.Arg(1)
	if path == "" {
		printHelp("emulate")
		return
	}

	file, err := os.Open(path)
	assert(err)

	dcpu := emulator.NewDCPU()
	ram := words.NewReadWriter(dcpu.RAM)
	_, err = io.Copy(ram, file)
	file.Close()
	assert(err)
	err = dcpu.Exec()
	assert(err)
}

func runDebugger() {
	flag.Parse()
	path := flag.Arg(1)
	if path == "" {
		printHelp("debug")
		return
	}

	file, err := os.Open(path)
	assert(err)

	dcpu := emulator.NewDCPU()
	ram := words.NewReadWriter(dcpu.RAM)
	_, err = io.Copy(ram, file)
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
		case "mem": words.Hexdump(dcpu.RAM, os.Stdout)
		case "r":   debugger.RDump(dcpu)
		case "op":  debugger.PrintInstruction(dcpu)
		}
	}
}

func runDisassembler() {
}

func runAssembler() {
	flag.Parse()
	srcPath := flag.Arg(1)
	if srcPath == "" {
		printHelp("assemble")
		return
	}
	src, err := ioutil.ReadFile(srcPath)
	assert(err)

	var destWriter io.Writer
	destPath := flag.Arg(2)
	if destPath == "" {
		destWriter = os.Stdout
	} else {
		destWriter, err = os.Open(destPath)
		assert(err)
	}

	gen, err := assembler.Assemble(src)
	assert(err)
	genReader := words.NewReadWriter(gen)
	_, err = io.Copy(destWriter, genReader)
	assert(err)
}

func runHexdump() {
	flag.Parse()
	srcPath := flag.Arg(1)
	if srcPath == "" {
		printHelp("hexdump")
		return
	}
	src, err := ioutil.ReadFile(srcPath)
	assert(err)
	w := make([]uint16, len(src)/2)
	words.CopyFromBytes(w, src)
	words.Hexdump(w, os.Stdout)
}

func printHelp(topic string) {
	switch topic {
	case "assemble":
		fmt.Println(`Usage: dcpu assemble dasmfile [binfile]`)
	case "debug":
		fmt.Println("Usage: dcpu debug binfile")
	case "emulate":
		fmt.Println("Usage: dcpu emulate binfile")
	case "hexdump":
		fmt.Println(`Usage: dcpu hexdump binfile`)
	default:
		fmt.Println(`Dcpu16 is an assembler suite targeting the DCPU-16.

Usage:
	
	dcpu16 command [arguments]

The commands and their shorthands are:

	assemble    a      converts assembler to machine code
	debug       d      debug a program in the emulator
	disassemble dis    converts machine code to assembler
	emulate     e      execute a program in the emulator
	hexdump     h      display a binary file in readable format

Use "dcpu help [command]" for more information about a command.`)
	}
}
