package imap

import (
	"io"
	"net/textproto"
)

type UpdateWriter interface {
	WriteNumMessages(n uint32) error
	WriteMessageFlags(msn uint32, flags []Flag) error
	WriteExpunge(msn uint32) error
}

func (c *conn) handleIdle(tag string) error {
	if c.state != SelectedState {
		return c.writeResponse(tag, &response{
			status: bad,
			text:   "Command is only valid in selected state",
		})
	}

	done := make(chan struct{})
	err := c.handler.Idle(&idleWriter{}, done, c.ctx)
	if err != nil {
		return err
	}

	err = c.tpc.PrintfLine("+ idling")
	if err != nil {
		return err
	}

	line, err := c.tpc.ReadLine()
	close(done)
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}

	_, cmd, _ := parseLine(line)
	if cmd != "DONE" {
		return c.writeResponse(tag, &response{
			status: bad,
			text:   "Expected DONE to end IDLE",
		})
	}

	return c.writeResponse(tag, &response{
		status: ok,
		text:   "IDLE terminated",
	})
}

type idleWriter struct {
	tpc *textproto.Conn
}

func (w *idleWriter) WriteNumMessages(n uint32) error {
	return w.tpc.PrintfLine("* %v EXISTS", n)
}

func (w *idleWriter) WriteMessageFlags(msn uint32, flags []Flag) error {
	return w.tpc.PrintfLine("* %v FETCH (%v)", msn, joinFlags(flags))
}

func (w *idleWriter) WriteExpunge(msn uint32) error {
	return w.tpc.PrintfLine("%v EXPUNGE", msn)
}
