package mail

import (
	"encoding/json"
	"fmt"
	"mokapi/engine/common"
	"mokapi/runtime/events"
	"mokapi/smtp"
	"sync"

	log "github.com/sirupsen/logrus"
)

// Log is pushed to the event store before the request has been processed. Duration and
// actions are set afterwards (see common.EventLog). MessageId and Subject follow changes of
// the event handlers (see UpdateContent), so readers lock mu. Other fields are final after
// NewLogEvent.
type Log struct {
	common.EventLog
	mu        sync.Mutex
	From      string
	To        []string
	MessageId string
	Subject   string
	Error     string
}

func (l *Log) MarshalJSON() ([]byte, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return json.Marshal(struct {
		From      string           `json:"from"`
		To        []string         `json:"to"`
		MessageId string           `json:"messageId"`
		Subject   string           `json:"subject"`
		Duration  int64            `json:"duration"`
		Error     string           `json:"error"`
		Actions   []*common.Action `json:"actions"`
	}{l.From, l.To, l.MessageId, l.Subject, l.GetDuration(), l.Error, l.GetActions()})
}

type LogInfo struct {
	MessageId string   `json:"messageId"`
	From      string   `json:"from"`
	To        []string `json:"to"`
	Subject   string   `json:"subject"`
	Duration  int64    `json:"duration"`
}

func NewLogEvent(msg *smtp.Message, ctx *smtp.ClientContext, eh events.Handler, traits events.Traits) *Log {
	event := &Log{
		From: ctx.From,
		To:   ctx.To,
	}

	if msg != nil {
		event.UpdateContent(msg)
	}

	_ = eh.Push(event, traits.WithNamespace("mail"))
	return event
}

// UpdateContent takes the mail fields shown in the log from msg, e.g. after event handlers
// changed the mail.
func (l *Log) UpdateContent(msg *smtp.Message) {
	subject, err := smtp.DecodeHeaderValue(msg.Subject)
	if err != nil {
		log.Errorf("failed to decode subject: %v", err)
		subject = msg.Subject
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.MessageId = msg.MessageId
	l.Subject = subject
}

func (l *Log) Title() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return fmt.Sprintf("%s", l.Subject)
}

func (l *Log) IndexFields() map[string]any {
	l.mu.Lock()
	defer l.mu.Unlock()
	m := map[string]any{
		"from":               l.From,
		"to":                 l.To,
		"messageId":          l.MessageId,
		"subject":            l.Subject,
		"metadata.messageId": l.MessageId,
	}
	return m
}
