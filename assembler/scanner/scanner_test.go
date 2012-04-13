package scanner

import (
	"testing"
	"dcpu16/assembler/token"
)

func scanExpect(t *testing.T, s *Scanner, expTok token.Token, expLit string) {
	gotTok, gotLit := s.Scan()
	if expTok != gotTok || expLit != gotLit {
		t.Fatalf("Expected %s '%s', but got %s '%s'\n", expTok, expLit, gotTok, gotLit)
	}
}

func TestOp(t *testing.T) {
	s := &Scanner{}
	s.Init([]byte(`SET A, B`))

	scanExpect(t, s, token.OP_SET, "SET")
	scanExpect(t, s, token.REGISTER, "A")
	scanExpect(t, s, token.COMMA, "")
	scanExpect(t, s, token.REGISTER, "B")
	scanExpect(t, s, token.EOF, "")
	scanExpect(t, s, token.EOF, "")
}

func TestWhitespaceComment(t *testing.T) {
	s := &Scanner{}
	s.Init([]byte(`
		SET A, B ; This is a comment
			MUL   A,   C ; DIV A, C
		SET 	 PUSH, SP;
		JSR J
	`))

	scanExpect(t, s, token.OP_SET, "SET")
	scanExpect(t, s, token.REGISTER, "A")
	scanExpect(t, s, token.COMMA, "")
	scanExpect(t, s, token.REGISTER, "B")
	scanExpect(t, s, token.COMMENT, " This is a comment")

	scanExpect(t, s, token.OP_MUL, "MUL")
	scanExpect(t, s, token.REGISTER, "A")
	scanExpect(t, s, token.COMMA, "")
	scanExpect(t, s, token.REGISTER, "C")
	scanExpect(t, s, token.COMMENT, " DIV A, C")

	scanExpect(t, s, token.OP_SET, "SET")
	scanExpect(t, s, token.PUSH, "PUSH")
	scanExpect(t, s, token.COMMA, "")
	scanExpect(t, s, token.SP, "SP")
	scanExpect(t, s, token.COMMENT, "")

	scanExpect(t, s, token.OP_JSR, "JSR")
	scanExpect(t, s, token.REGISTER, "J")

	scanExpect(t, s, token.EOF, "")
}

func TestNotch(t *testing.T) {
	s := &Scanner{}
	s.Init([]byte(`
	; Try some basic stuff
				  SET A, 0x30              ; 7c01 0030
				  SET [0x1000], 0x20       ; 7de1 1000 0020
				  SUB A, [0x1000]          ; 7803 1000
				  IFN A, 0x10              ; c00d 
					 SET PC, crash         ; 7dc1 001a [*]
				  
	; Do a loopy thing
				  SET I, 10                ; a861
				  SET A, 0x2000            ; 7c01 2000
	:loop         SET [0x2000+I], [A]      ; 2161 2000
				  SUB I, 1                 ; 8463
				  IFN I, 0                 ; 806d
					 SET PC, loop          ; 7dc1 000d [*]
	`))

	scanExpect(t, s, token.COMMENT, " Try some basic stuff")
	scanExpect(t, s, token.OP_SET, "SET")
	scanExpect(t, s, token.REGISTER, "A")
	scanExpect(t, s, token.COMMA, "")
	scanExpect(t, s, token.INT, "0x30")
	scanExpect(t, s, token.COMMENT, " 7c01 0030")

	scanExpect(t, s, token.OP_SET, "SET")
	scanExpect(t, s, token.LBRACK, "")
	scanExpect(t, s, token.INT, "0x1000")
	scanExpect(t, s, token.RBRACK, "")
	scanExpect(t, s, token.COMMA, "")
	scanExpect(t, s, token.INT, "0x20")
	scanExpect(t, s, token.COMMENT, " 7de1 1000 0020")

	scanExpect(t, s, token.OP_SUB, "SUB")
	scanExpect(t, s, token.REGISTER, "A")
	scanExpect(t, s, token.COMMA, "")
	scanExpect(t, s, token.LBRACK, "")
	scanExpect(t, s, token.INT, "0x1000")
	scanExpect(t, s, token.RBRACK, "")
	scanExpect(t, s, token.COMMENT, " 7803 1000")

	scanExpect(t, s, token.OP_IFN, "IFN")
	scanExpect(t, s, token.REGISTER, "A")
	scanExpect(t, s, token.COMMA, "")
	scanExpect(t, s, token.INT, "0x10")
	scanExpect(t, s, token.COMMENT, " c00d ")

	scanExpect(t, s, token.OP_SET, "SET")
	scanExpect(t, s, token.PC, "PC")
	scanExpect(t, s, token.COMMA, "")
	scanExpect(t, s, token.IDENT, "crash")
	scanExpect(t, s, token.COMMENT, " 7dc1 001a [*]")

	scanExpect(t, s, token.COMMENT, " Do a loopy thing")
	scanExpect(t, s, token.OP_SET, "SET")
	scanExpect(t, s, token.REGISTER, "I")
	scanExpect(t, s, token.COMMA, "")
	scanExpect(t, s, token.INT, "10")
	scanExpect(t, s, token.COMMENT, " a861")

	scanExpect(t, s, token.OP_SET, "SET")
	scanExpect(t, s, token.REGISTER, "A")
	scanExpect(t, s, token.COMMA, "")
	scanExpect(t, s, token.INT, "0x2000")
	scanExpect(t, s, token.COMMENT, " 7c01 2000")

	scanExpect(t, s, token.LABEL, "loop")
	scanExpect(t, s, token.OP_SET, "SET")
	scanExpect(t, s, token.LBRACK, "")
	scanExpect(t, s, token.INT, "0x2000")
	scanExpect(t, s, token.ADD, "")
	scanExpect(t, s, token.REGISTER, "I")
	scanExpect(t, s, token.RBRACK, "")
	scanExpect(t, s, token.COMMA, "")
	scanExpect(t, s, token.LBRACK, "")
	scanExpect(t, s, token.REGISTER, "A")
	scanExpect(t, s, token.RBRACK, "")
	scanExpect(t, s, token.COMMENT, " 2161 2000")

	scanExpect(t, s, token.OP_SUB, "SUB")
	scanExpect(t, s, token.REGISTER, "I")
	scanExpect(t, s, token.COMMA, "")
	scanExpect(t, s, token.INT, "1")
	scanExpect(t, s, token.COMMENT, " 8463")

	scanExpect(t, s, token.OP_IFN, "IFN")
	scanExpect(t, s, token.REGISTER, "I")
	scanExpect(t, s, token.COMMA, "")
	scanExpect(t, s, token.INT, "0")
	scanExpect(t, s, token.COMMENT, " 806d")

	scanExpect(t, s, token.OP_SET, "SET")
	scanExpect(t, s, token.PC, "PC")
	scanExpect(t, s, token.COMMA, "")
	scanExpect(t, s, token.IDENT, "loop")
	scanExpect(t, s, token.COMMENT, " 7dc1 000d [*]")
	scanExpect(t, s, token.EOF, "")
}
