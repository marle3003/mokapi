package mail

import (
	"mokapi/engine/common"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"mokapi/smtp"
	"time"
)

type Handler struct {
	config       *Config
	eventEmitter common.MailEventEmitter
	store        *Store
	eh           events.Handler
}

func NewHandler(config *Config, store *Store, eventEmitter common.MailEventEmitter, eh events.Handler) *Handler {
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
		for _, mb := range h.store.Mailboxes() {
			if username, password := mb.Credentials(); username == req.Username && password == req.Password {
				_ = rw.Write(&smtp.LoginResponse{Result: &smtp.AuthSucceeded})
				ctx.Auth = username
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
	m, _ := monitor.SmtpFromContext(ctx)
	event := NewLogEvent(r.Message, clientContext, h.eh, h.logTraits())
	defer func() {
		i := ctx.Value("time")
		if i != nil {
			t := i.(time.Time)
			event.SetDuration(time.Now().Sub(t).Milliseconds())
		}
	}()

	if rej := h.config.Rules.runMail(r.Message); rej != nil {
		h.writeRejection(rw, r, rej)
		return
	}

	h.deliver(r.Message, clientContext, event, m)

	_ = rw.Write(&smtp.DataResponse{Result: smtp.Ok})
}

func (h *Handler) serveMail(rw smtp.ResponseWriter, r *smtp.MailRequest, ctx *smtp.ClientContext) {
	if h.requiresAuth(r.From, ctx) {
		h.writeErrorResponse(rw, r, smtp.AuthRequired, "")
		return
	}

	if rej := h.checkSender(r.From); rej != nil {
		h.writeRejection(rw, r, rej)
		return
	}

	ctx.From = r.From
	_ = rw.Write(&smtp.MailResponse{Result: smtp.Ok})
}

func (h *Handler) serveRcpt(rw smtp.ResponseWriter, r *smtp.RcptRequest, ctx *smtp.ClientContext) {
	if rej := h.checkRcpt(r.To, len(ctx.To)); rej != nil {
		h.writeRejection(rw, r, rej)
		return
	}

	ctx.To = append(ctx.To, r.To)
	_ = rw.Write(&smtp.RcptResponse{Result: smtp.Ok})
}

// requiresAuth is true for a known sender whose mailbox has credentials, but the client
// has not logged in. Only SMTP clients authenticate, so Deliver doesn't call it.
func (h *Handler) requiresAuth(from string, ctx *smtp.ClientContext) bool {
	if len(h.config.Mailboxes) == 0 {
		return false
	}
	mb, ok := h.store.Mailbox(from)
	if !ok {
		return false
	}
	username, _ := mb.Credentials()
	return len(username) > 0 && len(ctx.Auth) == 0
}

func (h *Handler) writeErrorResponse(rw smtp.ResponseWriter, r smtp.Request, status smtp.Status, message string) {
	if len(message) > 0 {
		status.Message = message
	}
	h.logError(smtp.ClientFromContext(r.Context()), status.Message)
	_ = rw.Write(r.NewResponse(&status))
}

func (h *Handler) writeRejection(rw smtp.ResponseWriter, r smtp.Request, rej *RejectResponse) {
	h.logError(smtp.ClientFromContext(r.Context()), rej.Message)
	_ = rw.Write(r.NewResponse(&smtp.Status{
		StatusCode:         rej.StatusCode,
		EnhancedStatusCode: rej.EnhancedStatusCode,
		Message:            rej.Message,
	}))
}

func (h *Handler) logError(clientContext *smtp.ClientContext, message string) {
	l := NewLogEvent(nil, clientContext, h.eh, h.logTraits())
	l.Error = message
}

func (h *Handler) logTraits() events.Traits {
	return events.NewTraits().WithName(h.config.Info.Name)
}

func (h *Handler) AllowUnknownSenders() bool {
	if h.config.Settings == nil {
		return true
	}
	return h.config.Settings.AllowUnknownSenders
}
