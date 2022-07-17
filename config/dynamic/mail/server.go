package mail

import (
	"context"
	log "github.com/sirupsen/logrus"
	"mokapi/engine/common"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"mokapi/smtp"
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

func (h *Handler) Serve(rw smtp.ResponseWriter, r *smtp.Request) {
	switch r.Cmd {
	case smtp.Hello:
		rw.Write(smtp.StatusOk, smtp.Undefined, "Hello "+r.Param)
	case smtp.From:
		ctx := smtp.ClientFromContext(r.Context)
		ctx.From = r.Param
		rw.Write(smtp.StatusOk, smtp.Success, "Ok")
	case smtp.Recipient:
		ctx := smtp.ClientFromContext(r.Context)
		ctx.To = append(ctx.To, r.Param)
		rw.Write(smtp.StatusOk, smtp.Success, "Ok")
	case smtp.Message:
		rw.Write(smtp.StatusOk, smtp.Success, "Ok")
		h.processMail(NewMail(r.Message), r.Context)
	case smtp.Quit:
		rw.Write(smtp.StatusClose, smtp.Success, "Goodbye")
	}
}

func (h *Handler) processMail(m *Mail, ctx context.Context) {
	clientContext := smtp.ClientFromContext(ctx)

	log.Infof("recevied new mail on %v from client %v (%v)",
		h.config.Info.Name, clientContext.Client, clientContext.Addr)

	if m, ok := monitor.SmtpFromContext(ctx); ok {
		m.Mails.WithLabel(h.config.Info.Name).Add(1)
	}
	events.Push(m, events.NewTraits().WithNamespace("smtp").WithName(h.config.Info.Name))
	h.eventEmitter.Emit("smtp", m)
}
