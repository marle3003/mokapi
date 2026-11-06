package websocket

import (
	"encoding/json"
	"fmt"
	"mokapi/engine/common"
	"sync"
)

type Direction string

const (
	Receive Direction = "receive"
	Send    Direction = "send"
)

// actionLog holds the actions of an event handler. A log is stored in the event store
// before the handler has finished, so the actions are attached later while readers
// (e.g. the dashboard API) may already access the log.
type actionLog struct {
	m       sync.Mutex
	actions []*common.Action
}

func (l *actionLog) SetActions(actions []*common.Action) {
	l.m.Lock()
	defer l.m.Unlock()
	l.actions = actions
}

func (l *actionLog) GetActions() []*common.Action {
	l.m.Lock()
	defer l.m.Unlock()
	return l.actions
}

type MessageLog struct {
	actionLog
	Channel   string
	Message   *LogValue
	MessageId string
	Direction Direction
	Client    ClientLog
	Api       string
	// [messageId]error
	ValidationErrors map[string]string
}

func (l *MessageLog) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Channel          string            `json:"channel"`
		Message          *LogValue         `json:"message"`
		MessageId        string            `json:"messageId"`
		Direction        Direction         `json:"direction"`
		Client           ClientLog         `json:"client"`
		Api              string            `json:"api"`
		Actions          []*common.Action  `json:"actions"`
		ValidationErrors map[string]string `json:"validationErrors,omitempty"`
	}{l.Channel, l.Message, l.MessageId, l.Direction, l.Client, l.Api, l.GetActions(), l.ValidationErrors})
}

type MessageLogInfo struct {
	Channel          string            `json:"channel"`
	Message          *LogValue         `json:"message"`
	Direction        Direction         `json:"direction"`
	Client           ClientLog         `json:"client"`
	ValidationErrors map[string]string `json:"validationErrors,omitempty"`
}

type ClientLog struct {
	Id      string         `json:"id"`
	Query   map[string]any `json:"query"`
	Header  map[string]any `json:"header"`
	Address string         `json:"address"`
	Server  string         `json:"server"`
}

type LogValue struct {
	Value  string `json:"value"`
	Binary []byte `json:"binary"`
}

func (l *MessageLog) Title() string {
	return ""
}

func messageLog(c *Channel, data []byte, messageId string, client *Client, direction Direction) *MessageLog {
	l := &MessageLog{
		Channel: c.Name,
		Message: &LogValue{
			Value:  string(data),
			Binary: data,
		},
		MessageId: messageId,
		Api:       c.api,
		Client:    clientLog(client),
		Direction: direction,
	}
	return l
}

func clientLog(c *Client) ClientLog {
	return ClientLog{
		Id:      c.Id,
		Query:   c.Query,
		Header:  c.Headers,
		Address: c.RemoteAddr,
		Server:  c.ServerAddr,
	}
}

type ConnectLog struct {
	actionLog
	Type    string
	Channel string
	Client  ClientLog
	Api     string
}

func (l *ConnectLog) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type    string           `json:"type"`
		Channel string           `json:"channel"`
		Client  ClientLog        `json:"client"`
		Api     string           `json:"api"`
		Actions []*common.Action `json:"actions"`
	}{l.Type, l.Channel, l.Client, l.Api, l.GetActions()})
}

func connectLog(c *Channel, client *Client) *ConnectLog {
	return &ConnectLog{
		Type:    "connect",
		Channel: c.Name,
		Client:  clientLog(client),
		Api:     c.api,
	}
}

func (l *ConnectLog) Title() string {
	return fmt.Sprintf("Connect to %s", l.Channel)
}

type CloseLog struct {
	Type     string           `json:"type"`
	Channel  string           `json:"channel"`
	Reason   string           `json:"reason"`
	ClosedBy string           `json:"closedBy"`
	Client   ClientLog        `json:"client"`
	Api      string           `json:"api"`
	Actions  []*common.Action `json:"actions"`
}

func closeLog(c *Channel, client *Client, reason, closedBy string) *CloseLog {
	return &CloseLog{
		Type:     "close",
		Channel:  c.Name,
		Client:   clientLog(client),
		Api:      c.api,
		Reason:   reason,
		ClosedBy: closedBy,
	}
}

func (l *CloseLog) Title() string {
	return fmt.Sprintf("Connect to %s", l.Channel)
}
