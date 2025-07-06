package imap

import (
	"bufio"
	"fmt"
	"github.com/pkg/errors"
	"strings"
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
		case "BODY":
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
			case "RFC822.SIZE":
				r.Options.RFC822Size = true
			case "BODYSTRUCTURE":
				r.Options.BodyStructure = true
			case "BODY.PEEK":
				body := FetchBodySection{Peek: true}
				err = body.decode(d)
				if err != nil {
					return err
				}
				r.Options.Body = append(r.Options.Body, body)
			case "BODY":
				if !d.is("[") {
					r.Options.BodyStructure = true
				} else {
					body := FetchBodySection{}
					err = body.decode(d)
					if err != nil {
						return err
					}
					r.Options.Body = append(r.Options.Body, body)
				}
			}
			return nil
		})
	}

	return r, err
}

func (s *FetchBodySection) decode(d *Decoder) error {
	var err error
	if err = d.expect("["); err != nil {
		return err
	}

	var specifier string
	s.Parts, specifier = parseSectionParts(d)

	switch strings.ToUpper(specifier) {
	case "HEADER":
		s.Specifier = "header"
	case "HEADER.FIELDS":
		s.Specifier = "header"
		err = d.SP().List(func() error {
			var field string
			field, err = d.String()
			s.Fields = append(s.Fields, field)
			return err
		})
		if err != nil {
			return err
		}
	case "TEXT":
		s.Specifier = "text"
	default:
		s.Specifier = strings.ToLower(specifier)
	}
	if err = d.expect("]"); err != nil {
		return err
	}
	if d.is("<") {
		part := BodyPart{}
		_ = d.expect("<")
		part.Offset, err = d.Number()
		if err != nil {
			return err
		}
		if err = d.expect("."); err != nil {
			return err
		}
		part.Limit, err = d.Number()
		if err != nil {
			return err
		}
		if err = d.expect(">"); err != nil {
			return err
		}
		s.Partially = &part
	}
	return nil
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

func parseSectionParts(d *Decoder) (parts []int, specifier string) {
	for {
		if len(parts) > 0 {
			if err := d.expect("."); err != nil {
				return
			}
		}
		num, err := d.Number()
		if err != nil {
			specifier, _ = d.String()
			return
		}
		parts = append(parts, int(num))
	}
}
