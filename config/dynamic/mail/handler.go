package mail

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/engine/common"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"mokapi/smtp"
	"regexp"
)

type Handler struct {
	config       *Config
	eventEmitter common.EventEmitter
}

func NewHandler(config *Config, eventEmitter common.EventEmitter) *Handler {
	return &Handler{
		config:       config,
		eventEmitter: eventEmitter,
	}
}

func (h *Handler) ServeSMTP(rw smtp.ResponseWriter, r smtp.Request) {
	ctx := smtp.ClientFromContext(r.Context())
	switch req := r.(type) {
	case *smtp.LoginRequest:
		for _, acc := range h.config.Mailboxes {
			if acc.Username == req.Username && acc.Password == req.Password {
				rw.Write(&smtp.LoginResponse{Result: smtp.AuthSucceeded})
				ctx.Auth = acc.Username
			}
		}
		rw.Write(&smtp.LoginResponse{Result: smtp.InvalidAuthCredentials})
	case *smtp.MailRequest:
		h.serveMail(rw, req, ctx)
	case *smtp.RcptRequest:
		h.serveRcpt(rw, req, ctx)
	case *smtp.DataRequest:
		err := h.processMail(req.Message, r.Context())
		if err != nil {
			rw.Write(&smtp.DataResponse{Result: &smtp.SMTPStatus{
				Code:    smtp.StatusReject,
				Status:  smtp.EnhancedStatusCode{5, 0, 0}, // Invalid command arguments
				Message: err.Error(),
			}})
			return
		}
		rw.Write(&smtp.DataResponse{Result: smtp.Ok})
	}
}

func (h *Handler) processMail(msg *smtp.Message, ctx context.Context) error {
	clientContext := smtp.ClientFromContext(ctx)
	m := NewMail(msg)

	if err := h.runRules(m); err != nil {
		return err
	}

	log.Infof("recevied new mail on %v from client %v (%v)",
		h.config.Info.Name, clientContext.Client, clientContext.Addr)

	if m, ok := monitor.SmtpFromContext(ctx); ok {
		m.Mails.WithLabel(h.config.Info.Name).Add(1)
	}
	events.Push(m, events.NewTraits().WithNamespace("smtp").WithName(h.config.Info.Name))
	h.eventEmitter.Emit("smtp", msg)

	return nil
}

func (h *Handler) runRules(m *Mail) error {
	for _, r := range h.config.AllowList {
		if len(r.Sender) > 0 {
			var senderAddress string
			if m.Sender != nil {
				senderAddress = m.Sender.Address
			} else if len(m.From) > 0 {
				senderAddress = m.From[0].Address
			} else {
				return fmt.Errorf("required from address")
			}
			if b, err := regexp.Match(r.Sender, []byte(senderAddress)); err != nil {
				return err
			} else if !b {
				return fmt.Errorf("sender %v does not match allow rule: %v", senderAddress, r.Sender)
			}
		}
	}
	return nil
}

func (h *Handler) serveMail(rw smtp.ResponseWriter, r *smtp.MailRequest, ctx *smtp.ClientContext) {
	if len(h.config.Mailboxes) > 0 {
		if m, ok := h.config.getMailbox(r.From); !ok {
			rw.Write(&smtp.MailResponse{Result: smtp.AddressRejected})
			return
		} else if len(m.Username) > 0 && len(ctx.Auth) == 0 {
			rw.Write(&smtp.MailResponse{Result: smtp.AuthRequired})
			return
		}
	}
	ctx.From = r.From
	rw.Write(&smtp.MailResponse{Result: smtp.Ok})
}

func (h *Handler) serveRcpt(rw smtp.ResponseWriter, r *smtp.RcptRequest, ctx *smtp.ClientContext) {
	if len(h.config.Mailboxes) > 0 {
		if _, ok := h.config.getMailbox(r.To); !ok {
			rw.Write(&smtp.RcptResponse{Result: smtp.AddressRejected})
			return
		}
	}

	if h.config.MaxRecipients > 0 && len(ctx.To)+1 > h.config.MaxRecipients {
		rw.Write(&smtp.RcptResponse{Result: smtp.TooManyRecipientsWithMessage(fmt.Sprintf("Too many recipients of %v reached", h.config.MaxRecipients))})
		return
	}
	ctx.To = append(ctx.To, r.To)
	rw.Write(&smtp.RcptResponse{Result: smtp.Ok})
}
