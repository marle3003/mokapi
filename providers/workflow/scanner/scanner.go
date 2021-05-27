package scanner

import (
	"fmt"
	"mokapi/providers/workflow/token"
	"unicode"
)

type ErrorHandler func(pos token.Position, msg string)

type Scanner struct {
	pos token.Position
	src []byte

	err        ErrorHandler
	ErrorCount int

	ch            rune // current character
	offset        int  // character offset
	InsertLineEnd bool
}

func NewScanner(src []byte, err ErrorHandler) *Scanner {
	s := &Scanner{
		pos:        token.Position{Line: 1, Column: 0},
		err:        err,
		ErrorCount: 0,

		src: src,

		InsertLineEnd: false,
	}
	s.next()
	return s
}

func (s *Scanner) next() {
	if s.offset < len(s.src) {
		if s.ch == '\n' {
			s.pos.NewLine()
		} else {
			s.pos.Column++
		}

		r := rune(s.src[s.offset])
		if r == 0 {
			s.error(s.pos, "illegal character NUL")
		}
		s.ch = r
		s.offset++
	} else {
		s.offset++
		s.ch = -1 // eof
	}
}

func (s *Scanner) peek() byte {
	if s.offset < len(s.src) {
		return s.src[s.offset]
	}
	return 0
}

func (s *Scanner) error(pos token.Position, format string, args ...interface{}) {
	if s.err != nil {
		s.err(pos, fmt.Sprintf(format, args...))
	}
	s.ErrorCount++
}

func (s *Scanner) skipWhitespace() {
	for s.ch == ' ' || s.ch == '\t' || s.ch == '\r' || (s.ch == '\n' && !s.InsertLineEnd) {
		s.next()
	}
}

func (s *Scanner) scanIdentifier() string {
	offs := s.offset - 1
	for unicode.IsLetter(s.ch) || unicode.IsDigit(s.ch) || s.ch == '_' {
		s.next()
	}
	l := s.offset - 1 - offs
	if l == 0 {
		return string(s.src[offs])
	}
	return string(s.src[offs : s.offset-1])
}

func (s *Scanner) scanString() string {
	offs := s.offset

	for {
		s.next()
		if s.ch == '\n' || s.ch < 0 {
			s.error(s.pos, "string literal not terminated")
			break
		}
		if s.ch == '"' {
			break
		}
		if s.ch == '\\' {
			s.scanEscaped('"')
		}
	}

	return string(s.src[offs : s.offset-1])
}

func (s *Scanner) scanRawString() string {
	offs := s.offset

	for {
		s.next()
		if s.ch == '\n' || s.ch < 0 {
			s.error(s.pos, "string literal not terminated")
			break
		}
		if s.ch == '\'' {
			break
		}
		if s.ch == '\\' {
			s.scanEscaped('\'')
		}
	}

	return string(s.src[offs : s.offset-1])
}

func (s *Scanner) skipToLineEnd() {
	for {
		s.next()
		if s.ch == '\n' || s.ch == -1 {
			return
		}
	}
}

func (s *Scanner) scanEscaped(quote rune) bool {
	s.next()

	switch s.ch {
	case '\\', '$', quote:
		return true
	}
	s.error(s.pos, "escape sequence not terminated")
	return false
}

func (s *Scanner) UseLineEnd(b bool) {
	s.InsertLineEnd = b
}

func (s *Scanner) scanDigits() {
	for {
		s.next()
		if !unicode.IsDigit(s.ch) {
			break
		}
	}
}

func (s *Scanner) scanNumber() (token.Token, string) {
	offs := s.offset
	tok := token.INT

	s.scanDigits()

	intPart := s.offset
	if s.ch == '.' {
		tok = token.FLOAT
		s.scanDigits()
		if intPart == (s.offset - 1) {
			s.offset -= 1
			tok = token.INT
		}
	}

	return tok, string(s.src[offs-1 : s.offset-1])
}

func (s *Scanner) Scan() (pos token.Position, tok token.Token, lit string) {

	s.skipWhitespace()
	pos = s.pos

	switch ch := s.ch; {
	case unicode.IsLetter(ch):
		lit = s.scanIdentifier()
		tok = token.Loockup(lit)
	case unicode.IsDigit(ch):
		tok, lit = s.scanNumber()
	default:
		switch s.ch {
		case -1:
			tok = token.EOF
		case '\n':
			// TODO
			lit = "\\n"
		case '"':
			lit = s.scanString()
			tok = token.STRING
		case '\'':
			lit = s.scanRawString()
			tok = token.RSTRING
		case '(':
			tok = token.LPAREN
		case ')':
			tok = token.RPAREN
		case '{':
			tok = token.LBRACE
		case '}':
			tok = token.RBRACE
		case '[':
			tok = token.LBRACK
		case ']':
			tok = token.RBRACK
		case '+':
			s.next()
			if s.ch == '+' {
				tok = token.INC
			} else {
				s.offset--
				tok = token.ADD
			}
		case '-':
			s.next()
			if s.ch == '-' {
				tok = token.DEC
			} else {
				s.offset--
				tok = token.SUB
			}
		case '*':
			s.next()
			s.offset--
			tok = token.MUL
		case '/':
			s.next()
			s.offset--
			tok = token.QUO
		case '%':
			s.next()
			s.offset--
			tok = token.REM
		case '.':
			tok = token.PERIOD
			if s.ch == '.' && s.peek() == '.' {
				s.next()
				s.next()
				tok = token.ELLIPSIS
			}
		case '<':
			s.next()
			if s.ch == '=' {
				tok = token.LEQ
			} else {
				s.offset--
				tok = token.LSS
			}
		case '>':
			s.next()
			if s.ch == '=' {
				tok = token.GEQ
			} else {
				s.offset--
				tok = token.GTR
			}
		case '=':
			s.next()
			if s.ch == '=' {
				tok = token.EQL
			} else if s.ch == '>' {
				tok = token.LAMBDA
			}
		case '!':
			s.next()
			if s.ch == '=' {
				tok = token.NEQ
			} else {
				s.offset--
				tok = token.NOT
			}
		case '&':
			s.next()
			if s.ch == '&' {
				tok = token.LAND
			} else {
				s.offset--
			}
		case '|':
			s.next()
			if s.ch == '|' {
				tok = token.LOR
			} else {
				s.offset--
			}
		case ':':
			tok = token.COLON
		case ',':
			tok = token.COMMA
		default:
			lit = string(ch)
		}
		s.next()
	}

	return
}
