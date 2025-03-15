package mail

import (
	"context"
	"fmt"
	"mokapi/imap"
	"mokapi/smtp"
	"slices"
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

	mb.EnsureInbox()

	// Inbox is a special, mandatory mailbox that is case-insensitive
	if strings.ToLower(mailbox) == "inbox" {
		mailbox = "INBOX"
	}

	f, ok := mb.Folders[mailbox]
	if !ok {
		return nil, fmt.Errorf("mailbox not found")
	}

	c.Session["selected"] = mailbox
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
	mb.EnsureInbox()

	folders := mb.List(ref)

	var list []imap.ListEntry
	for _, folder := range folders {
		for _, f := range folder.List(pattern) {
			list = append(list, imap.ListEntry{
				Delimiter: "/",
				Flags:     f.Flags,
				Name:      f.Name,
			})
		}
	}

	slices.SortFunc(list, func(x, y imap.ListEntry) int {
		// Some clients expect INBOX to appear first, even if sorting alphabetically.
		if strings.ToUpper(x.Name) == "INBOX" {
			return -1
		}
		if strings.ToUpper(y.Name) == "INBOX" {
			return 1
		}
		return strings.Compare(x.Name, y.Name)
	})
	return list, nil
}

func (h *Handler) Store(req *imap.StoreRequest, res imap.FetchResponse, ctx context.Context) error {
	folder, err := getCurrentFolder(ctx)
	if err != nil {
		return err
	}

	do := func(action string, flags []imap.Flag, m *Mail) {
		switch action {
		case "add":
			for _, f := range flags {
				if m.HasFlag(f) {
					continue
				}
				m.Flags = append(m.Flags, f)
			}
		case "remove":
			for _, flag := range flags {
				m.RemoveFlag(flag)
			}
		case "replace":
			m.Flags = flags
		}
	}

	if req.Sequence.IsUid {
		doMessagesByUid(&req.Sequence, folder, func(m *Mail) {
			do(req.Action, req.Flags, m)
			if !req.Silent {
				w := res.NewMessage(m.UId)
				w.WriteFlags(m.Flags...)
			}
		})
	} else {
		doMessagesByMsn(&req.Sequence, folder, func(msn int, m *Mail) {
			do(req.Action, req.Flags, m)
			if !req.Silent {
				w := res.NewMessage(uint32(msn))
				w.WriteFlags(m.Flags...)
			}
		})
	}
	return nil
}

func (h *Handler) Expunge(id *imap.IdSet, w *imap.ExpungeWriter, ctx context.Context) error {
	folder, err := getCurrentFolder(ctx)
	if err != nil {
		return err
	}

	var slice []*Mail
	for i, m := range folder.Messages {
		if !m.HasFlag(imap.FlagDeleted) {
			slice = append(slice, m)
			continue
		}
		if id == nil {
			err = w.Write(uint32(i + 1))
		} else {
			err = w.Write(m.UId)
		}
		if err != nil {
			return err
		}
	}
	folder.Messages = slice
	return nil
}

func (h *Handler) Create(name string, opt *imap.CreateOptions, ctx context.Context) error {
	c := imap.ClientFromContext(ctx)
	mb := c.Session["mailbox"].(*Mailbox)

	var current *Folder
	for _, part := range strings.Split(name, "/") {
		folder := &Folder{Name: part}
		if current != nil {
			current.Folders[part] = folder
			current = folder
		} else {
			mb.Folders[part] = folder
			current = folder
		}
	}
	current.Flags = opt.Flags

	return nil
}

func (h *Handler) Move(set *imap.IdSet, dest string, w *imap.MoveWriter, ctx context.Context) error {
	cCtx := imap.ClientFromContext(ctx)
	mb := cCtx.Session["mailbox"].(*Mailbox)
	s, err := getCurrentFolder(ctx)
	if err != nil {
		return err
	}
	d := getFolder(mb, dest)
	if d == nil {
		return fmt.Errorf("folder not found")
	}

	c := &imap.Copy{UIDValidity: d.uidValidity, SourceUIDs: *set}
	if set.IsUid {
		doMessagesByUid(set, s, func(m *Mail) {
			d.Copy(m)
			c.DestUIDs.Append(imap.IdNum(m.UId))
		})
	} else {
		doMessagesByMsn(set, s, func(msn int, m *Mail) {
			d.Copy(m)
			c.DestUIDs.Append(imap.IdNum(m.UId))
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

func getCurrentFolder(ctx context.Context) (*Folder, error) {
	c := imap.ClientFromContext(ctx)
	mb := c.Session["mailbox"].(*Mailbox)
	selected := c.Session["selected"].(string)
	folder, ok := mb.Folders[selected]
	if !ok {
		return folder, fmt.Errorf("mailbox not found")
	}
	return folder, nil
}

func getFolder(mb *Mailbox, name string) *Folder {
	var current *Folder
	ok := false
	for _, part := range strings.Split(name, "/") {
		if current == nil {
			current, ok = mb.Folders[part]
		} else {
			current, ok = current.Folders[part]
		}
		if !ok {
			return nil
		}
	}
	return current
}
