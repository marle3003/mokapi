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

func parseFetch(s string) (*FetchRequest, error) {
	switch s {
	case "FAST":
		return &FetchRequest{Attributes: FetchFlags | FetchInternalDate | FetchRFC822Size}, nil
	case "ALL":
		return &FetchRequest{Attributes: FetchFlags | FetchInternalDate | FetchRFC822Size | FetchEnvelope}, nil
	case "FULL":
		return &FetchRequest{Attributes: FetchFlags | FetchInternalDate | FetchRFC822Size | FetchEnvelope | FetchBodyStructure}, nil
	}

	r := &FetchRequest{}
	s = strings.Trim(s, " ")
	p := fetchParser{r: bufio.NewReader(strings.NewReader(s))}

	err := p.parseList(func() error {
		name, err := p.consume(isFetchAttrNameChar)
		if err != nil {
			return err
		}
		if len(name) == 0 {
			return nil
		}
		switch name {
		case "UID":
			r.Attributes = r.Attributes | FetchUID
		case "INTERNALDATE":
			r.Attributes = r.Attributes | FetchInternalDate
		case "RFC822.SIZE":
			r.Attributes = r.Attributes | FetchRFC822Size
		case "FLAGS":
			r.Attributes = r.Attributes | FetchFlags
		case "BODY.PEEK":
			fb, err := parseFetchBody(&p)
			if err != nil {
				return err
			}
			r.Body = fb
		}
		return nil
	}, listTypeList)

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
