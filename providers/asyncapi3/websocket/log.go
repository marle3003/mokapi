package websocket

import "mokapi/engine/common"

type Direction string

const (
	Receive Direction = "receive"
	Send    Direction = "send"
)

type Log struct {
	Channel    string           `json:"channel"`
	Message    LogValue         `json:"message"`
	MessageId  string           `json:"messageId"`
	Client     ClientLog        `json:"client"`
	ScriptFile string           `json:"script"`
	Api        string           `json:"api"`
	Actions    []*common.Action `json:"actions"`
}

type ClientLog struct {
	Id        string         `json:"id"`
	Query     map[string]any `json:"query"`
	Header    map[string]any `json:"header"`
	Address   string         `json:"address"`
	Server    string         `json:"server"`
	Direction Direction      `json:"direction"`
}

type LogValue struct {
	Value  string `json:"value"`
	Binary []byte `json:"binary"`
}

func (l *Log) Title() string {
	return ""
}

func clientLog(c *Client, direction Direction) ClientLog {
	return ClientLog{
		Id:        c.Id,
		Query:     c.Query,
		Header:    c.Header,
		Address:   c.RemoteAddr,
		Server:    c.ServerAddr,
		Direction: direction,
	}
}
