package websocket

import (
	"github.com/dop251/goja"
	log "github.com/sirupsen/logrus"
)

type Event struct {
	Api     string       `json:"api"`
	Channel EventChannel `json:"channel"`
	Client  *EventClient `json:"client"`
	Message any          `json:"message"`
}

type EventChannel struct {
	Name string `json:"name"`
	ch   *Channel
}

type EventClient struct {
	RemoteAddress string         `json:"remoteAddress"`
	Query         map[string]any `json:"query"`
	Headers       map[string]any `json:"headers"`

	client *Client
}

func (e *Event) Reply(message any) {
	e.Client.Send(message)
}

func (e *Event) Broadcast(message goja.Value) {
	m := message.Export()
	_ = m
	for _, c := range e.Channel.ch.clients {
		err := c.sendMessage(message)
		if err != nil {
			panic(err)
		}
	}
}

func (c *EventClient) Send(message any) {
	log.Infof("sending message to %s", c.RemoteAddress)
	err := c.client.sendMessage(message)
	if err != nil {
		panic(err)
	}
}
