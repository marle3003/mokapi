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

func (e *Encoder) Number(i int) *Encoder {
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
	for i, seq := range set.Ranges {
		if i > 0 {
			e.buf = append(e.buf, ',')
		}
		if seq.Start.Value == seq.End.Value {
			e.Number(int(seq.Start.Value))
		} else if seq.End.Star {
			e.Number(int(seq.Start.Value))
			e.Byte(':')
			e.Byte('*')
		} else {
			e.Number(int(seq.Start.Value))
			e.Byte(':')
			e.Number(int(seq.End.Value))
		}
	}
	return e
}

func (e *Encoder) WriteTo(tpc *textproto.Conn) error {
	return tpc.PrintfLine(string(e.buf))
}

func (e *Encoder) String() string {
	return string(e.buf)
}
