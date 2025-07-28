package imap

import (
	"strings"
)

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
