package mail

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/engine/common"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"mokapi/smtp"
	"time"
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
				rw.Write(&smtp.LoginResponse{Result: &smtp.AuthSucceeded})
				ctx.Auth = acc.Username
			}
		}
		h.writeErrorResponse(rw, r, smtp.InvalidAuthCredentials, "")
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
	monitor, doMonitor := monitor.SmtpFromContext(ctx)
	m := NewMail(msg)
	event := NewLogEvent(m, clientContext, events.NewTraits().WithName(h.config.Info.Name))
	defer func() {
		i := ctx.Value("time")
		if i != nil {
			t := i.(time.Time)
			event.Duration = time.Now().Sub(t).Milliseconds()
		}
	}()

	if err := h.runRules(m); err != nil {
		return err
	}

	log.Infof("recevied new mail on %v from client %v (%v)",
		h.config.Info.Name, clientContext.Client, clientContext.Addr)

	if doMonitor {
		monitor.Mails.WithLabel(h.config.Info.Name).Add(1)
		monitor.LastMail.WithLabel(h.config.Info.Name).Set(float64(time.Now().Unix()))
	}

	event.Actions = h.eventEmitter.Emit("smtp", m)

	return nil
}

func (h *Handler) runRules(m *Mail) error {
	for _, r := range h.config.Rules {
		_ = r
		//if len(r.Sender) > 0 {
		//	var senderAddress string
		//	if m.Sender != nil {
		//		senderAddress = m.Sender.Address
		//	} else if len(m.From) > 0 {
		//		senderAddress = m.From[0].Address
		//	} else if r.Action == Allow {
		//		return fmt.Errorf("required from address")
		//	}
		//	if b, err := regexp.Match(r.Sender, []byte(senderAddress)); err != nil {
		//		return err
		//	} else if !b && r.Action == Allow {
		//		return fmt.Errorf("sender %v does not match allow rule: %v", senderAddress, r.Sender)
		//	} else if b && r.Action == Deny {
		//		return fmt.Errorf("sender %v does match deny rule: %v", senderAddress, r.Sender)
		//	}
		//}
	}
	return nil
}

func (h *Handler) serveMail(rw smtp.ResponseWriter, r *smtp.MailRequest, ctx *smtp.ClientContext) {
	if len(h.config.Mailboxes) > 0 {
		if m, ok := h.config.getMailbox(r.From); !ok {
			h.writeErrorResponse(rw, r, smtp.AddressRejected, fmt.Sprintf("Unknown mailbox %v", r.From))
			return
		} else if len(m.Username) > 0 && len(ctx.Auth) == 0 {
			h.writeErrorResponse(rw, r, smtp.AuthRequired, "")
			return
		}
	}

	for _, rule := range h.config.Rules {
		if rule.Sender != nil {
			match := rule.Sender.Match(r.From)
			if match && rule.Action == Deny {
				h.writeErrorResponse(rw, r, smtp.AddressRejected, fmt.Sprintf("sender %v does match deny rule: %v", r.From, rule.Sender))
				return
			} else if !match && rule.Action == Allow {
				h.writeErrorResponse(rw, r, smtp.AddressRejected, fmt.Sprintf("sender %v does not match allow rule: %v", r.From, rule.Sender))
				return
			}
		}
	}

	ctx.From = r.From
	rw.Write(&smtp.MailResponse{Result: smtp.Ok})
}

func (h *Handler) serveRcpt(rw smtp.ResponseWriter, r *smtp.RcptRequest, ctx *smtp.ClientContext) {
	if len(h.config.Mailboxes) > 0 {
		if _, ok := h.config.getMailbox(r.To); !ok {
			h.writeErrorResponse(rw, r, smtp.AddressRejected, fmt.Sprintf("Unknown mailbox %v", r.To))
			return
		}
	}

	for _, rule := range h.config.Rules {
		if rule.Recipient != nil {
			match := rule.Recipient.Match(r.To)
			if match && rule.Action == Deny {
				h.writeErrorResponse(rw, r, smtp.BadDestinationAddress, fmt.Sprintf("recipient %v does match deny rule: %v", r.To, rule.Recipient))
				return
			} else if !match && rule.Action == Allow {
				h.writeErrorResponse(rw, r, smtp.BadDestinationAddress, fmt.Sprintf("recipient %v does not match allow rule: %v", r.To, rule.Recipient))
				return
			}
		}
	}

	if h.config.MaxRecipients > 0 && len(ctx.To)+1 > h.config.MaxRecipients {
		h.writeErrorResponse(rw, r, smtp.TooManyRecipients, fmt.Sprintf("Too many recipients of %v reached", h.config.MaxRecipients))
		return
	}
	ctx.To = append(ctx.To, r.To)
	rw.Write(&smtp.RcptResponse{Result: smtp.Ok})
}

func (h *Handler) writeErrorResponse(rw smtp.ResponseWriter, r smtp.Request, status smtp.SMTPStatus, message string) {
	clientContext := smtp.ClientFromContext(r.Context())
	if len(message) > 0 {
		status.Message = message
	}
	res := r.NewResponse(&status)
	l := NewLogEvent(nil, clientContext, events.NewTraits().WithName(h.config.Info.Name))
	l.Error = status.Message
	_ = rw.Write(res)
}
