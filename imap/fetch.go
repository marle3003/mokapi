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
}

type SequenceSet []Sequence

type Sequence struct {
	Start uint32
	End   uint32
}

type FetchBody struct {
	Section      []int
	HeaderFields []string
}

type FetchRequest struct {
	Sequence SequenceSet
	Options  FetchOptions
	// nil means everything
	Body *FetchBody
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

	if err := c.writeFetchResponse(&res); err != nil {
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
}

func (c *Client) Fetch(id int, options FetchOptions) (*FetchCommand, error) {
	tag := c.nextTag()
	err := c.tpc.PrintfLine("%s FETCH %v (%s)", tag, id, options.list())
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
			}

			return err
		})

		if err != nil {
			return nil, err
		}

		cmd.Messages = append(cmd.Messages, msg)
	}
}

func (o *FetchOptions) list() string {
	var r []string
	if o.UID {
		r = append(r, "UID")
	}
	if o.Flags {
		r = append(r, "FLAGS")
	}
	if o.InternalDate {
		r = append(r, "INTERNALDATE")
	}
	if o.BodyStructure {
		r = append(r, "BODYSTRUCTURE")
	}
	if o.Envelope {
		r = append(r, "ENVELOPE")
	}
	if o.RFC822Size {
		r = append(r, "RFC822Size")
	}
	return strings.Join(r, " ")
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
