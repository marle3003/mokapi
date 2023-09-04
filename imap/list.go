package imap

import (
	"fmt"
	"strings"
)

type ListEntry struct {
	Flags []MailboxFlags
	Name  string
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
	list, err := c.handler.List(ref, pattern, c.ctx)
	if err != nil {
		return err
	}

	if err := c.writeList(list); err != nil {
		return err
	}

	return c.writeResponse(tag, &response{
		status: ok,
		text:   "List completed",
	})
}

func (c *conn) writeList(list []ListEntry) error {
	for _, entry := range list {
		if err := c.writeListEntry(entry); err != nil {
			return err
		}
	}
	return nil
}

func (c *conn) writeListEntry(entry ListEntry) error {
	return c.writeResponse(untagged, &response{
		text: fmt.Sprintf("LIST () NIL %v", entry.Name),
	})
}
