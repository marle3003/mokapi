package websocket

import (
	"context"
	engine "mokapi/engine/common"
	"mokapi/media"
	"mokapi/providers/asyncapi3"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"mokapi/schema/encoding"
	"sync"
	"time"

	"github.com/coder/websocket"
	log "github.com/sirupsen/logrus"
)

type Channel struct {
	Name    string
	api     string
	clients map[string]*Client
	m       sync.RWMutex
	cfg     *asyncapi3.Channel
	emitter engine.EventEmitter
	log     func(log *Log, traits events.Traits)
	monitor *monitor.Websocket
}

func (s *Store) Channel(name string) (*Channel, bool) {
	s.m.RLock()
	if t, ok := s.Channels[name]; ok {
		s.m.RUnlock()
		return t, ok
	}
	s.m.RUnlock()

	s.m.Lock()
	defer s.m.Unlock()

	for _, ref := range s.cfg.Channels {
		if ref == nil || ref.Value == nil {
			continue
		}
		ch := ref.Value
		if !ch.IsChannelAvailable("ws") {
			continue
		}
		if len(ch.Parameters) == 0 {
			continue
		}

		err := ch.IsNameValid(name)
		if err != nil {
			continue
		}

		if s.Channels == nil {
			s.Channels = make(map[string]*Channel)
		}

		c := &Channel{
			Name:    name,
			api:     s.cfg.Info.Name,
			emitter: s.emitter,
			cfg:     ch,
			log:     s.log,
			monitor: s.monitor,
		}
		s.Channels[name] = c
		return c, true
	}

	return nil, false
}

func (c *Channel) readLoop(ctx context.Context, conn *websocket.Conn, client *Client) error {
	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			return err
		}

		var msg any
		var messageId string
		for id, m := range c.cfg.Messages {
			if m.Value == nil || m.Value.Payload == nil || m.Value.Payload.Value == nil {
				continue
			}
			messageId = id
			var p encoding.Parser
			p, err = m.Value.Payload.GetParser(m.Value.ContentType)

			if err != nil {
				log.Errorf("unsupported payload type: %T", m.Value.Payload.Value)
			}
			msg, err = encoding.Decode(data, encoding.WithContentType(media.ParseContentType(m.Value.ContentType)), encoding.WithParser(p))
			if err == nil {
				break
			}
		}
		if err != nil {
			return err
		}

		channelName := c.cfg.ResolveAddress()

		l := c.newLog(data, messageId, client, Send)
		// log event before run event engine to have the correct log order if event handler sends a message
		c.log(l, events.NewTraits().With("channel", channelName).With("clientId", client.Id))

		l.Actions = c.runEvent(client, msg)

		labels := []string{c.api, channelName}
		c.monitor.Messages.WithLabel(labels...).Add(1)
		c.monitor.LastMessage.WithLabel(labels...).Set(float64(time.Now().Unix()))
	}
}

func (c *Channel) addClient(client *Client) {
	c.m.Lock()
	if c.clients == nil {
		c.clients = make(map[string]*Client)
	}
	c.clients[client.Id] = client
	c.m.Unlock()
}

func (c *Channel) removeClient(client *Client) {
	c.m.Lock()
	defer c.m.Unlock()
	delete(c.clients, client.Id)
}

func (c *Channel) newEvent(client *Client, v any) *Event {
	evt := &Event{
		Api:     c.api,
		Channel: newEventChannel(c),
		Client:  newEventClient(client),
		Message: v,
	}
	return evt
}

func (c *Channel) newLog(data []byte, messageId string, client *Client, direction Direction) *Log {
	l := &Log{
		Channel: c.Name,
		Message: LogValue{
			Value:  string(data),
			Binary: data,
		},
		MessageId: messageId,
		Api:       c.api,
		Client:    clientLog(client, direction),
	}
	return l
}

func (c *Channel) runEvent(client *Client, msg any) []*engine.Action {
	evt := c.newEvent(client, msg)
	if client.Query != nil {
		evt.Client.Query = client.Query
	} else {
		evt.Client.Query = map[string]any{}
	}
	if client.Header != nil {
		evt.Client.Headers = client.Header
	} else {
		evt.Client.Headers = map[string]any{}
	}
	return c.emitter.Emit("websocket", evt)
}
