package imap

type Copy struct {
	UIDValidity uint32
	SourceUIDs  IdSet
	DestUIDs    IdSet
}

type CopyWriter interface {
	WriteCopy(copy *Copy) error
}

type copyWriter struct {
	c   *conn
	tag string
}

func (c *conn) handleCopy(tag string, d *Decoder, useUid bool) error {
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

	w := &copyWriter{c: c, tag: tag}
	err = c.handler.Copy(&set, dest, w, c.ctx)
	if err != nil {
		return err
	}

	return c.writeResponse(tag, &response{
		status: ok,
		text:   "COPY completed",
	})
}

func (w *copyWriter) WriteCopy(copy *Copy) error {
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
