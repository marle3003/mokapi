package mail

import (
	"context"
	"fmt"
	"mokapi/imap"
	"mokapi/smtp"
	"strings"
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
	if c.Session["selected"] == mailbox {

	} else {
		c.Session["selected"] = mailbox
	}

	mb.EnsureInbox()
	f, ok := mb.Folders[mailbox]
	if !ok {
		return nil, fmt.Errorf("mailbox not found")
	}

	unseen := f.FirstUnseen()

	return &imap.Selected{
		Flags:       []imap.Flag{imap.FlagAnswered, imap.FlagFlagged, imap.FlagDeleted, imap.FlagSeen, imap.FlagDraft},
		NumMessages: uint32(len(f.Messages)),
		NumRecent:   uint32(f.NumRecent()),
		FirstUnseen: uint32(unseen),
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
			Delimiter: "/",
			Flags:     []imap.MailboxFlags{imap.HasNoChildren},
			Name:      "INBOX",
		},
	}, nil
}

func addressListToString(list []smtp.Address) string {
	var sb strings.Builder
	for _, addr := range list {
		if sb.Len() > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(addressToString(addr))
	}
	return sb.String()
}

func addressToString(addr smtp.Address) string {
	if addr.Name == "" {
		return addr.Address
	}
	return fmt.Sprintf("%s <%s>", addr.Name, addr.Address)
}
