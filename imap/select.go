package imap

import (
	"fmt"
	"strconv"
	"strings"
)

type Selected struct {
	Flags       []Flag
	NumMessages uint32
	NumRecent   uint32
	FirstUnseen uint32
	UIDValidity uint32
	UIDNext     uint32

	conn *conn
	tag  string
}

func (c *conn) handleSelect(tag, param string) error {
	d := Decoder{msg: param}
	mailbox, err := d.String()
	if err != nil {
		return err
	}

	if mailbox == "" {
		return c.writeResponse(tag, &response{
			status: no,
			code:   cannot,
			text:   "Invalid mailbox name: Name is empty",
		})
	}

	if c.state == SelectedState {
		if err := c.handler.Unselect(c.ctx); err != nil {
			return err
		}
		c.state = AuthenticatedState
	}
	if c.state != AuthenticatedState {
		return c.writeResponse(tag, &response{
			status: bad,
			text:   "Command is only valid in authenticated state",
		})
	}

	selected, err := c.handler.Select(mailbox, c.ctx)
	if err != nil {
		return c.writeResponse(tag, &response{
			status: no,
			text:   "No such mailbox, can't access mailbox",
		})
	}
	c.state = SelectedState
	selected.tag = tag
	selected.conn = c

	return selected.write()
}

func (c *conn) handleUnselect(tag string, close bool) error {
	if close {
		// close also performs an implicit expunge but no responses are sent
		if err := c.handler.Expunge(nil, &expungeWriter{}, c.ctx); err != nil {
			return err
		}
	}

	if err := c.handler.Unselect(c.ctx); err != nil {
		return err
	}
	c.state = AuthenticatedState
	return c.writeResponse(tag, &response{
		status: ok,
		text:   "CLOSE completed",
	})
}

func (s *Selected) write() error {
	if err := s.writeExists(); err != nil {
		return err
	}
	if err := s.writeRecent(); err != nil {
		return err
	}
	if err := s.writeUnseen(); err != nil {
		return err
	}
	if err := s.writeUIDValidity(); err != nil {
		return err
	}
	if err := s.writeUIDNext(); err != nil {
		return err
	}
	if err := s.writeFlags(); err != nil {
		return err
	}

	return s.conn.writeResponse(s.tag, &response{
		status: ok,
		code:   readWrite,
		text:   "SELECT completed",
	})
}

func (s *Selected) writeExists() error {
	return s.conn.writeResponse(untagged, &response{
		text: fmt.Sprintf("%v EXISTS", s.NumMessages),
	})
}

func (s *Selected) writeRecent() error {
	return s.conn.writeResponse(untagged, &response{
		text: fmt.Sprintf("%v RECENT", s.NumRecent),
	})
}

func (s *Selected) writeUnseen() error {
	return s.conn.writeResponse(untagged, &response{
		status: ok,
		code:   responseCode(fmt.Sprintf("UNSEEN %v", s.FirstUnseen)),
		text:   fmt.Sprintf("Message %v is first unseen", s.FirstUnseen),
	})
}

func (s *Selected) writeUIDValidity() error {
	return s.conn.writeResponse(untagged, &response{
		status: ok,
		code:   responseCode(fmt.Sprintf("UIDVALIDITY %v", s.UIDValidity)),
		text:   "UIDs valid",
	})
}

func (s *Selected) writeUIDNext() error {
	return s.conn.writeResponse(untagged, &response{
		status: ok,
		code:   responseCode(fmt.Sprintf("UIDNEXT %v", s.UIDNext)),
		text:   "Predicted next UID",
	})
}

func (s *Selected) writeFlags() error {
	return s.conn.writeResponse(untagged, &response{
		text: fmt.Sprintf("FLAGS (%s)", flagsToString(s.Flags)),
	})
}

func (c *Client) Select(folder string) (Selected, error) {
	tag := c.nextTag()
	err := c.tpc.PrintfLine("%s SELECT %s", tag, folder)
	if err != nil {
		return Selected{}, err
	}

	d := &Decoder{}
	sel := Selected{}
	for {
		d.msg, err = c.tpc.ReadLine()
		if err != nil {
			return sel, err
		}

		if d.is(tag) {
			return sel, d.EndCmd(tag)
		}

		if err = d.expect("*"); err != nil {
			return sel, err
		}

		var str string
		str, err = d.SP().String()
		if err != nil {
			return sel, err
		}
		if isNum(str) {
			var num uint32
			num, err = toNum(str)
			if err != nil {
				return sel, err
			}

			var key string
			key, err = d.SP().String()
			if err != nil {
				return sel, err
			}
			switch strings.ToUpper(key) {
			case "EXISTS":
				sel.NumMessages = num
			case "RECENT":
				sel.NumRecent = num
			}
		} else {
			switch strings.ToUpper(str) {
			case "FLAGS":
				err = d.SP().List(func() error {
					var flag string
					flag, err = d.ReadFlag()
					if err != nil {
						return err
					}
					sel.Flags = append(sel.Flags, Flag(flag))
					return nil
				})
			case "OK":
				err = d.SP().expect("[")
				if err != nil {
					return sel, err
				}
				var key string
				key, err = d.Read(isAtom)
				if err != nil {
					return sel, err
				}
				switch strings.ToUpper(key) {
				case "UNSEEN":
					var num uint32
					num, err = d.SP().Number()
					if err != nil {
						return sel, err
					}
					sel.FirstUnseen = num
				case "UIDVALIDITY":
					var num uint32
					num, err = d.SP().Number()
					if err != nil {
						return sel, err
					}
					sel.UIDValidity = num
				case "UIDNEXT":
					var num uint32
					num, err = d.SP().Number()
					if err != nil {
						return sel, err
					}
					sel.UIDNext = num
				}
			}

			if err != nil {
				return sel, err
			}
		}
	}
}

func (c *Client) Close() error {
	tag := c.nextTag()
	err := c.tpc.PrintfLine("%s CLOSE", tag)
	if err != nil {
		return err
	}

	d := &Decoder{}
	d.msg, err = c.tpc.ReadLine()
	if err != nil {
		return err
	}
	return d.EndCmd(tag)
}

func isNum(s string) bool {
	return s[0] >= '0' && s[0] <= '9'
}

func toNum(s string) (uint32, error) {
	v, err := strconv.ParseUint(s, 0, 32)
	if err != nil {
		return 0, err
	}
	return uint32(v), nil
}
