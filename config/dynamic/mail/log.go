package mail

import (
	"mokapi/engine/common"
	"mokapi/runtime/events"
	"mokapi/smtp"
)

type Log struct {
	From     string           `json:"from"`
	To       []string         `json:"to"`
	Mail     *Mail            `json:"mail"`
	Duration int64            `json:"duration"`
	Actions  []*common.Action `json:"actions"`
}

func NewLogEvent(msg *Mail, ctx *smtp.ClientContext, traits events.Traits) *Log {
	event := &Log{
		From:     ctx.From,
		To:       ctx.To,
		Mail:     msg,
		Duration: 0,
		Actions:  nil,
	}
	_ = events.Push(event, traits.WithNamespace("smtp"))
	return event
}
