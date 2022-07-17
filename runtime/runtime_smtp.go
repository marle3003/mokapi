package runtime

import (
	"mokapi/config/dynamic/mail"
	"mokapi/runtime/monitor"
	"mokapi/smtp"
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

func (h *SmtpHandler) Serve(rw smtp.ResponseWriter, r *smtp.Request) {
	r.Context = monitor.NewSmtpContext(r.Context, h.smtp)
	h.next.Serve(rw, r)
}
