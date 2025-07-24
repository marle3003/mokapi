package mail

import (
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
	store        *Store
	eh           events.Handler
}

func NewHandler(config *Config, store *Store, eventEmitter common.EventEmitter, eh events.Handler) *Handler {
	return &Handler{
		config:       config,
		eventEmitter: eventEmitter,
		store:        store,
		eh:           eh,
	}
}

func (h *Handler) ServeSMTP(rw smtp.ResponseWriter, r smtp.Request) {
	ctx := smtp.ClientFromContext(r.Context())
	switch req := r.(type) {
	case *smtp.LoginRequest:
		// todo disable Auth capability on server if no mailbox is defined
		for _, acc := range h.config.Mailboxes {
			if acc.Username == req.Username && acc.Password == req.Password {
				rw.Write(&smtp.LoginResponse{Result: &smtp.AuthSucceeded})
				ctx.Auth = acc.Username
				return
			}
		}
		h.writeErrorResponse(rw, r, smtp.InvalidAuthCredentials, "")
	case *smtp.MailRequest:
		h.serveMail(rw, req, ctx)
	case *smtp.RcptRequest:
		h.serveRcpt(rw, req, ctx)
	case *smtp.DataRequest:
		h.processMail(rw, req)
	}
}

func (h *Handler) processMail(rw smtp.ResponseWriter, r *smtp.DataRequest) {
	ctx := r.Context()
	clientContext := smtp.ClientFromContext(ctx)
	m, doMonitor := monitor.SmtpFromContext(ctx)
	event := NewLogEvent(r.Message, clientContext, h.eh, events.NewTraits().WithName(h.config.Info.Name))
	defer func() {
		i := ctx.Value("time")
		if i != nil {
			t := i.(time.Time)
			event.Duration = time.Now().Sub(t).Milliseconds()
		}
	}()

	if res := h.config.Rules.runMail(r.Message); res != nil {
		h.writeRuleResponse(rw, r, res)
		return
	}

	for _, rcpt := range clientContext.To {
		box := h.store.Mailboxes[rcpt]
		box.Append(r.Message)
	}

	log.Infof("received new mail on %v from client %v (%v)",
		h.config.Info.Name, clientContext.Client, clientContext.Addr)

	if doMonitor {
		m.Mails.WithLabel(h.config.Info.Name, clientContext.From).Add(1)
		m.LastMail.WithLabel(h.config.Info.Name).Set(float64(time.Now().Unix()))
	}

	res := &smtp.DataResponse{Result: smtp.Ok}
	event.Actions = h.eventEmitter.Emit("smtp", r.Message, res.Result)

	rw.Write(res)
}

func (h *Handler) serveMail(rw smtp.ResponseWriter, r *smtp.MailRequest, ctx *smtp.ClientContext) {
	if len(h.config.Mailboxes) > 0 {

		if m, ok := h.store.Mailboxes[r.From]; !ok {
			h.writeErrorResponse(rw, r, smtp.AddressRejected, fmt.Sprintf("Unknown mailbox %v", r.From))
			return
		} else if len(m.Username) > 0 && len(ctx.Auth) == 0 {
			h.writeErrorResponse(rw, r, smtp.AuthRequired, "")
			return
		}
	}

	res := h.config.Rules.RunSender(r.From)
	if res != nil {
		h.writeRuleResponse(rw, r, res)
	} else {
		ctx.From = r.From
		rw.Write(&smtp.MailResponse{Result: smtp.Ok})
	}
}

func (h *Handler) serveRcpt(rw smtp.ResponseWriter, r *smtp.RcptRequest, ctx *smtp.ClientContext) {
	if err := h.store.EnsureMailbox(r.To); err != nil {
		h.writeErrorResponse(rw, r, smtp.AddressRejected, fmt.Sprintf("Unknown mailbox %v", r.To))
		return
	}

	res := h.config.Rules.RunSender(r.To)
	if res != nil {
		h.writeRuleResponse(rw, r, res)
		return
	}

	if h.config.Settings != nil {
		if h.config.Settings.MaxRecipients > 0 && len(ctx.To)+1 > h.config.Settings.MaxRecipients {
			h.writeErrorResponse(rw, r, smtp.TooManyRecipients, fmt.Sprintf("Too many recipients of %v reached", h.config.Settings.MaxRecipients))
			return
		}
	}

	ctx.To = append(ctx.To, r.To)
	rw.Write(&smtp.RcptResponse{Result: smtp.Ok})
}

func (h *Handler) writeErrorResponse(rw smtp.ResponseWriter, r smtp.Request, status smtp.Status, message string) {
	clientContext := smtp.ClientFromContext(r.Context())
	if len(message) > 0 {
		status.Message = message
	}
	res := r.NewResponse(&status)
	l := NewLogEvent(nil, clientContext, h.eh, events.NewTraits().WithName(h.config.Info.Name))
	l.Error = status.Message
	_ = rw.Write(res)
}

func (h *Handler) writeRuleResponse(rw smtp.ResponseWriter, r smtp.Request, response *RejectResponse) {
	clientContext := smtp.ClientFromContext(r.Context())
	res := r.NewResponse(&smtp.Status{
		StatusCode:         response.StatusCode,
		EnhancedStatusCode: response.EnhancedStatusCode,
		Message:            response.Text,
	})
	l := NewLogEvent(nil, clientContext, h.eh, events.NewTraits().WithName(h.config.Info.Name))
	l.Error = response.Text
	_ = rw.Write(res)
}
