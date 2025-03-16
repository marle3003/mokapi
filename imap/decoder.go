package imap

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type Decoder struct {
	msg string
	err error
}

func NewDecoder(msg string) *Decoder {
	return &Decoder{msg: msg}
}

func (d *Decoder) SP() *Decoder {
	if d.IsSP() {
		_ = d.expect(" ")
	} else if d.is("(") {
		// SP is optional if parenthesized list follows
	} else if d.err == nil {
		d.err = fmt.Errorf("expected SP, got %s", d.msg)
	}
	return d
}

func (d *Decoder) IsSP() bool {
	return d.is(" ")
}

func (d *Decoder) ExpectSP() error {
	return d.expect(" ")
}

func (d *Decoder) ExpectTag(tag string) error {
	if d.is(tag) {
		return d.expect(tag)
	}
	return d.returnErr(fmt.Errorf("tag '%s' does not exist", tag))
}

func (d *Decoder) List(f func() error) error {
	if err := d.expect("("); err != nil {
		return d.returnErr(fmt.Errorf("expected list '()': %v", d.msg))
	}
	if d.is(")") {
		_ = d.expect(")")
		return nil
	}

	for {
		if err := f(); err != nil {
			return err
		}

		if d.is(")") {
			_ = d.expect(")")
			return nil
		}

		d.SP()
		if d.err != nil {
			return d.err
		}
	}
}

func (d *Decoder) NilList(f func() error) error {
	if d.is("NIL") {
		return d.expect("NIL")
	}
	return d.List(f)
}

func (d *Decoder) expect(s string) error {
	if d.is(s) {
		d.msg = d.msg[len(s):]
		return nil
	}
	return d.returnErr(fmt.Errorf("expected %s in %s", s, d.msg))
}

func (d *Decoder) is(s string) bool {
	if len(d.msg) == 0 {
		return false
	}
	if len(s) > len(d.msg) {
		return false
	}
	return d.msg[:len(s)] == s
}

func (d *Decoder) next() (byte, error) {
	if len(d.msg) == 0 {
		return 0, io.EOF
	}
	return d.msg[0], nil
}

func (d *Decoder) readByte() (byte, error) {
	if len(d.msg) == 0 {
		return 0, io.EOF
	}
	b := d.msg[0]
	d.msg = d.msg[1:]
	return b, nil
}

func (d *Decoder) Quoted() (string, error) {
	if err := d.expect("\""); err != nil {
		return "", d.returnErr(err)
	}

	var sb strings.Builder
	for {
		if d.is("\"") {
			return sb.String(), d.expect("\"")
		}

		b, err := d.readByte()
		if err != nil {
			return "", d.returnErr(err)
		}
		sb.WriteByte(b)
	}
}

func (d *Decoder) String() (string, error) {
	if d.is("\"") {
		return d.Quoted()
	}
	return d.Read(isAtom)
}

func (d *Decoder) NilString() (*string, error) {
	if d.is("NIL") {
		return nil, d.expect("NIL")
	}
	s, err := d.String()
	return &s, err
}

func (d *Decoder) Text() (string, error) {
	return d.Read(func(r byte) bool {
		return true
	})
}

func (d *Decoder) Pattern() (string, error) {
	if d.is("\"") {
		return d.Quoted()
	}
	return d.Read(func(r byte) bool {
		return r != ' '
	})
}

func (d *Decoder) Number() (uint32, error) {
	s, err := d.Read(func(r byte) bool {
		return r >= '0' && r <= '9'
	})
	if err != nil {
		return 0, d.returnErr(err)
	}
	return parseNum(s)
}

func (d *Decoder) Read(valid func(r byte) bool) (string, error) {
	var sb strings.Builder
	for {
		b, err := d.next()
		if err == io.EOF {
			return sb.String(), nil
		}

		if !valid(b) {
			return sb.String(), nil
		}

		sb.WriteByte(b)
		_, _ = d.readByte()
	}
}

func (d *Decoder) ReadFlag() (string, error) {
	var flag string
	isSystem := d.is("\\")
	if isSystem {
		_, _ = d.readByte()
	}
	flag, err := d.Read(isAtom)
	if err != nil {
		return "", d.returnErr(err)
	}
	if isSystem {
		flag = "\\" + flag
	}
	return flag, d.returnErr(err)
}

func (d *Decoder) EndCmd(tag string) error {
	if err := d.expect(tag); err != nil {
		return d.returnErr(err)
	}

	if status, err := d.SP().String(); err != nil {
		return d.returnErr(err)
	} else {
		switch status {
		case "OK":
			return nil
		case "NO", "BAD":
			if text, err := d.SP().Text(); err != nil {
				return d.returnErr(err)
			} else {
				return fmt.Errorf("imap status [%v]: %v", status, text)
			}
		default:
			return fmt.Errorf("imap unknown status: %s", status)
		}
	}
}

func (d *Decoder) Sequence() (IdSet, error) {
	set := IdSet{}
	s, err := d.Read(func(r byte) bool { return r == '*' || isAtom(r) })
	if err != nil {
		return set, d.returnErr(err)
	}

	for _, v := range strings.Split(s, ",") {
		if i := strings.IndexRune(v, ':'); i >= 0 {
			r := &Range{}
			r.Start, err = parseNumSet(v[:i])
			if err != nil {
				return set, d.returnErr(err)
			}
			r.End, err = parseNumSet(v[i+1:])
			if err != nil {
				return set, d.returnErr(err)
			}
			set.Ids = append(set.Ids, r)
		} else {
			var n SeqNum
			n, err = parseNumSet(v)
			if err != nil {
				return set, d.returnErr(err)
			}
			set.Ids = append(set.Ids, IdNum(n.Value))
		}
	}

	return set, nil
}

func parseNumSet(s string) (SeqNum, error) {
	numSet := SeqNum{}
	if s == "*" {
		numSet.Star = true
	} else {
		var err error
		numSet.Value, err = parseNum(s)
		return numSet, err
	}
	return numSet, nil
}

func (d *Decoder) Date() (time.Time, error) {
	s, err := d.Quoted()
	if err != nil {
		return time.Time{}, d.returnErr(err)
	}
	return time.Parse(DateTimeLayout, s)
}

func (d *Decoder) DiscardValue() error {
	if d.is("(") {
		err := d.List(func() error {
			return d.DiscardValue()
		})
		return d.returnErr(err)
	}
	_, err := d.Read(isAtom)
	return d.returnErr(err)
}

func (d *Decoder) returnErr(err error) error {
	if d.err == nil {
		d.err = err
	}
	return d.err
}

func (d *Decoder) Error() error {
	return d.err
}

func isAtom(b byte) bool {
	switch b {
	case '(', ')', '{', '}', ' ', '%', '*', '"', '\\', '[', ']':
		return false
	default:
		return !unicode.IsControl(rune(b))
	}
}

func parseNum(s string) (uint32, error) {
	v, err := strconv.ParseUint(s, 0, 32)
	if err != nil {
		return 0, err
	}
	return uint32(v), nil
}
