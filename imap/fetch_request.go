package imap

import (
	"bufio"
	"fmt"
	"github.com/pkg/errors"
	"strings"
	"unicode"
)

const (
	listTypeList    = 0
	listTypeSection = 1
)

type fetchParser struct {
	r *bufio.Reader
}

func parseFetch(d *Decoder) (*FetchRequest, error) {
	r := &FetchRequest{}
	var err error

	r.Sequence, err = d.Sequence()
	if err != nil {
		return r, err
	}

	if !d.SP().is("(") {
		macro, err := d.String()
		if err != nil {
			return nil, err
		}
		switch macro {
		case "FAST":
			r.Options.Flags = true
			r.Options.InternalDate = true
			r.Options.RFC822Size = true
		case "ALL":
			r.Options.Flags = true
			r.Options.InternalDate = true
			r.Options.RFC822Size = true
			r.Options.Envelope = true
		case "FULL":
			r.Options.Flags = true
			r.Options.InternalDate = true
			r.Options.RFC822Size = true
			r.Options.Envelope = true
			r.Options.BodyStructure = true
		}
	} else {
		err = d.List(func() error {
			var key string
			key, err = d.String()
			if err != nil {
				return err
			}
			switch strings.ToUpper(key) {
			case "UID":
				r.Options.UID = true
			case "FLAGS":
				r.Options.Flags = true
			case "INTERNALDATE":
				r.Options.InternalDate = true
			case "BODYSTRUCTURE":
				r.Options.BodyStructure = true
			}
			return nil
		})
	}

	return r, err
}

func parseFetchBody(p *fetchParser) (*FetchBody, error) {
	if err := p.expect('['); err != nil {
		return nil, errors.Wrap(err, "expected section start")
	}
	section, err := p.consume(isAtomChar)
	if err != nil {
		return nil, err
	}
	fb := &FetchBody{}
	switch section {
	case "HEADER.FIELDS":
		if err := p.expect(' '); err != nil {
			return nil, err
		}
		err = p.parseList(func() error {
			field, err := p.consume(isAtomChar)
			if err != nil {
				return err
			}
			fb.HeaderFields = append(fb.HeaderFields, field)
			return nil
		}, listTypeList)
		if err != nil {
			return nil, err
		}
	}

	if err := p.expect(']'); err != nil {
		return nil, errors.Wrap(err, "expected section end")
	}

	return fb, nil
}

func (p *fetchParser) parseList(parseItem func() error, listType int) error {
	var openChar, closeChar byte
	switch listType {
	case 0:
		openChar = '('
		closeChar = ')'
	case 1:
		openChar = '['
		closeChar = ']'
	}

	if err := p.expect(openChar); err != nil {
		return errors.Wrap(err, "expected list")
	}

	if p.hasNext(closeChar) {
		return nil
	}

	for {
		if err := parseItem(); err != nil {
			return err
		}

		if p.hasNext(closeChar) {
			return nil
		}
		if err := p.expect(' '); err != nil {
			return nil
		}
	}
}

func (p *fetchParser) expect(expect byte) error {
	b, err := p.r.ReadByte()
	if err != nil {
		return err
	}
	if b != expect {
		return fmt.Errorf("expected %c, got %c", expect, b)
	}
	return nil
}

func (p *fetchParser) hasNext(b byte) bool {
	if read, err := p.r.Peek(1); err != nil {
		return false
	} else if read[0] == b {
		_, err = p.r.ReadByte()
		return err == nil
	}
	return false
}

func (p *fetchParser) consume(valid func(b byte) bool) (string, error) {
	var sb strings.Builder
	for {
		next, err := p.r.Peek(1)
		if err != nil {
			return "", err
		}
		if !valid(next[0]) {
			break
		}
		b, err := p.r.ReadByte()
		if err != nil {
			return "", err
		}
		sb.WriteByte(b)
	}
	if sb.Len() == 0 {
		return "", nil
	}
	return sb.String(), nil
}

func isFetchAttrNameChar(b byte) bool {
	return b != '[' && isAtomChar(b)
}

func isAtomChar(b byte) bool {
	switch b {
	case '(', ')', '{', ' ', '%', '*', '"', '\\', ']':
		return false
	default:
		return !unicode.IsControl(rune(b))
	}
}
