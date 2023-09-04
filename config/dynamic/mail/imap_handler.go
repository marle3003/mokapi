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

	firstUnseen := mb.FirstUnseen()
	unseen := uint32(0)
	if firstUnseen != nil {
		unseen = firstUnseen.SeqNum
	}

	return &imap.Selected{
		Flags:       []imap.Flag{imap.FlagAnswered, imap.FlagFlagged, imap.FlagDeleted, imap.FlagSeen, imap.FlagDraft},
		NumMessages: uint32(len(mb.Messages)),
		NumRecent:   uint32(mb.NumRecent()),
		FirstUnseen: unseen,
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
	m := mb.Messages[0]

	w := res.NewMessage(1)
	w.WriteInternalDate(m.Time)
	w.WriteRFC822Size(m.Size())
	w.WriteUID(1)
	var values []string
	for _, field := range req.Body.HeaderFields {
		switch field {
		case "date":
			values = append(values, m.Time.Format(imap.DateTimeLayout))
		case "subject":
			values = append(values, m.Subject)
		//case "from":
		//	values = append(values, fmt.Sprintf("%s <%s>", m.From))
		case "to":
		case "cc":
		case "message-id":
		case "in-reply-to":
		default:
			continue
		}
	}

	return nil
}
