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
	emitter engine.WebsocketEventEmitter
	log     func(log events.EventData, traits events.Traits)
	monitor *monitor.Websocket
	params  map[string]string
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

		params, err := ch.ExtractParams(name)
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
			params:  params,
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
		validationErrors := make(map[string]string)
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
			validationErrors[id] = err.Error()
		}
		channelName := c.cfg.ResolveAddress()

		l := messageLog(c, data, messageId, client, Send)
		// log event before run event engine to have the correct log order if event handler sends a message
		c.log(l, events.NewTraits().
			With("channel", channelName).
			With("type", "message").
			With("clientId", client.Id))

		labels := []string{c.api, channelName}
		c.monitor.Messages.WithLabel(labels...).Add(1)
		c.monitor.LastMessage.WithLabel(labels...).Set(float64(time.Now().Unix()))

		if err != nil {
			c.monitor.MessagesError.WithLabel(labels...).Add(1)
			l.ValidationErrors = validationErrors
			return err
		}
		l.Actions = c.runMessageEvent(client, msg)
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

func (c *Channel) Broadcast(message any) {
	for _, c := range c.clients {
		err := c.sendMessage(message)
		if err != nil {
			panic(err)
		}
	}
}

func (c *Channel) newConnectEvent(client *Client) *engine.WebsocketConnectEvent {
	evt := &engine.WebsocketConnectEvent{
		WebsocketEvent: engine.WebsocketEvent{
			Type:    engine.WebsocketConnectEventType,
			Api:     c.api,
			Channel: newEventChannel(c),
			Client:  newEventClient(client),
		},
		Conn: client,
	}
	return evt
}

func (c *Channel) newMessageEvent(client *Client, v any) *engine.WebsocketMessageEvent {
	evt := &engine.WebsocketMessageEvent{
		WebsocketEvent: engine.WebsocketEvent{
			Type:    engine.WebsocketMessageEventType,
			Api:     c.api,
			Channel: newEventChannel(c),
			Client:  newEventClient(client),
		},
		Message: v,
		Conn:    client,
	}
	return evt
}

func (c *Channel) newCloseEvent(client *Client, reason, closedBy string) *engine.WebsocketCloseEvent {
	evt := &engine.WebsocketCloseEvent{
		WebsocketEvent: engine.WebsocketEvent{
			Type:    engine.WebsocketCloseEventType,
			Api:     c.api,
			Channel: newEventChannel(c),
			Client:  newEventClient(client),
		},
		Reason:   reason,
		ClosedBy: closedBy,
		Conn:     client,
	}
	return evt
}

func (c *Channel) runMessageEvent(client *Client, msg any) []*engine.Action {
	evt := c.newMessageEvent(client, msg)
	return c.emitter.EmitWebsocketMessage(evt)
}

func (c *Channel) runConnectEvent(client *Client) []*engine.Action {
	evt := c.newConnectEvent(client)
	return c.emitter.EmitWebsocketConnect(evt)
}

func (c *Channel) runCloseEvent(client *Client, reason, closedBy string) []*engine.Action {
	evt := c.newCloseEvent(client, reason, closedBy)
	return c.emitter.EmitWebsocketClose(evt)
}
