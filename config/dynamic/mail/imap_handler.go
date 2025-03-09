package mail

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
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

func (h *Handler) UidFetch(req *imap.FetchRequest, res imap.FetchResponse, ctx context.Context) error {
	c := imap.ClientFromContext(ctx)
	mb := c.Session["mailbox"].(*Mailbox)
	selected := c.Session["selected"].(string)
	f, ok := mb.Folders[selected]
	if !ok {
		return fmt.Errorf("mailbox not found")
	}

	for _, msg := range f.Messages {
		if req.Sequence.Contains(msg.UId) {
			w := res.NewMessage(msg.UId)
			writeMessage(msg, req.Options, w)
		}
	}

	return nil
}

func (h *Handler) Fetch(req *imap.FetchRequest, res imap.FetchResponse, ctx context.Context) error {
	c := imap.ClientFromContext(ctx)
	mb := c.Session["mailbox"].(*Mailbox)
	selected := c.Session["selected"].(string)
	f, ok := mb.Folders[selected]
	if !ok {
		return fmt.Errorf("mailbox not found")
	}

	for _, r := range req.Sequence.Ranges {
		start := 0
		end := int(r.End.Value)
		if r.Start.Value > 0 {
			start = int(r.Start.Value) - 1
		}
		if r.End.Star {
			end = len(f.Messages)
		}

		for i, msg := range f.Messages[start:end] {
			w := res.NewMessage(uint32(i + 1))
			writeMessage(msg, req.Options, w)
		}
	}

	return nil
}

func writeMessage(msg *Mail, opt imap.FetchOptions, w imap.MessageWriter) {
	if opt.UID {
		w.WriteUID(msg.UId)
	}
	if opt.InternalDate {
		w.WriteInternalDate(msg.Time)
	}
	if opt.RFC822Size {
		w.WriteRFC822Size(uint32(msg.Size))
	}
	if opt.Flags {
		w.WriteFlags(msg.Flags...)
	}

	for _, body := range opt.Body {
		bw := w.WriteBody2(body)
		if body.Type == "header" {
			for _, field := range body.Fields {
				switch strings.ToLower(field) {
				case "date":
					bw.WriteHeader("date", msg.Time.Format(imap.DateTimeLayout))
				case "subject":
					bw.WriteHeader("subject", msg.Subject)
				case "from":
					bw.WriteHeader("from", addressListToString(msg.From))
				case "to":
					bw.WriteHeader("to", addressListToString(msg.To))
				case "cc":
					if msg.Cc != nil {
						bw.WriteHeader("cc", addressListToString(msg.Cc))
					}
				case "message-id":
					bw.WriteHeader("message-id", msg.MessageId)
				case "content-type":
					bw.WriteHeader("content-type", msg.ContentType)
				default:
					log.Warnf("imap header field '%s' not supported", field)
				}
			}
		} else if body.Type == "text" {
			bw.WriteBody(msg.Body)
		} else if body.Type == "" {
			for k, v := range msg.Headers {
				bw.WriteHeader(k, v)
			}
			bw.WriteBody(msg.Body)
		}
		bw.Close()
	}
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
