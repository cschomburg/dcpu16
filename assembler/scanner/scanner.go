package scanner

import (
	"unicode"
	"dcpu16/assembler/token"
)

type Scanner struct {
	src []byte
	ch rune
	offset int
}

func (s *Scanner) Init(src []byte) {
	s.src = src
	s.ch = ' '
	s.offset = -1

	s.next()
}

func (s *Scanner) next() {
	rdOffset := s.offset + 1
	if rdOffset < len(s.src) {
		s.ch = rune(s.src[rdOffset])
		s.offset = rdOffset
	} else {
		s.offset = len(s.src)
		s.ch = -1 // EOF
	}
}

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}

func (s *Scanner) skipWhitespace() {
	for s.ch == ' ' || s.ch == '\t' || s.ch == '\n' || s.ch == '\r' {
		s.next()
	}
}

func (s *Scanner) scanIdentifier() string {
	offs := s.offset
	for isLetter(s.ch) || isDigit(s.ch) {
		s.next()
	}
	return string(s.src[offs:s.offset])
}

func digitVal(ch rune) int {
	switch {
	case '0' <= ch && ch <= '9':
		return int(ch - '0')
	case 'a' <= ch && ch <= 'f':
		return int(ch - 'a' + 10)
	case 'A' <= ch && ch <= 'F':
		return int(ch - 'A' + 10)
	}
	return 16 // larger than any legal digit val
}

func (s *Scanner) scanMantissa(base int) {
	for digitVal(s.ch) < base {
		s.next()
	}
}

func (s *Scanner) scanNumber() (token.Token, string) {
	offs := s.offset
	tok := token.INT

	if s.ch == '0' {
		s.next()
		if s.ch == 'x' || s.ch == 'X' {
			// hexadecimal int
			s.next()
			s.scanMantissa(16)
		} else {
			// octal int
			s.scanMantissa(8)
		}
	} else {
		s.scanMantissa(10)
	}

	return tok, string(s.src[offs:s.offset])
}

func (s *Scanner) scanComment() string {
	offs := s.offset
	for s.ch != '\n' {
		s.next()
	}
	return string(s.src[offs:s.offset])
}

func isRegister(ident string) bool {
	if len(ident) != 1 {
		return false
	}
	ch := unicode.ToUpper(rune(ident[0]))
	return (ch >= 'A' && ch <= 'C' || ch >= 'X' && ch <= 'Z' || ch == 'I' || ch == 'J')
}

func (s *Scanner) Scan() (tok token.Token, lit string) {
	s.skipWhitespace()

	switch ch := s.ch; {
	case isLetter(ch):
		lit = s.scanIdentifier()
		if isRegister(lit) {
			tok = token.REGISTER
		} else {
			tok = token.Lookup(lit)
		}
	case digitVal(ch) < 10:
		tok, lit = s.scanNumber()
	case ch == -1:
		tok = token.EOF
	default:
		s.next();
		switch ch {
		case ':':
			lit = s.scanIdentifier()
			tok = token.LABEL
		case '[':
			tok = token.LBRACK
		case ']':
			tok = token.RBRACK
		case ';':
			tok = token.COMMENT
			lit = s.scanComment()
		case ',':
			tok = token.COMMA
		case '+':
			tok = token.ADD
		case '-':
			tok = token.SUB
		default:
			tok = token.ILLEGAL
			lit = string(ch)
		}
	}

	return
}
