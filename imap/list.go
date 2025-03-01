package imap

import (
	"fmt"
	"strings"
)

type ListEntry struct {
	Flags     []MailboxFlags
	Delimiter string
	Name      string
}

func (c *conn) handleList(tag, param string) error {
	if c.state != AuthenticatedState && c.state != SelectedState {
		return c.writeResponse(tag, &response{
			status: bad,
			text:   "Command is only valid in authenticated state",
		})
	}

	args := strings.SplitN(param, " ", 2)
	ref := args[0]
	pattern := args[1]
	list, err := c.handler.List(ref, pattern, nil, c.ctx)
	if err != nil {
		return err
	}

	w := listWriter{
		conn: c,
		tag:  tag,
		list: list,
		cmd:  "LIST",
	}
	return w.write()
}

func (c *conn) handleLSub(tag, param string) error {
	if c.state != AuthenticatedState && c.state != SelectedState {
		return c.writeResponse(tag, &response{
			status: bad,
			text:   "Command is only valid in authenticated state",
		})
	}

	args := strings.SplitN(param, " ", 2)
	ref := args[0]
	pattern := args[1]
	list, err := c.handler.List(ref, pattern, []MailboxFlags{Subscribed}, c.ctx)
	if err != nil {
		return err
	}

	w := listWriter{
		conn: c,
		tag:  tag,
		list: list,
		cmd:  "LSUB",
	}
	return w.write()
}

type listWriter struct {
	conn *conn
	tag  string
	list []ListEntry
	cmd  string
}

func (w *listWriter) write() error {
	for _, entry := range w.list {
		if err := w.writeListEntry(entry); err != nil {
			return err
		}
	}
	return w.conn.writeResponse(w.tag, &response{
		status: ok,
		text:   fmt.Sprintf("%s completed", w.cmd),
	})
}

func (w *listWriter) writeListEntry(entry ListEntry) error {
	return w.conn.writeResponse(untagged, &response{
		text: fmt.Sprintf(`%v (%v) "%v" %v`, w.cmd, joinMailboxFlags(entry.Flags), entry.Delimiter, entry.Name),
	})
}

func (c *Client) List(ref, pattern string) ([]ListEntry, error) {
	tag := c.nextTag()
	if ref == "" {
		ref = `""`
	}
	err := c.tpc.PrintfLine("%s LIST %s %s", tag, ref, pattern)
	if err != nil {
		return nil, err
	}
	list := []ListEntry{}
	d := &Decoder{}
	for {
		d.msg, err = c.tpc.ReadLine()
		if err != nil {
			return nil, err
		}

		if d.is(tag) {
			return list, d.EndCmd(tag)
		}

		if err = d.expect("*"); err != nil {
			return nil, err
		}

		if err = d.SP().expect("LIST"); err != nil {
			return nil, err
		}

		e := ListEntry{}

		if err = d.SP().List(func() error {
			var flag string
			flag, err = d.ReadFlag()
			if err != nil {
				return err
			}
			e.Flags = append(e.Flags, MailboxFlags(flag))
			return nil
		}); err != nil {
			return nil, err
		}

		e.Delimiter, err = d.SP().Quoted()
		if err != nil {
			return nil, err
		}

		var name string
		name, err = d.SP().String()
		if err != nil {
			return nil, err
		}
		e.Name, err = DecodeUTF7(name)
		if err != nil {
			return nil, err
		}

		list = append(list, e)
	}
}

func (c *Client) LSub(ref, pattern string) ([]ListEntry, error) {
	tag := c.nextTag()
	if ref == "" {
		ref = `""`
	}
	err := c.tpc.PrintfLine("%s LSUB %s %s", tag, ref, pattern)
	if err != nil {
		return nil, err
	}
	list := []ListEntry{}
	d := &Decoder{}

	for {
		d.msg, err = c.tpc.ReadLine()
		if err != nil {
			return nil, err
		}

		if d.is(tag) {
			return list, d.EndCmd(tag)
		}

		if err = d.expect("*"); err != nil {
			return nil, err
		}

		if err = d.SP().expect("LSUB"); err != nil {
			return nil, err
		}

		e := ListEntry{}

		if err = d.SP().List(func() error {
			var flag string
			flag, err = d.ReadFlag()
			if err != nil {
				return err
			}
			e.Flags = append(e.Flags, MailboxFlags(flag))
			return nil
		}); err != nil {
			return nil, err
		}

		e.Delimiter, err = d.SP().Quoted()
		if err != nil {
			return nil, err
		}

		var name string
		name, err = d.SP().String()
		if err != nil {
			return nil, err
		}
		e.Name, err = DecodeUTF7(name)
		if err != nil {
			return nil, err
		}

		list = append(list, e)
	}
}
