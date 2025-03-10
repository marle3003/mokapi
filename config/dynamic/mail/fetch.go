package mail

import (
	"context"
	"fmt"
	"mokapi/imap"
	"strings"
)

func (h *Handler) Fetch(req *imap.FetchRequest, res imap.FetchResponse, ctx context.Context) error {
	c := imap.ClientFromContext(ctx)
	mb := c.Session["mailbox"].(*Mailbox)
	selected := c.Session["selected"].(string)
	f, ok := mb.Folders[selected]
	if !ok {
		return fmt.Errorf("mailbox not found")
	}

	if req.Sequence.IsUid {
		doMessagesByUid(&req.Sequence, f, func(m *Mail) {
			w := res.NewMessage(m.UId)
			writeMessage(m, req.Options, w)
		})
	} else {
		doMessagesByMsn(&req.Sequence, f, func(msn int, m *Mail) {
			w := res.NewMessage(uint32(msn))
			writeMessage(m, req.Options, w)
		})
	}
	return nil
}

func doMessagesByUid(set *imap.IdSet, folder *Folder, action func(m *Mail)) {
	for _, msg := range folder.Messages {
		if set.Contains(msg.UId) {
			action(msg)
		}
	}
}

func doMessagesByMsn(set *imap.IdSet, folder *Folder, action func(msn int, m *Mail)) {
	for _, r := range set.Ranges {
		start := 0
		end := int(r.End.Value)
		if r.Start.Value > 0 {
			start = int(r.Start.Value) - 1
		}
		if r.End.Star {
			end = len(folder.Messages)
		}

		for i, msg := range folder.Messages[start:end] {
			action(i+1, msg)
		}
	}
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
		bw := w.WriteBody(body)
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
					if v, ok := msg.Headers[field]; ok {
						bw.WriteHeader(field, v)
					}
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
