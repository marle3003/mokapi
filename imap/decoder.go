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

func (d *Decoder) SP() *Decoder {
	if d.IsSP() {
		_ = d.expect(" ")
	} else if d.is("(") {
		// SP is optional if a parenthesized list follows
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

func (d *Decoder) IsList() bool {
	return d.is("(")
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
	return d.Read(isAtom), nil
}

func (d *Decoder) Atom() string {
	return d.Read(isAtom)
}

func (d *Decoder) NilString() (*string, error) {
	if d.is("NIL") {
		return nil, d.expect("NIL")
	}
	s, err := d.String()
	return &s, err
}

func (d *Decoder) Text() string {
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
	}), nil
}

func (d *Decoder) Number() (uint32, error) {
	s := d.Read(func(r byte) bool {
		return r >= '0' && r <= '9'
	})
	return parseNum(s)
}

func (d *Decoder) Int64() (int64, error) {
	s := d.Read(func(r byte) bool {
		return r >= '0' && r <= '9'
	})
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, d.returnErr(err)
	}
	return i, nil
}

func (d *Decoder) Read(valid func(r byte) bool) string {
	var sb strings.Builder
	for {
		b, err := d.next()
		if err == io.EOF {
			return sb.String()
		}

		if !valid(b) {
			return sb.String()
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
	flag = d.Read(isAtom)
	if isSystem {
		flag = "\\" + flag
	}
	return flag, nil
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
			text := d.SP().Text()
			return fmt.Errorf("imap status [%v]: %v", status, text)
		default:
			return fmt.Errorf("imap unknown status: %s", status)
		}
	}
}

func (d *Decoder) Sequence() (IdSet, error) {
	s := d.Read(func(r byte) bool { return r == '*' || isAtom(r) })

	set, err := parseSequence(s)
	if err != nil {
		return set, d.returnErr(err)
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
	_ = d.Read(isAtom)
	return d.err
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
