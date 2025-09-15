package imap

import (
	"fmt"
	"net/textproto"
)

type Encoder struct {
	buf []byte
}

func (e *Encoder) Atom(s string) *Encoder {
	e.buf = append(e.buf, s...)
	return e
}

func (e *Encoder) Number(i uint32) *Encoder {
	e.buf = append(e.buf, fmt.Sprintf("%d", i)...)
	return e
}

func (e *Encoder) SP() *Encoder {
	return e.Byte(' ')
}

func (e *Encoder) Byte(b byte) *Encoder {
	e.buf = append(e.buf, b)
	return e
}

func (e *Encoder) BeginList() *Encoder {
	e.buf = append(e.buf, '(')
	return e
}

func (e *Encoder) EndList() *Encoder {
	e.buf = append(e.buf, ')')
	return e
}

func (e *Encoder) ListItem(s string) *Encoder {
	if e.buf[len(e.buf)-1] != '(' {
		e.SP()
	}
	e.Atom(s)
	return e
}

func (e *Encoder) SequenceSet(set IdSet) *Encoder {
	e.buf = append(e.buf, set.String()...)
	return e
}

func (e *Encoder) WriteTo(tpc *textproto.Conn) error {
	return tpc.PrintfLine("%s", string(e.buf))
}

func (e *Encoder) String() string {
	return string(e.buf)
}
