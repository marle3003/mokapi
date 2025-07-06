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

func (h *Handler) Select(mailbox string, readonly bool, ctx context.Context) (*imap.Selected, error) {
	mb, _ := getContext(ctx)
	mb.m.Lock()
	defer mb.m.Unlock()

	f := mb.Select(mailbox)
	if f == nil {
		return nil, fmt.Errorf("mailbox not found")
	}

	c := imap.ClientFromContext(ctx)
	c.Session["selected"] = mailbox
	c.Session["readonly"] = readonly
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
	mb, _ := getContext(ctx)
	mb.m.Lock()
	defer mb.m.Unlock()

	folders := mb.List(ref)

	var list []imap.ListEntry
	for _, folder := range folders {
		for _, f := range folder.List(pattern, flags) {
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
	mb, folder := getContext(ctx)
	if folder == nil {
		return fmt.Errorf("folder not found")
	}
	mb.m.Lock()
	defer mb.m.Unlock()

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
				w.WriteUID(m.UId)
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

func (h *Handler) Expunge(id *imap.IdSet, w imap.ExpungeWriter, ctx context.Context) error {
	mb, folder := getContext(ctx)
	if folder == nil {
		return fmt.Errorf("folder not found")
	}
	mb.m.Lock()
	defer mb.m.Unlock()

	var slice []*Mail
	for i, m := range folder.Messages {
		if !m.HasFlag(imap.FlagDeleted) {
			slice = append(slice, m)
			continue
		}

		msn := uint32(i + 1)
		var err error
		if id == nil {
			err = w.Write(msn)
		} else {
			if id.IsUid && id.Contains(m.UId) {
				err = w.Write(m.UId)
			} else if id.Contains(msn) {
				err = w.Write(msn)
			} else {
				slice = append(slice, m)
			}
		}
		if err != nil {
			return err
		}
	}
	folder.Messages = slice
	return nil
}

func (h *Handler) Create(name string, opt *imap.CreateOptions, ctx context.Context) error {
	mb, _ := getContext(ctx)
	mb.m.Lock()
	defer mb.m.Unlock()

	var current *Folder
	for _, part := range strings.Split(name, "/") {
		if strings.ToUpper(part) == "INBOX" {
			current = mb.Folders["INBOX"]
			continue
		}
		folder := &Folder{Name: part}
		if current != nil {
			current.AddFolder(folder)
			current = folder
		} else {
			mb.Folders[part] = folder
			current = folder
		}
	}
	current.Flags = opt.Flags

	return nil
}

func (h *Handler) Delete(mailbox string, ctx context.Context) error {
	mb, _ := getContext(ctx)
	mb.m.Lock()
	defer mb.m.Unlock()

	return mb.DeleteFolder(mailbox)
}

func (h *Handler) Rename(existingName, newName string, ctx context.Context) error {
	mb, _ := getContext(ctx)
	mb.m.Lock()
	defer mb.m.Unlock()

	return mb.RenameFolder(existingName, newName)
}

func (h *Handler) Copy(set *imap.IdSet, dest string, w imap.CopyWriter, ctx context.Context) error {
	mb, source := getContext(ctx)
	if source == nil {
		return fmt.Errorf("folder not found")
	}
	mb.m.Lock()
	defer mb.m.Unlock()

	d := mb.Select(dest)
	if d == nil {
		return fmt.Errorf("folder not found")
	}

	c := &imap.Copy{UIDValidity: d.UidValidity(), SourceUIDs: *set}
	if set.IsUid {
		c.DestUIDs.IsUid = true
		doMessagesByUid(set, source, func(m *Mail) {
			d.Copy(m)
			c.DestUIDs.Append(imap.IdNum(m.UId))
		})
	} else {
		doMessagesByMsn(set, source, func(msn int, m *Mail) {
			d.Copy(m)
			c.DestUIDs.Append(imap.IdNum(msn))
		})
	}

	if err := w.WriteCopy(c); err != nil {
		return err
	}

	return nil
}

func (h *Handler) Move(set *imap.IdSet, dest string, w imap.MoveWriter, ctx context.Context) error {
	mb, source := getContext(ctx)
	if source == nil {
		return fmt.Errorf("folder not found")
	}
	mb.m.Lock()
	defer mb.m.Unlock()

	d := mb.Select(dest)
	if d == nil {
		return fmt.Errorf("folder not found")
	}

	c := &imap.Copy{UIDValidity: d.UidValidity(), SourceUIDs: *set}
	if set.IsUid {
		c.DestUIDs.IsUid = true
		doMessagesByUid(set, source, func(m *Mail) {
			source.Remove(m)
			m = d.Copy(m)
			c.DestUIDs.Append(imap.IdNum(m.UId))
			w.WriteExpunge(m.UId)
		})
	} else {
		doMessagesByMsn(set, source, func(msn int, m *Mail) {
			source.Remove(m)
			m = d.Copy(m)
			c.DestUIDs.Append(imap.IdNum(msn))
			w.WriteExpunge(uint32(msn))
		})
	}

	if err := w.WriteCopy(c); err != nil {
		return err
	}

	return nil
}

func (h *Handler) Status(req *imap.StatusRequest, ctx context.Context) (imap.StatusResult, error) {
	mb, _ := getContext(ctx)
	mb.m.Lock()
	defer mb.m.Unlock()

	folder := mb.Select(req.Mailbox)
	if folder == nil {
		return imap.StatusResult{}, fmt.Errorf("folder not found")
	}
	return folder.Status(), nil
}

func (h *Handler) Subscribe(mailbox string, ctx context.Context) error {
	mb, _ := getContext(ctx)
	mb.m.Lock()
	defer mb.m.Unlock()

	f := mb.Select(mailbox)
	if f == nil {
		return fmt.Errorf("folder not found")
	}
	f.Flags = append(f.Flags, imap.Subscribed)
	return nil
}

func (h *Handler) Unsubscribe(mailbox string, ctx context.Context) error {
	mb, _ := getContext(ctx)
	mb.m.Lock()
	defer mb.m.Unlock()

	f := mb.Select(mailbox)
	if f == nil {
		return fmt.Errorf("folder not found")
	}

	f.RemoveFlag(imap.Subscribed)
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

func getContext(ctx context.Context) (*Mailbox, *Folder) {
	c := imap.ClientFromContext(ctx)
	mb := c.Session["mailbox"].(*Mailbox)
	mb.EnsureInbox()
	selected, ok := c.Session["selected"]
	if !ok {
		return mb, nil
	}
	folder := mb.Select(selected.(string))
	return mb, folder
}
