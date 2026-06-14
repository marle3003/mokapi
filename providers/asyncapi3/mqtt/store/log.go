package store

import (
	"mokapi/engine/common"
	"mokapi/mqtt"
	"mokapi/runtime/events"
)

type LogMessage struct {
	Topic      string           `json:"topic"`
	Message    LogValue         `json:"message"`
	Retain     bool             `json:"retain"`
	MessageId  string           `json:"messageId"`
	Api        string           `json:"api"`
	ClientId   string           `json:"clientId"`
	ScriptFile string           `json:"script"`
	Actions    []*common.Action `json:"actions"`
}

type LogValue struct {
	Value  string `json:"value"`
	Binary []byte `json:"binary"`
}

func (l *LogMessage) Title() string {
	return l.Topic
}

type RequestLogEvent struct {
	Api      string     `json:"api"`
	Type     mqtt.Type  `json:"type"`
	Request  RequestLog `json:"request"`
	Response any        `json:"response"`
}

func (r *RequestLogEvent) Title() string {
	return r.Request.Title()
}

type RequestLog interface {
	Title() string
}

type ConnectRequest struct {
	Version      byte            `json:"version"`
	CleanSession bool            `json:"cleanSession"`
	KeepAlive    int16           `json:"keepAlive"`
	Message      *PublishMessage `json:"message,omitempty"`
	Username     string          `json:"username,omitempty"`
	Password     string          `json:"password,omitempty"`
}

func (r *ConnectRequest) Title() string {
	return "Connect"
}

type PublishMessage struct {
	QoS     byte   `json:"qos"`
	Retain  bool   `json:"retain"`
	Topic   string `json:"topic"`
	Message string `json:"value"`
}

type ConnectResponse struct {
	SessionPresent bool      `json:"sessionPresent"`
	ReasonCode     mqtt.Code `json:"reasonCode"`
}

type SubscribeRequest struct {
	MessageId uint16           `json:"messageId"`
	Topics    []SubscribeTopic `json:"topics"`
}

func (r *SubscribeRequest) Title() string {
	return "Subscribe"
}

type SubscribeTopic struct {
	Name string `json:"name"`
	QoS  byte   `json:"qos"`
}

type SubscribeResponse struct {
	ReasonCodes []mqtt.SubscriptionReason `json:"reasonCodes"`
}

type DisconnectRequest struct {
	Reason mqtt.DisconnectReason `json:"reason"`
}

func (r *DisconnectRequest) Title() string {
	return "Disconnect"
}

func (s *Store) logRequest(req RequestLog, res any, ctx *mqtt.ClientContext) {
	log := &RequestLogEvent{Api: s.cfg.Info.Name, Request: req}

	switch req.(type) {
	case *ConnectRequest:
		log.Type = mqtt.CONNECT
	case *SubscribeRequest:
		log.Type = mqtt.SUBSCRIBE
	case *DisconnectRequest:
		log.Type = mqtt.DISCONNECT
	}

	log.Response = res
	t := events.NewTraits().
		WithNamespace("mqtt").
		WithName(s.cfg.Info.Name).
		With("type", "request")
	if ctx != nil {
		t.With("clientId", ctx.ClientId)
	}
	_ = s.eh.Push(log, t)
}
