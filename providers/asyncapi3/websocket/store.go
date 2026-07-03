package websocket

import (
	"context"
	"errors"
	"io"
	engine "mokapi/engine/common"
	"mokapi/media"
	"mokapi/providers/asyncapi3"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"mokapi/schema/encoding"
	"mokapi/schema/json/parser"
	"mokapi/schema/json/schema"
	"net/http"
	"strings"
	"sync"

	"github.com/coder/websocket"
	log "github.com/sirupsen/logrus"
)

type Store struct {
	Channels map[string]*Channel

	cfg     *asyncapi3.Config
	emitter engine.EventEmitter
	eh      events.Handler
	m       sync.RWMutex
	monitor *monitor.Websocket
}

type Channel struct {
	api     string
	path    string
	clients map[string]*Client
	m       sync.RWMutex
	cfg     *asyncapi3.Channel
	emitter engine.EventEmitter
	log     func(log *Log, traits events.Traits)
	monitor *monitor.Websocket
}

type Message struct {
	Type    MessageType
	Payload []byte
}

type MessageType uint8

const (
	MessageTypeText MessageType = iota
	MessageTypeBinary
)

func New(cfg *asyncapi3.Config, emitter engine.EventEmitter, eh events.Handler, m *monitor.Websocket) *Store {
	s := &Store{
		cfg:     cfg,
		emitter: emitter,
		eh:      eh,
		monitor: m,
	}
	s.Update(cfg)
	return s
}

func (s *Store) Update(cfg *asyncapi3.Config) {
	s.cfg = cfg
	for path, c := range cfg.Channels {
		if c.Value == nil {
			continue
		}
		if !c.Value.IsChannelAvailable("ws") {
			continue
		}
		if c.Value.Address != "" {
			path = c.Value.Address
		}

		if c.Value.Bindings.Websocket.Method != "" {
			if strings.ToUpper(c.Value.Bindings.Websocket.Method) != "GET" {
				log.Warnf("channel %s: mokapi only supports WebSocket method GET, ignoring method %q", path, c.Value.Bindings.Websocket.Method)
			}
		}

		if s.Channels == nil {
			s.Channels = make(map[string]*Channel)
		}
		ch, ok := s.Channels[path]
		if !ok {
			ch = &Channel{
				api:     s.cfg.Info.Name,
				path:    path,
				emitter: s.emitter,
				cfg:     c.Value,
				log:     s.log,
				monitor: s.monitor,
			}
			s.Channels[path] = ch
		}
	}
}

func (s *Store) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.m.RLock()
	ch, ok := s.Channels[r.URL.Path]
	s.m.RUnlock()
	if !ok {
		http.NotFound(w, r)
		return
	}

	query, err := parseQuery(r, ch)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	header, err := parseHeader(r, ch)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{})
	if err != nil {
		// Accept writes the error response itself, no need to write again
		return
	}
	defer func() { _ = conn.CloseNow() }()

	client := &Client{
		channel:    ch,
		query:      query,
		header:     header,
		remoteAddr: r.RemoteAddr,
		send:       make(chan Message, 16),
		closeCh:    make(chan struct{}),
	}
	ch.addClient(client)
	defer ch.removeClient(client)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 2)
	go func() { errCh <- ch.readLoop(ctx, conn, client, ch) }()
	go func() { errCh <- s.writeLoop(ctx, conn, client) }()

	if err = <-errCh; err != nil {
		closeStatus := websocket.CloseStatus(err)
		switch {
		case closeStatus == websocket.StatusNormalClosure,
			closeStatus == websocket.StatusGoingAway:
			// client closed cleanly, nothing to log
		case errors.Is(err, context.Canceled):
			// we cancelled it ourselves (e.g. spec reload), nothing to log
		case errors.Is(err, io.EOF),
			strings.Contains(err.Error(), "EOF"):
			// client disconnected without sending a close frame
			// CloseNow() does this — it's not an error
		default:
			_ = conn.Close(websocket.StatusUnsupportedData, err.Error())
			log.Errorf("websocket connection error on channel %s: %v", ch.path, err)
		}
	}
}

func (c *Channel) readLoop(ctx context.Context, conn *websocket.Conn, client *Client, ch *Channel) error {
	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			return err
		}

		var v any
		for _, m := range c.cfg.Messages {
			if m.Value == nil || m.Value.Payload == nil || m.Value.Payload.Value == nil {
				continue
			}
			var p encoding.Parser
			p, err = m.Value.Payload.GetParser(m.Value.ContentType)
			if err != nil {
				log.Errorf("unsupported payload type: %T", m.Value.Payload.Value)
			}
			v, err = encoding.Decode(data, encoding.WithContentType(media.ParseContentType(m.Value.ContentType)), encoding.WithParser(p))
		}
		if err != nil {
			return err
		}

		evt := &Event{
			Api: c.api,
			Channel: EventChannel{
				Name: ch.path,
				ch:   ch,
			},
			Client: &EventClient{
				RemoteAddress: client.remoteAddr,
				Headers:       nil,
				client:        client,
			},
			Message: v,
		}
		if client.query != nil {
			evt.Client.Query = client.query
		} else {
			evt.Client.Query = map[string]any{}
		}
		if client.header != nil {
			evt.Client.Headers = client.header
		} else {
			evt.Client.Headers = map[string]any{}
		}

		l := &Log{}
		l.Actions = c.emitter.Emit("websocket", evt)
		c.log(l, events.NewTraits().With("channel", ch.path))
	}
}

func (s *Store) writeLoop(ctx context.Context, conn *websocket.Conn, client *Client) error {
	for {
		select {
		case msg := <-client.send:
			wsType := toWebSocketType(msg.Type)
			if err := conn.Write(ctx, wsType, msg.Payload); err != nil {
				return err
			}
		case <-client.closeCh:
			conn.Close(websocket.StatusNormalClosure, "server closing")
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (c *Channel) addClient(client *Client) {
	c.m.Lock()
	if c.clients == nil {
		c.clients = make(map[string]*Client)
	}
	c.clients[client.remoteAddr] = client
	c.m.Unlock()
}

func (c *Channel) removeClient(client *Client) {
	c.m.Lock()
	defer c.m.Unlock()
	delete(c.clients, client.remoteAddr)
}

func toMessageType(msgType websocket.MessageType) MessageType {
	switch msgType {
	case websocket.MessageText:
		return MessageTypeText
	case websocket.MessageBinary:
		return MessageTypeBinary
	default:
		panic("unknown MessageType")
	}
}

func toWebSocketType(msgType MessageType) websocket.MessageType {
	switch msgType {
	case MessageTypeText:
		return websocket.MessageText
	case MessageTypeBinary:
		return websocket.MessageBinary
	default:
		panic("unknown MessageType")
	}
}

func (s *Store) log(log *Log, traits events.Traits) {
	log.Api = s.cfg.Info.Name
	t := traits.WithNamespace("websocket").
		WithName(s.cfg.Info.Name)
	_ = s.eh.Push(log, t)
}

func parseQuery(r *http.Request, ch *Channel) (map[string]any, error) {
	s := ch.cfg.Bindings.Websocket.Query
	if s == nil {
		return map[string]any{}, nil
	}
	raw := map[string]string{}

	// AsyncAPI spec definition:
	// A Schema object containing the definitions for each query parameter.
	// This schema MUST be of type object and have a properties key.
	for name, values := range r.URL.Query() {
		p := s.Properties.Get(name)
		if p == nil {
			continue
		}
		val := strings.Join(values, ",")
		raw[name] = val
	}

	p := parser.Parser{Schema: s, ConvertStringToNumber: true, ConvertStringToBoolean: true}
	v, err := p.Parse(raw)
	if err != nil {
		return nil, err
	}
	return v.(map[string]any), nil
}

func parseHeader(r *http.Request, ch *Channel) (map[string]any, error) {
	s := ch.cfg.Bindings.Websocket.Headers
	if s == nil {
		return map[string]any{}, nil
	}
	raw := map[string]string{}

	// AsyncAPI spec definition:
	// A Schema object containing the definitions for each query parameter.
	// This schema MUST be of type object and have a properties key.
	for name, values := range r.Header {
		// RFC7230 states header names are case-insensitive.
		key, p := getPropertyIgnoreCase(name, s)
		if p == nil {
			continue
		}
		val := strings.Join(values, ",")
		// we use the name used in the specification
		raw[key] = val
	}

	p := parser.Parser{Schema: s, ConvertStringToNumber: true, ConvertStringToBoolean: true}
	v, err := p.Parse(raw)
	if err != nil {
		return nil, err
	}
	return v.(map[string]any), nil
}

func getPropertyIgnoreCase(name string, s *schema.Schema) (string, *schema.Schema) {
	name = strings.ToLower(name)
	for it := s.Properties.Iter(); it.Next(); {
		key := strings.ToLower(it.Key())
		if key == name {
			return it.Key(), it.Value()
		}
	}
	return "", nil
}
