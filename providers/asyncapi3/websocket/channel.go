package websocket

import (
	"context"
	"maps"
	engine "mokapi/engine/common"
	"mokapi/media"
	"mokapi/providers/asyncapi3"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"mokapi/schema/encoding"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
	log "github.com/sirupsen/logrus"
)

type Channel struct {
	Name    string
	api     string
	clients map[string]*Client
	m       sync.RWMutex
	// removed is set under m when the channel leaves the specification; clients that connect
	// afterwards, with a channel looked up before the removal, are closed on arrival.
	removed bool
	// state is replaced as a whole by Store.Update and read lock-free.
	state   atomic.Pointer[channelState]
	emitter engine.WebsocketEventEmitter
	log     func(log events.EventData, traits events.Traits)
	monitor *monitor.Websocket
}

// channelState is the part of a channel that changes with a reload. Config and params are
// swapped together because a changed address pattern can change the parameter names.
type channelState struct {
	config *asyncapi3.Channel
	params map[string]string
}

// Config returns the definition from the current specification.
func (c *Channel) Config() *asyncapi3.Channel {
	return c.state.Load().config
}

func (c *Channel) params() map[string]string {
	return c.state.Load().params
}

func (s *Store) newChannel(name string, cfg *asyncapi3.Channel, params map[string]string) *Channel {
	c := &Channel{
		Name:    name,
		api:     s.cfg.Info.Name,
		emitter: s.emitter,
		log:     s.log,
		monitor: s.monitor,
	}
	c.state.Store(&channelState{config: cfg, params: params})
	return c
}

func (s *Store) Channel(name string) (*Channel, bool) {
	s.m.RLock()
	if t, ok := s.channels[name]; ok {
		s.m.RUnlock()
		return t, ok
	}
	s.m.RUnlock()

	s.m.Lock()
	defer s.m.Unlock()

	// another request may have created the channel while waiting for the lock
	if t, ok := s.channels[name]; ok {
		return t, true
	}

	definition, params, ok := s.matchLocked(name)
	if !ok {
		return nil, false
	}
	if s.channels == nil {
		s.channels = make(map[string]*Channel)
	}
	ch := s.newChannel(name, definition, params)
	s.channels[name] = ch
	return ch, true
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
		cfg := c.Config()
		for id, m := range cfg.Messages {
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
		channelName := cfg.ResolveAddress()

		l := messageLog(c, data, messageId, client, Send)
		if err != nil {
			l.ValidationErrors = validationErrors
		}
		traits := events.NewTraits().
			With("channel", channelName).
			With("type", "message").
			With("clientId", client.Id)
		if c.Name != channelName {
			traits = traits.With("instance", c.Name)
		}

		// log event before run event engine to have the correct log order if event handler sends a message
		c.log(l, traits)

		labels := []string{c.api, channelName}
		if c.Name != channelName {
			labels = append(labels, c.Name)
		}
		c.monitor.Messages.WithLabel(labels...).Add(1)
		c.monitor.LastMessage.WithLabel(labels...).Set(float64(time.Now().Unix()))

		if err != nil {
			c.monitor.MessagesError.WithLabel(labels...).Add(1)
			return err
		}
		l.SetActions(c.runMessageEvent(client, msg))
	}
}

func (c *Channel) addClient(client *Client) {
	c.m.Lock()
	if c.clients == nil {
		c.clients = make(map[string]*Client)
	}
	c.clients[client.Id] = client
	removed := c.removed
	c.m.Unlock()
	if removed {
		client.close(websocket.StatusGoingAway, channelRemovedReason)
	}
}

func (c *Channel) removeClient(client *Client) {
	c.m.Lock()
	defer c.m.Unlock()
	delete(c.clients, client.Id)
}

// connectedClients returns a snapshot, because clients connect and disconnect concurrently
func (c *Channel) connectedClients() []*Client {
	c.m.Lock()
	defer c.m.Unlock()
	return slices.Collect(maps.Values(c.clients))
}

// closeClients closes all connected clients because the channel no longer exists.
func (c *Channel) closeClients() {
	c.m.Lock()
	c.removed = true
	clients := slices.Collect(maps.Values(c.clients))
	c.m.Unlock()
	if len(clients) == 0 {
		return
	}
	log.Infof("websocket: channel %s removed, closing %d clients", c.Name, len(clients))
	for _, client := range clients {
		client.close(websocket.StatusGoingAway, channelRemovedReason)
	}
}

func (c *Channel) Broadcast(message any) error {
	return c.BroadcastFromScript(message, "")
}

// BroadcastFromScript sends the message to all clients and records the script file in the message log.
func (c *Channel) BroadcastFromScript(message any, scriptFile string) error {
	for _, c := range c.connectedClients() {
		err := c.sendMessage(message, scriptFile)
		if err != nil {
			return err
		}
	}
	return nil
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
