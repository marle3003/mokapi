package websocket

import (
	"fmt"
	"mokapi/engine/common"
)

type Direction string

const (
	Receive Direction = "receive"
	Send    Direction = "send"
)

type MessageLog struct {
	Channel   string           `json:"channel"`
	Message   LogValue         `json:"message"`
	MessageId string           `json:"messageId"`
	Direction Direction        `json:"direction"`
	Client    ClientLog        `json:"client"`
	Api       string           `json:"api"`
	Actions   []*common.Action `json:"actions"`
	// [messageId]error
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
		Message: LogValue{
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
	Type    string           `json:"type"`
	Channel string           `json:"channel"`
	Client  ClientLog        `json:"client"`
	Api     string           `json:"api"`
	Actions []*common.Action `json:"actions"`
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
