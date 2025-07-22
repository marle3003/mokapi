package mail

import (
	"context"
	"fmt"
	"mokapi/imap"
	"mokapi/media"
	"mokapi/smtp"
	"strings"
)

func (h *Handler) Fetch(req *imap.FetchRequest, res imap.FetchResponse, ctx context.Context) error {
	c := imap.ClientFromContext(ctx)
	mb := c.Session["mailbox"].(*Mailbox)
	selected := c.Session["selected"].(string)
	f := mb.Select(selected)
	if f == nil {
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
	for i, msg := range folder.Messages {
		msn := i + 1
		if set.Contains(uint32(msn)) {
			action(msn, msg)
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
	if opt.Envelope {
		env := &imap.Envelope{
			Date:      msg.Time,
			Subject:   msg.Subject,
			From:      toEnvelopeAddressList(msg.From),
			ReplyTo:   toEnvelopeAddressList(msg.ReplyTo),
			To:        toEnvelopeAddressList(msg.To),
			Cc:        toEnvelopeAddressList(msg.Cc),
			Bcc:       toEnvelopeAddressList(msg.Bcc),
			InReplyTo: msg.InReplyTo,
			MessageId: msg.MessageId,
		}

		sender := toEnvelopeAddress(msg.Sender)
		if sender != nil {
			env.Sender = []imap.Address{*sender}
		}

		w.WriteEnvelope(env)
	}
	if opt.BodyStructure {
		ct := media.ParseContentType(msg.ContentType)

		bs := &imap.BodyStructure{
			Type:     ct.Type,
			Subtype:  ct.Subtype,
			Params:   ct.Parameters,
			Encoding: msg.ContentTransferEncoding,
			Size:     uint32(msg.Size),
		}

		for _, part := range msg.Attachments {
			partType := media.ParseContentType(part.ContentType)
			p := imap.BodyStructure{
				Type:        partType.Type,
				Subtype:     partType.Subtype,
				Params:      partType.Parameters,
				Encoding:    part.Header["Content-Transfer-Encoding"],
				Size:        uint32(len(part.Data)),
				Disposition: part.Disposition,
			}
			if part.ContentId != "" {
				p.ContentId = &part.ContentId
			}
			if part.Disposition != "" && part.Disposition != "inline" {
				p.Disposition = part.Disposition
			}
			bs.Parts = append(bs.Parts, p)

		}

		w.WriteBodyStructure(bs)
	}

	for _, body := range opt.Body {
		bw := w.WriteBody(body)

		if body.Specifier == "header" {
			if len(body.Fields) == 0 {
				bw.WriteHeader("date", msg.Time.Format(imap.DateTimeLayout))
				bw.WriteHeader("subject", msg.Subject)
				bw.WriteHeader("from", addressListToString(msg.From))
				bw.WriteHeader("to", addressListToString(msg.To))
				if msg.Cc != nil {
					bw.WriteHeader("cc", addressListToString(msg.Cc))
				}
				bw.WriteHeader("message-id", msg.MessageId)
				bw.WriteHeader("content-type", msg.ContentType)
				// Add any additional headers
				for name, value := range msg.Headers {
					// avoid writing duplicates
					if !isStandardHeader(name) {
						bw.WriteHeader(name, value)
					}
				}
			} else {
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
			}
		} else if body.Specifier == "text" {
			if strings.HasPrefix(msg.ContentType, "multipart/") && len(body.Parts) == 0 {
				// https://datatracker.ietf.org/doc/html/rfc3501#section-7.4.2
				continue
			}
			if len(body.Parts) == 0 || len(msg.Attachments) == 0 {
				bw.WriteBody(msg.Body)
			} else {
				for _, part := range body.Parts {
					index := part - 1
					if index < 0 || index > len(msg.Attachments) {
						continue
					}
					p := msg.Attachments[index]
					bw.WriteBody(string(p.Data))
				}
			}
		} else if body.Specifier == "" {
			if body.Parts == nil {
				for k, v := range msg.Headers {
					bw.WriteHeader(k, v)
				}
				bw.WriteBody(msg.Body)
			} else {
				index := body.Parts[0] - 1
				if index < 0 || index > len(msg.Attachments) {
					continue
				}
				att := msg.Attachments[index]
				for k, v := range att.Header {
					bw.WriteHeader(k, v)
				}
				bw.WriteBody(string(att.Data))
			}

		}
		bw.Close()
	}
}

func isStandardHeader(name string) bool {
	switch strings.ToLower(name) {
	case "date", "subject", "from", "to", "cc", "message-id", "content-type":
		return true
	default:
		return false
	}
}

func toEnvelopeAddressList(list []smtp.Address) []imap.Address {
	if len(list) == 0 {
		return nil
	}

	result := make([]imap.Address, len(list))
	for i, addr := range list {
		result[i] = *toEnvelopeAddress(&addr)
	}
	return result
}

func toEnvelopeAddress(addr *smtp.Address) *imap.Address {
	if addr == nil {
		return nil
	}
	parts := strings.Split(addr.Address, "@")
	return &imap.Address{
		Name:    addr.Name,
		Mailbox: parts[0],
		Host:    parts[1],
	}
}
