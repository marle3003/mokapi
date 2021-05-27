package parser

import (
	"fmt"
	"github.com/pkg/errors"
	"mokapi/providers/workflow/token"
	"strings"
)

type Error struct {
	Pos token.Position
	Msg string
}

func (e Error) String() string {
	return fmt.Sprintf("%v:%v: %v", e.Pos.Line, e.Pos.Column, e.Msg)
}

type ErrorList []*Error

func (l *ErrorList) Add(pos token.Position, msg string) {
	*l = append(*l, &Error{Pos: pos, Msg: msg})
}

func (l *ErrorList) Addf(pos token.Position, format string, args ...interface{}) {
	*l = append(*l, &Error{Pos: pos, Msg: fmt.Sprintf(format, args...)})
}

func (l ErrorList) Err() error {
	if len(l) == 0 {
		return nil
	}
	sb := strings.Builder{}
	for i, e := range l {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(e.String())
	}
	return errors.New(sb.String())
}
