package mail

import (
	"context"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"mime/multipart"
	"mokapi/engine/common"
	"mokapi/media"
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
		h.processMail(r.Message, r.Context)
	case smtp.Quit:
		rw.Write(smtp.StatusClose, smtp.Success, "Goodbye")
	}
}

func (h *Handler) processMail(m *smtp.MailMessage, ctx context.Context) {
	clientContext := smtp.ClientFromContext(ctx)
	mail := NewMail(m)
	mime := media.ParseContentType(m.ContentType)
	switch {
	case mime.Key() == "multipart/mixed":
		r := multipart.NewReader(m.Body, mime.Parameters["boundary"])
		for {
			p, err := r.NextPart()
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Errorf("smtp: unable to read message part: %v", err)
				break
			}
			if p.Header.Get("Content-Disposition") == "attachment" {
				mail.Attachment = append(mail.Attachment, newAttachment(p))
			} else {
				b, err := ioutil.ReadAll(p)
				if err != nil {
					log.Errorf("smtp: unable to read part: %v", err)
				}
				mail.Body += string(b)
			}
		}
	}

	log.Infof("recevied new mail on %v from client %v (%v)",
		h.config.Info.Name, clientContext.Client, clientContext.Addr)

	if m, ok := monitor.SmtpFromContext(ctx); ok {
		m.Mails.WithLabel(h.config.Info.Name).Add(1)
	}
	events.Push(m, events.NewTraits().WithNamespace("smtp").WithName(h.config.Info.Name))
	h.eventEmitter.Emit("smtp", mail)
}
