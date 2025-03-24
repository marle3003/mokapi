package imap

type MoveWriter interface {
	CopyWriter
	WriteExpunge(id uint32) error
}

type moveWriter struct {
	copyWriter
}

func (c *conn) handleMove(tag string, d *Decoder, useUid bool) error {
	if c.state != AuthenticatedState && c.state != SelectedState {
		return c.writeResponse(tag, &response{
			status: bad,
			text:   "Command is only valid in authenticated state",
		})
	}

	set, err := d.Sequence()
	if err != nil {
		return err
	}
	set.IsUid = useUid

	var dest string
	dest, err = d.SP().String()

	w := &moveWriter{copyWriter: copyWriter{
		c:   c,
		tag: tag,
	}}
	err = c.handler.Move(&set, dest, w, c.ctx)
	if err != nil {
		return err
	}

	return c.writeResponse(tag, &response{
		status: ok,
		text:   "MOVE completed",
	})
}

func (w *moveWriter) WriteExpunge(id uint32) error {
	return w.c.writeExpunge(id)
}
