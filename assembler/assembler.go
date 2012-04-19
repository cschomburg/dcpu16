package assembler

import (
	"github.com/xconstruct/dcpu16/assembler/scanner"
	"github.com/xconstruct/dcpu16/assembler/token"
	"fmt"
	"runtime"
	"strconv"
)

type TokenType struct {
	Tok token.Token
	Lit string
}

type UnexpectedTokenError struct {
	Got TokenType
	Exp TokenType
}

func (e *UnexpectedTokenError) Error() string {
	if e.Exp.Tok == token.ILLEGAL {
		return fmt.Sprintf("assembler: unexpected %s", e.Got.Tok)
	}
	return fmt.Sprintf("assembler: unexpected %s, expected %s", e.Got.Tok, e.Exp.Tok)
}

type ExpectedTokenError struct {
	exp TokenType
	got TokenType
}

func OpCode(tok token.Token) byte {
	switch {
	case tok.IsBasicOp():
		return byte(tok - token.OP_SET + 1)
	case tok.IsComplexOp():
		return byte(tok - token.OP_JSR + 1)
	}
	return 0x0;
}

type Parser struct {
	tok TokenType
	tokens []TokenType
	offset int
	gen []uint16
	pc int
}

func (p *Parser) next() {
	rdOffset := p.offset + 1
	if rdOffset < len(p.tokens) {
		p.tok = p.tokens[rdOffset]
		p.offset = rdOffset
	} else {
		p.offset = len(p.tokens)
		p.tok = TokenType{token.EOF, ""}
	}
}

func (p *Parser) skipIgnored() {
	for p.tok.Tok == token.COMMENT {
		p.next()
	}
}

func (p *Parser) nextImportant() {
	p.next()
	p.skipIgnored()
}

func (p *Parser) expect(tok token.Token) {
	if p.tok.Tok != tok {
		panic(&UnexpectedTokenError{p.tok, TokenType{tok, ""}})
	}
}

func (p *Parser) unexpectedError() {
	panic(&UnexpectedTokenError{p.tok, TokenType{}})
}

func (p *Parser) parseOp() {
	offs := p.pc
	op := uint16(OpCode(p.tok.Tok))
	p.pc++

	if p.tok.Tok.IsBasicOp() {
		p.nextImportant()
		a := p.parseValue()
		p.expect(token.COMMENT)
		p.nextImportant()
		b := p.parseValue()
		op |= uint16(a) << 4
		op |= uint16(b) << 10
	} else {
		a := p.parseValue()
		op |= uint16(a) << 4
	}

	p.gen[offs] = op
}

func registerOpCode(reg string) byte {
	reg = strings.ToUpper(reg)
	op := reg[0] - 'A'
	if op < 0x00 || op > 0x07 {
		return 0x00
	}
	return op
}

func (p *Parser) parseValue() (value byte) {
	if p.tok.Tok == token.LBRACK { // indirect
		p.nextImportant()

		var register string
		lastOp := token.ADD
		var nextWord uint16

		FOR: for {
			switch p.tok.Tok {
			case token.INT:
				n, err := strconv.ParseUint(p.tok.Lit, 0, 16)
				if err != nil {
					panic(err)
				}
				switch lastOp {
				case token.ADD: nextWord += uint16(n)
				case token.SUB: nextWord -= uint16(n)
				default: unexpectedError()
				}
				lastOp = token.EMPTY
			case token.REGISTER:
				if register != "" {
					expect(token.RBRACK)
				}
				register = register.Lit
				lastOp = token.EMPTY
			case token.ADD:
				lastOp = token.ADD
			case token.SUB:
				lastOp = token.SUB
			case token.RBRACK:
				switch {
				case register != "" && nextWord != 0:
					value = registerOpCode(register) + 0x10
					p.gen[p.pc] = nextWord
					p.pc++
				case register != "":
					value = registerOpCode(register) + 0x08
				case nextWord != 0:
					value = 0x1e
					p.gen[p.pc] = nextWord
					p.pc++
				default:
					unexpectedError()
				}
				p.nextImportant()
				break FOR
			default:
				unexpectedError()
			}
			p.nextImportant()
		}
	} else {
		switch p.tok.Tok {
		case token.POP: value = 0x18
		case token.PEEK: value = 0x19
		case token.PUSH: value = 0x1a
		case token.SP: value = 0x1b
		case token.PC: value = 0x1c
		case token.O: value = 0x1d
		case token.REGISTER: value = registerOpCode(p.tok.Lit)
		case token.INT:
			n, err := strconv.ParseUint(p.tok.Lit, 0, 16)
			if err != nil {
				panic(err)
			}
			if n <= 0x1f { // literal value 0x00-0x1f
				value = 0x20 + byte(n)
			} else { // next word (literal)
				value = 0x1f
				p.gen[p.pc] = uint16(n)
				p.pc++
			}
		default:
			p.unexpectedError()
		}
		p.nextImportant()
	}

	return
}

func (p *Parser) Parse(tokens []TokenType) (gen []uint16, err error) {
	defer func() {
		if r:= recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()

	p.tokens = tokens
	p.offset = -1
	p.gen = make([]uint16, 0)
	p.pc = 0

	p.nextImportant()
	for {
		switch {
		case p.tok.Tok.IsOp():
			p.parseOp()
		default:
			unexpectedError();
		}
	}

	return p.gen, nil
}

func Assemble(src []byte) (gen []uint16, err error) {
	s := &scanner.Scanner{}
	s.Init(src)

	tokens := make([]TokenType, 0)
	for {
		tok, lit := s.Scan()
		tokens = append(tokens, TokenType{tok, lit})
		if tok == token.EOF {
			break
		}
	}

	parser := &Parser{}
	gen, err = parser.Parse(tokens)
	return
}
