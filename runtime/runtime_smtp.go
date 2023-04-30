package runtime

import (
	"context"
	"mokapi/config/dynamic/mail"
	"mokapi/runtime/monitor"
	"mokapi/smtp"
	"time"
)

type SmtpInfo struct {
	*mail.Config
}

type SmtpHandler struct {
	smtp *monitor.Smtp
	next smtp.Handler
}

func NewSmtpHandler(smtp *monitor.Smtp, next smtp.Handler) *SmtpHandler {
	return &SmtpHandler{smtp: smtp, next: next}
}

func (h *SmtpHandler) ServeSMTP(rw smtp.ResponseWriter, r smtp.Request) {
	r.WithContext(monitor.NewSmtpContext(r.Context(), h.smtp))
	r.WithContext(context.WithValue(r.Context(), "time", time.Now()))

	h.next.ServeSMTP(rw, r)
}
