package imap

import (
	"fmt"
	"strings"
	"time"
)

type FetchOptions struct {
	UID           bool
	Flags         bool
	InternalDate  bool
	RFC822Size    bool
	Envelope      bool
	BodyStructure bool
	Body          []FetchBodySection
}

type FetchBodySection struct {
	Type   string
	Fields []string
	Parts  []int
	Peek   bool
}

type IdSet struct {
	Ranges []Range
	IsUid  bool
}

type Range struct {
	Start SeqNum
	End   SeqNum
}

type SeqNum struct {
	Value uint32
	Star  bool
}

type FetchBody struct {
	Section      []int
	HeaderFields []string
}

type FetchRequest struct {
	Sequence IdSet
	Options  FetchOptions
}

func (c *conn) handleFetch(tag, param string) error {
	if c.state != AuthenticatedState && c.state != SelectedState {
		return c.writeResponse(tag, &response{
			status: bad,
			text:   "Command is only valid in authenticated state",
		})
	}

	d := Decoder{msg: param}

	req, err := parseFetch(&d)
	if err != nil {
		return err
	}

	res := fetchResponse{}
	if err = c.handler.Fetch(req, &res, c.ctx); err != nil {
		return err
	}

	if err = c.writeFetchResponse(&res); err != nil {
		return err
	}

	return c.writeResponse(tag, &response{
		status: ok,
		text:   "FETCH completed",
	})
}

func (c *conn) writeFetchResponse(res *fetchResponse) error {
	for _, msg := range res.messages {
		m := strings.Trim(msg.sb.String(), " ")
		err := c.writeResponse(untagged, &response{
			text: fmt.Sprintf("%v FETCH (%v)", msg.sequenceNumber, m),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

type FetchCommand struct {
	Messages []Message
}

type Message struct {
	SeqNumber     uint32
	UID           uint32
	Flags         []Flag
	InternalDate  time.Time
	Size          uint32
	BodyStructure BodyStructure
	Body          []FetchData
}

func (c *Client) Fetch(set IdSet, options FetchOptions) (*FetchCommand, error) {
	tag := c.nextTag()

	e := &Encoder{}
	e.Atom(tag).SP().Atom("FETCH").SP().SequenceSet(set)
	options.write(e.SP())

	err := e.WriteTo(c.tpc)
	if err != nil {
		return nil, err
	}

	d := Decoder{}
	cmd := &FetchCommand{}
	var msg Message
	for {
		d.msg, err = c.tpc.ReadLine()
		if err != nil {
			return nil, err
		}

		if d.is(tag) {
			return cmd, d.EndCmd(tag)
		}

		if err = d.expect("*"); err != nil {
			return nil, err
		}

		var num uint32
		num, err = d.SP().Number()
		if err != nil {
			return nil, err
		}
		msg.SeqNumber = num

		err = d.SP().expect("FETCH")
		if err != nil {
			return nil, err
		}

		err = d.SP().List(func() error {
			var key string
			key, err = d.String()
			if err != nil {
				return err
			}

			switch strings.ToUpper(key) {
			case "UID":
				num, err = d.SP().Number()
				if err != nil {
					return err
				}
				msg.UID = num
			case "FLAGS":
				err = d.SP().List(func() error {
					var flag string
					flag, err = d.ReadFlag()
					if err != nil {
						return err
					}
					msg.Flags = append(msg.Flags, Flag(flag))
					return nil
				})
			case "INTERNALDATE":
				msg.InternalDate, err = d.SP().Date()
			case "BODYSTRUCTURE":
				b := BodyStructure{}
				err = d.SP().List(func() error {
					return b.readPart(&d)
				})
				msg.BodyStructure = b
			case "RFC822Size":
				msg.Size, err = d.SP().Number()
			case "BODY":
				body := FetchData{}
				if err = d.expect("["); err != nil {
					return err
				}
				key, err = d.String()
				if err != nil {
					return err
				}
				switch key {
				case "HEADER.FIELDS":
					body.Def.Type = "header"
					err = d.SP().List(func() error {
						var field string
						field, err = d.String()
						body.Def.Fields = append(body.Def.Fields, field)
						return nil
					})
					if err != nil {
						return err
					}
					if err = d.expect("]"); err != nil {
						return err
					}
					if d.SP().is("{") {
						_ = d.expect("{")
						var size uint32
						size, err = d.Number()
						if err != nil {
							return err
						}
						if err = d.expect("}"); err != nil {
							return err
						}
						b := make([]byte, size)
						_, err = c.tpc.R.Read(b)
						if err != nil {
							return err
						}
						body.Data = string(b)
						d.msg, err = c.tpc.ReadLine()
						if err != nil {
							return err
						}
					} else {
						body.Data, err = d.String()
						if err != nil {
							return err
						}
					}
					msg.Body = append(msg.Body, body)
				}
			}

			return err
		})

		if err != nil {
			return nil, err
		}

		cmd.Messages = append(cmd.Messages, msg)
	}
}

func (o *FetchOptions) write(e *Encoder) {
	e.BeginList()
	if o.UID {
		e.ListItem("UID")
	}
	if o.Flags {
		e.ListItem("FLAGS")
	}
	if o.InternalDate {
		e.ListItem("INTERNALDATE")
	}
	if o.BodyStructure {
		e.ListItem("BODYSTRUCTURE")
	}
	if o.Envelope {
		e.ListItem("ENVELOPE")
	}
	if o.RFC822Size {
		e.ListItem("RFC822Size")
	}
	for _, body := range o.Body {
		e.ListItem(body.encode())
	}

	e.EndList()
}

func (s *FetchBodySection) encode() string {
	b := Encoder{}
	b.Atom("BODY")
	if s.Peek {
		b.Atom(".PEEK")
	}
	b.Byte('[')
	switch strings.ToLower(s.Type) {
	case "header":
		b.Atom("HEADER")
		if len(s.Fields) > 0 {
			b.Atom(".FIELDS").SP().BeginList()
			for _, field := range s.Fields {
				b.ListItem(field)
			}
			b.EndList()
		}

	}
	b.Byte(']')
	return b.String()
}

func (b *BodyStructure) readPart(d *Decoder) error {
	if d.is("(") {
		for d.is("(") {
			_ = d.List(func() error {
				part := BodyStructure{}
				err := part.readPart(d)
				if err != nil {
					return err
				}
				b.Parts = append(b.Parts, part)
				return nil
			})
		}
	} else {
		b.Type, _ = d.Quoted()
	}

	b.Subtype, _ = d.SP().Quoted()

	_ = d.SP().NilList(func() error {
		if b.Params == nil {
			b.Params = map[string]string{}
		}

		var k string
		var v string
		k, _ = d.Quoted()
		v, _ = d.SP().Quoted()
		b.Params[k] = v
		return d.Error()
	})

	if len(b.Parts) == 0 {
		_ = d.SP().DiscardValue() // body id
		_ = d.SP().DiscardValue() // body description

		b.Encoding, _ = d.SP().Quoted()
		b.Size, _ = d.SP().Number()
		b.MD5, _ = d.SP().NilString()
	}

	_ = d.SP().NilList(func() error {
		if b.Disposition == nil {
			b.Disposition = map[string]map[string]string{}
		}
		key, _ := d.Quoted()
		m := map[string]string{}
		_ = d.SP().NilList(func() error {
			name, _ := d.Quoted()
			val, _ := d.SP().Quoted()
			m[name] = val
			return nil
		})
		b.Disposition[key] = m
		return nil
	})
	b.Language, _ = d.SP().NilString()
	b.Location, _ = d.SP().NilString()

	for d.IsSP() {
		_ = d.SP().DiscardValue()
	}

	return d.Error()
}

func (s *IdSet) Contains(num uint32) bool {
	for _, set := range s.Ranges {
		if set.Contains(num) {
			return true
		}
	}
	return false
}

func (s *Range) Contains(num uint32) bool {
	if num < s.Start.Value {
		return false
	}
	if s.End.Star {
		return true
	} else {
		return s.Start.Value <= num && s.End.Value >= num
	}
}
