package websocket

import (
	"github.com/dop251/goja"
	log "github.com/sirupsen/logrus"
)

type EventType = string

const (
	ConnectEventType EventType = "connect"
	CloseEventType   EventType = "close"
	MessageEventType EventType = "message"
)

type Event struct {
	Type    EventType     `json:"type"`
	Api     string        `json:"api"`
	Channel *EventChannel `json:"channel"`
	Client  *EventClient  `json:"client"`
}

type ConnectEvent struct {
	Event
}

type CloseEvent struct {
	Event
	Reason   string `json:"reason"`
	ClosedBy string `json:"closedBy"`
}

type MessageEvent struct {
	Event
	Message any `json:"message"`
}

type EventChannel struct {
	Name    string         `json:"name"`
	Clients []*EventClient `json:"clients"`

	ch *Channel
}

type EventClient struct {
	RemoteAddress string         `json:"remoteAddress"`
	Query         map[string]any `json:"query"`
	Headers       map[string]any `json:"headers"`

	client *Client
}

func (e *MessageEvent) Reply(message any) {
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

func newEventChannel(ch *Channel) *EventChannel {
	ec := &EventChannel{
		Name: ch.Name,
		ch:   ch,
	}
	for _, c := range ch.clients {
		ec.Clients = append(ec.Clients, newEventClient(c))
	}
	return ec
}

func newEventClient(client *Client) *EventClient {
	c := &EventClient{
		RemoteAddress: client.RemoteAddr,
		Headers:       map[string]any{},
		Query:         map[string]any{},
		client:        client,
	}
	for k, v := range client.Query {
		c.Query[k] = v
	}
	for k, v := range client.Headers {
		c.Headers[k] = v
	}
	return c
}
