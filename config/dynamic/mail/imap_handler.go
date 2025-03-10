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
	for _, m := range h.MailStore.Mailboxes {
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
	c := imap.ClientFromContext(ctx)
	mb := c.Session["mailbox"].(*Mailbox)

	var list []imap.ListEntry
	for _, f := range mb.Folders {
		list = append(list, imap.ListEntry{
			Delimiter: "/",
			Flags:     f.Flags,
			Name:      f.Name,
		})
	}

	return list, nil
}

func (h *Handler) Store(req *imap.StoreRequest, res imap.FetchResponse, ctx context.Context) error {
	c := imap.ClientFromContext(ctx)
	mb := c.Session["mailbox"].(*Mailbox)
	selected := c.Session["selected"].(string)
	f, ok := mb.Folders[selected]
	if !ok {
		return fmt.Errorf("mailbox not found")
	}

	do := func(action string, flags []imap.Flag, m *Mail) {
		switch action {
		case "add":
			m.Flags = append(m.Flags, flags...)
		case "remove":
			for _, flag := range flags {
				m.RemoveFlag(flag)
			}
		case "replace":
			m.Flags = flags
		}
	}

	if req.Sequence.IsUid {
		doMessagesByUid(&req.Sequence, f, func(m *Mail) {
			do(req.Action, req.Flags, m)
			if !req.Silent {
				w := res.NewMessage(m.UId)
				w.WriteFlags(m.Flags...)
			}
		})
	} else {
		doMessagesByMsn(&req.Sequence, f, func(msn int, m *Mail) {
			do(req.Action, req.Flags, m)
			if !req.Silent {
				w := res.NewMessage(uint32(msn))
				w.WriteFlags(m.Flags...)
			}
		})
	}
	return nil
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
