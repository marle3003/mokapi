package mail

import (
	"context"
	"fmt"
	"mokapi/imap"
)

func (h *Handler) Login(username, password string, ctx context.Context) error {
	c := imap.ClientFromContext(ctx)
	for _, m := range h.Store.Mailboxes {
		if m.Username == username && m.Password == password {
			c.Session["mailbox"] = m
			return nil
		}
	}
	return fmt.Errorf("invalid credentials")
}

func (h *Handler) Select(mailbox string, ctx context.Context) (*imap.Selected, error) {
	c := imap.ClientFromContext(ctx)
	mb := c.Session["mailbox"].(*Mailbox)
	c.Session["selected"] = mailbox
	return &imap.Selected{
		Flags:       []imap.Flag{imap.FlagSeen},
		NumMessages: uint32(len(mb.Messages)),
		NumRecent:   0,
		FirstUnseen: 1,
		UIDValidity: mb.uidValidity,
		UIDNext:     mb.messageSequenceNumber,
	}, nil
}

func (h *Handler) Unselect(ctx context.Context) error {
	c := imap.ClientFromContext(ctx)
	c.Session["selected"] = ""
	return nil
}

func (h *Handler) List(ref, pattern string, ctx context.Context) ([]imap.ListEntry, error) {
	return []imap.ListEntry{
		{
			Flags: []imap.MailboxFlags{imap.UnMarked},
			Name:  "INBOX",
		},
	}, nil
}

func (h *Handler) Fetch(req *imap.FetchRequest, res imap.FetchResponse, ctx context.Context) error {
	c := imap.ClientFromContext(ctx)
	mb := c.Session["mailbox"].(*Mailbox)
	msg := mb.Messages[0]
	w := res.NewMessage(1)
	w.WriteInternalDate(msg.Time)
	w.WriteRFC822Size(msg.Size())
	return nil
}
