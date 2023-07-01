package imap

import "fmt"

type Selected struct {
	Flags       []Flag
	NumMessages uint32
	NumRecent   uint32
	FirstUnseen uint32
	UIDValidity uint32
	UIDNext     uint32
}

func (c *conn) canSelect() bool {
	return c.state == AuthenticatedState
}

func (c *conn) handleSelect(tag, mailbox string) error {
	selected := c.handler.Select(mailbox)

	if err := c.writeExists(selected.NumMessages); err != nil {
		return err
	}
	if err := c.writeRecent(selected.NumRecent); err != nil {
		return err
	}
	if err := c.writeUnseen(selected.FirstUnseen); err != nil {
		return err
	}
	if err := c.writeUIDValidity(selected.UIDValidity); err != nil {
		return err
	}
	if err := c.writeUIDNext(selected.UIDNext); err != nil {
		return err
	}
	if err := c.writeFlags(selected.Flags); err != nil {
		return err
	}

	return c.writeResponse(tag, &response{
		status: ok,
		code:   readWrite,
		text:   "SELECT completed",
	})
}

func (c *conn) writeExists(exists uint32) error {
	return c.writeResponse(untagged, &response{
		text: fmt.Sprintf("%v EXISTS", exists),
	})
}

func (c *conn) writeRecent(recent uint32) error {
	return c.writeResponse(untagged, &response{
		text: fmt.Sprintf("%v RECENT", recent),
	})
}

func (c *conn) writeUnseen(firstUnseen uint32) error {
	return c.writeResponse(untagged, &response{
		status: ok,
		code:   responseCode(fmt.Sprintf("UNSEEN %v", firstUnseen)),
		text:   fmt.Sprintf("Message %v is first unseen", firstUnseen),
	})
}

func (c *conn) writeUIDValidity(v uint32) error {
	return c.writeResponse(untagged, &response{
		status: ok,
		code:   responseCode(fmt.Sprintf("UIDVALIDITY %v", v)),
		text:   "UIDs valid",
	})
}

func (c *conn) writeUIDNext(v uint32) error {
	return c.writeResponse(untagged, &response{
		status: ok,
		code:   responseCode(fmt.Sprintf("UIDNEXT %v", v)),
		text:   "Predicted next UID",
	})
}

func (c *conn) writeFlags(flags []Flag) error {
	return c.writeResponse(untagged, &response{
		text: fmt.Sprintf("FLAGS (%s)", flagsToString(flags)),
	})
}
