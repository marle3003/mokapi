package mail

import (
	"fmt"
	"mokapi/engine/common"
	"mokapi/runtime/events"
	"mokapi/smtp"

	log "github.com/sirupsen/logrus"
)

type Log struct {
	From      string           `json:"from"`
	To        []string         `json:"to"`
	MessageId string           `json:"messageId"`
	Subject   string           `json:"subject"`
	Duration  int64            `json:"duration"`
	Error     string           `json:"error"`
	Actions   []*common.Action `json:"actions"`
}

func NewLogEvent(msg *smtp.Message, ctx *smtp.ClientContext, eh events.Handler, traits events.Traits) *Log {
	event := &Log{
		From:     ctx.From,
		To:       ctx.To,
		Duration: 0,
		Actions:  nil,
	}

	if msg != nil {
		event.MessageId = msg.MessageId
		subject, err := smtp.DecodeHeaderValue(msg.Subject)
		if err != nil {
			log.Errorf("failed to decode subject: %v", err)
			event.Subject = msg.Subject
		} else {
			event.Subject = subject
		}
	}

	_ = eh.Push(event, traits.WithNamespace("mail"))
	return event
}

func (l *Log) Title() string {
	return fmt.Sprintf("%s", l.Subject)
}
