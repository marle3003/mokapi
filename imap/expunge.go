package imap

import "fmt"

type ExpungeWriter struct {
	c *conn
}

func (c *conn) handleExpunge(tag string, dec *Decoder, useUid bool) error {
	if c.state != AuthenticatedState && c.state != SelectedState {
		return c.writeResponse(tag, &response{
			status: bad,
			text:   "Command is only valid in authenticated state",
		})
	}

	var set *IdSet
	if useUid {
		s, err := dec.Sequence()
		if err != nil {
			return err
		}
		set = &s
	}

	w := &ExpungeWriter{c: c}
	if err := c.handler.Expunge(set, w, c.ctx); err != nil {
		return err
	}
	return c.writeResponse(tag, &response{
		status: ok,
		text:   "EXPUNGE completed",
	})
}

func (w *ExpungeWriter) Write(id uint32) error {
	return w.c.writeExpunge(id)
}

func (c *conn) writeExpunge(id uint32) error {
	if c == nil {
		return nil
	}

	return c.writeResponse(untagged, &response{
		text: fmt.Sprintf("%v EXPUNGE", id),
	})
}
