package assembler

import (
	"github.com/xconstruct/dcpu16/assembler/scanner"
	"github.com/xconstruct/dcpu16/assembler/token"
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"strings"
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
	got, exp := e.Got.Tok.String(), e.Exp.Tok.String()
	if e.Got.Lit != "" {
		got += `("`+e.Got.Lit+`")`
	}
	if e.Exp.Lit != "" {
		exp += `("`+e.Exp.Lit+`")`
	}
	if e.Exp.Tok == token.EMPTY {
		return fmt.Sprintf("assembler: unexpected %s", got)
	}
	return fmt.Sprintf("assembler: unexpected %s, expected %s", got, exp)
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

type FixLabel struct {
	Offset uint16
	Label string
}

type Parser struct {
	tok TokenType
	tokens []TokenType
	offset int
	gen []uint16
	pc int
	labels map[string]uint16
	fixlabels []FixLabel
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
	offs := len(p.gen)
	op := uint16(0)
	p.gen = append(p.gen, 0x00)

	if p.tok.Tok.IsBasicOp() {
		op += uint16(OpCode(p.tok.Tok))
		p.nextImportant()
		a := p.parseValue()
		p.expect(token.COMMA)
		p.nextImportant()
		b := p.parseValue()
		op |= uint16(a) << 4
		op |= uint16(b) << 10
	} else {
		op += uint16(OpCode(p.tok.Tok)) << 4
		p.nextImportant()
		a := p.parseValue()
		op |= uint16(a) << 10
	}

	p.gen[offs] = op
}

var registers = []byte("ABCXYZIJ")

func registerOpCode(reg string) byte {
	r := strings.ToUpper(reg)[0]
	for i, ch := range registers {
		if ch == r {
			return byte(i)
		}
	}
	return 0x0
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
				default: p.unexpectedError()
				}
				lastOp = token.EMPTY
			case token.REGISTER:
				if register != "" {
					p.expect(token.RBRACK)
				}
				register = p.tok.Lit
				lastOp = token.EMPTY
			case token.ADD:
				lastOp = token.ADD
			case token.SUB:
				lastOp = token.SUB
			case token.RBRACK:
				switch {
				case register != "" && nextWord != 0:
					value = registerOpCode(register) + 0x10
					p.gen = append(p.gen, nextWord)
				case register != "":
					value = registerOpCode(register) + 0x08
				case nextWord != 0:
					value = 0x1e
					p.gen = append(p.gen, nextWord)
				default:
					p.unexpectedError()
				}
				p.nextImportant()
				break FOR
			default:
				p.unexpectedError()
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
		case token.IDENT:
			value = 0x1f
			p.fixlabels = append(p.fixlabels, FixLabel{uint16(len(p.gen)), p.tok.Lit})
			p.gen = append(p.gen, 0x0000)
		case token.INT:
			n, err := strconv.ParseUint(p.tok.Lit, 0, 16)
			if err != nil {
				panic(err)
			}
			if n <= 0x1f { // literal value 0x00-0x1f
				value = 0x20 + byte(n)
			} else { // next word (literal)
				value = 0x1f
				p.gen = append(p.gen, uint16(n))
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
	p.labels = make(map[string]uint16)
	p.fixlabels = make([]FixLabel, 0)

	p.nextImportant()
	FOR: for {
		switch {
		case p.tok.Tok.IsOp():
			p.parseOp()
		case p.tok.Tok == token.EOF:
			break FOR
		case p.tok.Tok == token.LABEL:
			if _, ok := p.labels[p.tok.Lit]; ok {
				return nil, errors.New(fmt.Sprintf(`assembler: label "%s" already defined!`, p.tok.Lit))
			}
			p.labels[p.tok.Lit] = uint16(len(p.gen))
			p.nextImportant()
		default:
			p.unexpectedError();
		}
	}

	for _, fix := range p.fixlabels {
		offset, ok := p.labels[fix.Label]
		if !ok {
			return nil, errors.New(fmt.Sprintf(`assembler: undefined label "%s"!`, p.tok.Lit))
		}
		p.gen[fix.Offset] = offset
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
