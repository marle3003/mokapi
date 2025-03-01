package mail

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
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

	mb.EnsureInbox()
	f, ok := mb.Folders[mailbox]
	if !ok {
		return nil, fmt.Errorf("mailbox not found")
	}

	firstUnseen := f.FirstUnseen()
	unseen := uint32(0)
	if firstUnseen != nil {
		unseen = firstUnseen.SeqNum
	}

	return &imap.Selected{
		Flags:       []imap.Flag{imap.FlagAnswered, imap.FlagFlagged, imap.FlagDeleted, imap.FlagSeen, imap.FlagDraft},
		NumMessages: uint32(len(f.Messages)),
		NumRecent:   uint32(f.NumRecent()),
		FirstUnseen: unseen,
		UIDValidity: f.uidValidity,
		UIDNext:     f.uidNext,
	}, nil
}

func (h *Handler) Unselect(ctx context.Context) error {
	c := imap.ClientFromContext(ctx)
	c.Session["selected"] = ""
	return nil
}

func (h *Handler) List(ref, pattern string, flags []imap.MailboxFlags, ctx context.Context) ([]imap.ListEntry, error) {
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
	selected := c.Session["selected"].(string)
	f, ok := mb.Folders[selected]
	if !ok {
		return fmt.Errorf("mailbox not found")
	}
	m := f.Messages[0]

	seqNum, ok := c.Session["sequence_number"].(uint32)
	if !ok {
		seqNum = uint32(1)
		c.Session["sequence_number"] = seqNum
	}

	for _, msg := range f.Messages {
		w := res.NewMessage(seqNum)
		w.WriteInternalDate(msg.Time)
		w.WriteRFC822Size(msg.Size())
		w.WriteUID(msg.UId)
		w.WriteFlags()

		body := map[string]string{}
		for _, field := range req.Body.HeaderFields {
			switch field {
			case "date":
				body["date"] = m.Time.Format(imap.DateTimeLayout)
			case "subject":
				body["subject"] = m.Subject
			case "from":

			default:
				log.Warnf("imap header field '%s' not supported", field)
			}
		}
		w.WriteBody(body)

	}

	return nil
}
