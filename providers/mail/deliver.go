package mail

import (
	"errors"
	"fmt"
	"mokapi/engine/common"
	"mokapi/runtime/monitor"
	"mokapi/smtp"
	"time"

	log "github.com/sirupsen/logrus"
)

const scriptClient = "mokapi-script"

// Deliver stores a mail in the mail API without an SMTP connection, as a script does with
// send(). It applies the checks an SMTP client would hit and is all or nothing: if the
// sender, one recipient or the mail itself is rejected, nothing is stored and the error
// carries the reject message. There is no authentication. monitor may be nil.
func (h *Handler) Deliver(from string, msg *smtp.Message, rcpts []string, monitor *monitor.Mail) error {
	start := time.Now()
	clientContext := &smtp.ClientContext{Client: scriptClient, From: from, To: rcpts}

	if rej := h.checkDelivery(from, msg, rcpts); rej != nil {
		h.logError(clientContext, rej.Message)
		return errors.New(rej.Message)
	}

	msg.UpdateSize()
	event := NewLogEvent(msg, clientContext, h.eh, h.logTraits())
	h.deliver(msg, clientContext, event, monitor)
	event.SetDuration(time.Since(start).Milliseconds())
	return nil
}

func (h *Handler) checkDelivery(from string, msg *smtp.Message, rcpts []string) *RejectResponse {
	if rej := h.checkSender(from); rej != nil {
		return rej
	}
	for accepted, rcpt := range rcpts {
		if rej := h.checkRcpt(rcpt, accepted); rej != nil {
			return rej
		}
	}
	return h.config.Rules.runMail(msg)
}

func (h *Handler) checkSender(from string) *RejectResponse {
	if len(h.config.Mailboxes) > 0 && !h.AllowUnknownSenders() {
		if _, ok := h.store.Mailbox(from); !ok {
			return rejectionOf(smtp.AddressRejected, fmt.Sprintf("Unknown mailbox %v", from))
		}
	}
	return h.config.Rules.RunSender(from)
}

// checkRcpt checks a recipient without creating its mailbox, because the mail may still be
// rejected later. accepted is the number of recipients accepted before.
func (h *Handler) checkRcpt(to string, accepted int) *RejectResponse {
	if _, ok := h.store.Mailbox(to); !ok && !h.config.Settings.AutoCreatesMailboxes() {
		return rejectionOf(smtp.AddressRejected, fmt.Sprintf("Unknown mailbox %v", to))
	}

	if rej := h.config.Rules.runRcpt(to); rej != nil {
		return rej
	}

	if s := h.config.Settings; s != nil && s.MaxRecipients > 0 && accepted+1 > s.MaxRecipients {
		return rejectionOf(smtp.TooManyRecipients, fmt.Sprintf("Too many recipients of %v reached", s.MaxRecipients))
	}
	return nil
}

func rejectionOf(status smtp.Status, message string) *RejectResponse {
	return &RejectResponse{
		StatusCode:         status.StatusCode,
		EnhancedStatusCode: status.EnhancedStatusCode,
		Message:            message,
	}
}

// deliver runs the event handlers, which may change the mail, then appends it to the INBOX
// of every recipient. Recipients are the envelope in clientContext.To, not the mail headers.
func (h *Handler) deliver(msg *smtp.Message, clientContext *smtp.ClientContext, event *Log, m *monitor.Mail) {
	event.SetActions(h.eventEmitter.EmitSmtp(&common.MailEvent{
		Api:        h.config.Info.Name,
		Recipients: clientContext.To,
		Message:    msg,
	}))
	event.UpdateContent(msg)

	for _, rcpt := range clientContext.To {
		box, ok := h.mailboxFor(rcpt)
		if !ok {
			// removed by a reload between RCPT and DATA
			log.Warnf("mailbox '%v' of '%v' no longer exists, mail not delivered to it", rcpt, h.config.Info.Name)
			continue
		}
		box.Append(msg)
	}

	log.Infof("received new mail on %v from client %v (%v)",
		h.config.Info.Name, clientContext.Client, clientContext.Addr)

	if m != nil {
		m.Mails.WithLabel(h.config.Info.Name, clientContext.From).Add(1)
		m.LastMail.WithLabel(h.config.Info.Name).Set(float64(time.Now().Unix()))
	}
}

// mailboxFor creates the mailbox on first delivery if the settings allow it.
func (h *Handler) mailboxFor(name string) (*Mailbox, bool) {
	if box, ok := h.store.Mailbox(name); ok {
		return box, true
	}
	if err := h.store.EnsureMailbox(name, h.config.Settings); err != nil {
		return nil, false
	}
	return h.store.Mailbox(name)
}
