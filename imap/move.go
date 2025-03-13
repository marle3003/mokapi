package imap

type MoveWriter struct {
	c   *conn
	tag string
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

	w := &MoveWriter{c: c, tag: tag}
	err = c.handler.Move(&set, dest, w, c.ctx)
	if err != nil {
		return err
	}

	return c.writeResponse(tag, &response{
		status: ok,
		text:   "MOVE completed",
	})
}

func (w *MoveWriter) WriteCopy(copy *Copy) error {
	e := Encoder{}

	e.Byte('[')
	e.Atom("COPYUID")
	e.SP().Number(copy.UIDValidity)
	e.SP().SequenceSet(copy.SourceUIDs)
	e.SP().SequenceSet(copy.DestUIDs)
	e.Byte(']')
	e.SP().Atom("COPY")

	return w.c.writeResponse(untagged, &response{
		text: e.String(),
	})
}

func (w *MoveWriter) WriteExpunge(id uint32) error {
	return w.c.writeExpunge(id)
}
