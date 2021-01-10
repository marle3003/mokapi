package lang

import (
	"fmt"
	"unicode"
)

type ErrorHandler func(pos Position, msg string)

type Scanner struct {
	pos Position
	src []byte

	err        ErrorHandler
	ErrorCount int

	ch            rune // current character
	offset        int  // character offset
	insertLineEnd bool
}

type Position struct {
	Line   int
	Column int
}

func NewScanner(src []byte, err ErrorHandler) *Scanner {
	keywords = make(map[string]Token)
	for i := keywordsStart + 1; i < keywordsEnd; i++ {
		keywords[tokens[i]] = i
	}
	s := &Scanner{
		pos:        Position{Line: 1, Column: 0},
		err:        err,
		ErrorCount: 0,

		src: src,

		insertLineEnd: false,
	}
	s.next()
	return s
}

func (s *Scanner) next() {
	if s.offset < len(s.src) {
		if s.ch == '\n' {
			s.pos.newLine()
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
		s.ch = -1 // eof
	}
}

func (s *Scanner) error(pos Position, format string, args ...interface{}) {
	if s.err != nil {
		s.err(pos, fmt.Sprintf(format, args))
	}
	s.ErrorCount++
}

func (s *Scanner) skipWhitespace() {
	for s.ch == ' ' || s.ch == '\t' || s.ch == '\r' || (s.ch == '\n' && !s.insertLineEnd) {
		s.next()
	}
}

func (s *Scanner) scanIdentifier() string {
	offs := s.offset - 1
	for unicode.IsLetter(s.ch) || unicode.IsDigit(s.ch) {
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
	s.insertLineEnd = b
}

func (s *Scanner) scanNumber() string {
	offs := s.offset

	for {
		s.next()
		if !unicode.IsDigit(s.ch) {
			break
		}
	}

	return string(s.src[offs-1 : s.offset-1])
}

func (s *Scanner) Scan() (pos Position, tok Token, lit string) {

	s.skipWhitespace()
	pos = s.pos

	switch ch := s.ch; {
	case unicode.IsLetter(ch):
		lit = s.scanIdentifier()
		tok = lockup(lit)
	case unicode.IsDigit(ch):
		lit = s.scanNumber()
		tok = NUMBER
	default:
		switch s.ch {
		case -1:
			tok = EOF
		case '\n':
			tok = SEMICOLON
			lit = "\\n"
		case '"':
			lit = s.scanString()
			tok = STRING
		case '\'':
			lit = s.scanRawString()
			tok = RSTRING
		case '(':
			tok = LPAREN
		case ')':
			tok = RPAREN
		case '{':
			tok = LBRACE
		case '}':
			tok = RBRACE
		case '[':
			tok = LBRACK
		case ']':
			tok = RBRACK
		case '+':
			s.next()
			if s.ch == '=' {
				tok = ADD_ASSIGN
			} else {
				s.offset--
				tok = ADD
			}
		case '-':
			s.next()
			if s.ch == '=' {
				tok = SUB_ASSIGN
			} else {
				s.offset--
				tok = SUB
			}
		case '*':
			s.next()
			if s.ch == '=' {
				tok = MUL_ASSIGN
			} else {
				s.offset--
				tok = MUL
			}
		case '/':
			s.next()
			if s.ch == '/' {
				s.skipToLineEnd()
				return s.Scan()
			} else if s.ch == '=' {
				tok = QUO_ASSIGN
			} else {
				s.offset--
				tok = QUO
			}
		case '%':
			s.next()
			if s.ch == '=' {
				tok = REM_ASSIGN
			} else {
				s.offset--
				tok = REM
			}
		case '.':
			tok = PERIOD
		case '<':
			s.next()
			if s.ch == '=' {
				tok = LEQ
			} else {
				s.offset--
				tok = LSS
			}
		case '>':
			s.next()
			if s.ch == '=' {
				tok = GEQ
			} else {
				s.offset--
				tok = GTR
			}
		case '=':
			s.next()
			if s.ch == '=' {
				tok = EQL
			} else if s.ch == '>' {
				tok = LAMBDA
			} else {
				s.offset--
				tok = ASSIGN
			}
		case '!':
			s.next()
			if s.ch == '=' {
				tok = NEQ
			} else {
				s.offset--
				tok = NOT
			}
		case '&':
			s.next()
			if s.ch == '&' {
				tok = LAND
			} else {
				s.offset--
			}
		case '|':
			s.next()
			if s.ch == '|' {
				tok = LOR
			} else {
				s.offset--
			}
		case ':':
			s.next()
			if s.ch == '=' {
				tok = DEFINE
			} else {
				s.offset--
				tok = COLON
			}
		case ',':
			tok = COMMA
		default:
			lit = string(ch)
		}
		s.next()
	}

	return
}

func (p *Position) newLine() {
	p.Line++
	p.Column = 0
}
