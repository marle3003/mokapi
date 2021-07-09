package actions

import (
	"fmt"
	"github.com/emersion/go-smtp"
	"mokapi/providers/workflow/runtime"
	"net"
	"strings"
)

type SendMail struct {
}

func (s *SendMail) Run(ctx *runtime.ActionContext) error {
	var server, from, to string

	server, _ = ctx.GetInputString("server")

	from, _ = ctx.GetInputString("from")

	to, _ = ctx.GetInputString("to")

	var subject, body, contentType, encoding string
	subject, _ = ctx.GetInputString("subject")

	body, _ = ctx.GetInputString("body")

	contentType, _ = ctx.GetInputString("contentType")

	encoding, _ = ctx.GetInputString("encoding")

	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("From: %s\r\nTo: %s\r\n", from, to))

	if len(subject) > 0 {
		msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	}

	if len(contentType) > 0 {
		msg.WriteString(fmt.Sprintf("Content-Type: %s\r\n", contentType))
	}

	if len(encoding) > 0 {
		msg.WriteString(fmt.Sprintf("Content-Transfer-Encoding: %s\r\n", encoding))
	}

	msg.WriteString(fmt.Sprintf("\r\n%s\rn\n", body))

	host, port, err := net.SplitHostPort(server)
	if err != nil {
		host = server
		port = "25"
	}

	return smtp.SendMail(fmt.Sprintf("%v:%v", host, port), nil, from, []string{to}, strings.NewReader(msg.String()))
}
