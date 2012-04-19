package token

import (
	"strings"
	"strconv"
)

type Token int

const (
	// Special tokens
	EMPTY Token = iota
	ILLEGAL
	EOF
	COMMENT
	LABEL

	// Values
	IDENT
	REGISTER
	INT
	keyword_beg
	SP
	PC
	PUSH
	PEEK
	POP
	O

	// Basic opcodes
	OP_SET
	OP_ADD
	OP_SUB
	OP_MUL
	OP_DIV
	OP_MOD
	OP_SHL
	OP_SHR
	OP_AND
	OP_BOR
	OP_XOR
	OP_IFE
	OP_IFN
	OP_IFG
	OP_IFB

	// Non-basic opcodes
	OP_JSR

	// extended
	OP_DAT
	keyword_end

	// Delimiters and misc
	COMMA
	LBRACK
	RBRACK
	ADD
	SUB
)

var tokens = [...]string{
	EMPTY: "EMPTY",
	ILLEGAL: "ILLEGAL",
	EOF: "EOF",
	COMMENT: "COMMENT",
	LABEL: "LABEL",

	// Values,
	IDENT: "IDENT",
	REGISTER: "REGISTER",
	INT: "INT",
	SP: "SP",
	PC: "PC",
	PUSH: "PUSH",
	PEEK: "PEEK",
	POP: "POP",
	O: "O",

	// Basic opcodes,
	OP_SET: "SET",
	OP_ADD: "ADD",
	OP_SUB: "SUB",
	OP_MUL: "MUL",
	OP_DIV: "DIV",
	OP_MOD: "MOD",
	OP_SHL: "SHL",
	OP_SHR: "SHR",
	OP_AND: "AND",
	OP_BOR: "BOR",
	OP_XOR: "XOR",
	OP_IFE: "IFE",
	OP_IFN: "IFN",
	OP_IFG: "IFG",
	OP_IFB: "IFB",

	// Non-basic opcodes,
	OP_JSR: "JSR",

	// extended
	OP_DAT: "DAT",

	// Delimiters and misc,
	COMMA: "','",
	LBRACK: "'['",
	RBRACK: "']'",
	ADD: "'+'",
	SUB: "'-'",
}

func (tok Token) String() string {
	s := ""
	if 0 <= tok && tok < Token(len(tokens)) {
		s = tokens[tok]
	}
	if s == "" {
		s = "token(" + strconv.Itoa(int(tok)) + ")"
	}
	return s
}

func (tok Token) IsBasicOp() bool {
	return tok >= OP_SET && tok <= OP_IFB
}

func (tok Token) IsComplexOp() bool {
	return tok == OP_JSR
}

func (tok Token) IsOp() bool {
	return tok.IsBasicOp() || tok.IsComplexOp()
}

var keywords map[string]Token

func init() {
	keywords = make(map[string]Token)
	for i := keyword_beg + 1; i < keyword_end; i++ {
		keywords[tokens[i]] = i
	}
}

func Lookup (ident string) Token {
	if tok, is_keyword := keywords[ident]; is_keyword {
		return tok
	}
	ident = strings.ToLower(ident)
	if tok, is_keyword := keywords[ident]; is_keyword {
		return tok
	}
	return IDENT
}
