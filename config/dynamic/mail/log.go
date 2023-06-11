package mail

import (
	"mokapi/engine/common"
	"mokapi/runtime/events"
	"mokapi/smtp"
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

func NewLogEvent(msg *smtp.Message, ctx *smtp.ClientContext, traits events.Traits) *Log {
	event := &Log{
		From:     ctx.From,
		To:       ctx.To,
		Duration: 0,
		Actions:  nil,
	}

	if msg != nil {
		event.MessageId = msg.MessageId
		event.Subject = msg.Subject
	}

	_ = events.Push(event, traits.WithNamespace("smtp"))
	return event
}
