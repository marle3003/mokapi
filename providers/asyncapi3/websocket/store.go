package websocket

import (
	"context"
	"errors"
	"io"
	"maps"
	engine "mokapi/engine/common"
	"mokapi/providers/asyncapi3"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"mokapi/schema/json/parser"
	"mokapi/schema/json/schema"
	"net/http"
	"slices"
	"strings"
	"sync"

	"github.com/coder/websocket"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type Store struct {
	channels map[string]*Channel

	cfg     *asyncapi3.Config
	emitter engine.WebsocketEventEmitter
	eh      events.Handler
	m       sync.RWMutex
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

func New(cfg *asyncapi3.Config, emitter engine.WebsocketEventEmitter, eh events.Handler, m *monitor.Websocket) *Store {
	s := &Store{
		cfg:     cfg,
		emitter: emitter,
		eh:      eh,
		monitor: m,
	}
	s.Update(cfg)
	return s
}

// Channels returns a snapshot of the channels created so far.
func (s *Store) Channels() []*Channel {
	s.m.RLock()
	defer s.m.RUnlock()
	return slices.Collect(maps.Values(s.channels))
}

const channelRemovedReason = "channel removed"

// Update applies a reloaded specification. Channels keep their identity because connected clients
// and scripts hold them: a channel that still exists gets the new definition, one that is gone from
// the specification is removed and its clients are closed.
func (s *Store) Update(cfg *asyncapi3.Config) {
	// Clients are closed after the lock is released: closing logs through Store.log, which needs s.m.
	s.m.Lock()
	afterUnlock := s.updateLocked(cfg)
	s.m.Unlock()

	for _, f := range afterUnlock {
		f()
	}
}

func (s *Store) updateLocked(cfg *asyncapi3.Config) (afterUnlock []func()) {
	s.cfg = cfg

	for name, ch := range s.channels {
		definition, params, ok := s.matchLocked(name)
		if ok {
			ch.state.Store(&channelState{config: definition, params: params})
			continue
		}
		delete(s.channels, name)
		afterUnlock = append(afterUnlock, func() { ch.closeClients() })
	}

	for key, ref := range cfg.Channels {
		if ref == nil || ref.Value == nil {
			continue
		}
		c := ref.Value
		if !c.IsChannelAvailable("ws") {
			continue
		}
		path := staticPath(key, c)

		if c.Bindings != nil && c.Bindings.Websocket.Method != "" {
			if strings.ToUpper(c.Bindings.Websocket.Method) != "GET" {
				log.Warnf("channel %s: mokapi only supports WebSocket method GET, ignoring method %q", path, c.Bindings.Websocket.Method)
			}
		}

		if len(c.Parameters) == 0 {
			if s.channels == nil {
				s.channels = make(map[string]*Channel)
			}
			if _, ok := s.channels[path]; !ok {
				s.channels[path] = s.newChannel(path, c, nil)
			}
		}
	}
	return afterUnlock
}

// staticPath returns the path a channel without parameters is served at.
func staticPath(key string, c *asyncapi3.Channel) string {
	if c.Address != "" {
		return c.Address
	}
	return key
}

// matchLocked finds the channel definition that serves the path in the current specification:
// a channel without parameters by its path, otherwise a parameterized channel whose address
// pattern matches.
func (s *Store) matchLocked(path string) (*asyncapi3.Channel, map[string]string, bool) {
	var parameterized []*asyncapi3.Channel
	for key, ref := range s.cfg.Channels {
		if ref == nil || ref.Value == nil || !ref.Value.IsChannelAvailable("ws") {
			continue
		}
		if len(ref.Value.Parameters) > 0 {
			parameterized = append(parameterized, ref.Value)
		} else if staticPath(key, ref.Value) == path {
			return ref.Value, nil, true
		}
	}
	for _, c := range parameterized {
		if params, err := c.ExtractParams(path); err == nil {
			return c, params, true
		}
	}
	return nil, nil, false
}

func (s *Store) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ch, ok := s.Channel(r.URL.Path)
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

	headers, err := parseHeader(r, ch)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	server, ok := ServerFromContext(r.Context())
	serverAddr := ""
	if ok {
		serverAddr = server.Host
	}

	client := &Client{
		Id:         uuid.New().String(),
		Query:      query,
		Headers:    headers,
		RemoteAddr: r.RemoteAddr,
		ServerAddr: serverAddr,
		channel:    ch,
		send:       make(chan Message, 16),
		closeCh:    make(chan struct{}),
		closing:    &closeRequest{},
	}
	// Register before Accept: the client's Dial returns with the handshake response, so a
	// message sent right after it must already find the client. It is buffered in client.send
	// until writeLoop starts.
	ch.addClient(client)
	defer ch.removeClient(client)

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		// Accept writes the error response itself, no need to write again
		log.Errorf("websocket: %s", err)
		return
	}
	defer func() { _ = conn.CloseNow() }()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	l := connectLog(ch, client)
	// log event before run event engine to have the correct log order if event handler sends a message
	s.log(l, events.NewTraits().
		With("channel", ch.Name).
		With("type", "connection").
		With("clientId", client.Id))
	l.SetActions(ch.runConnectEvent(client))

	errCh := make(chan error, 2)
	go func() { errCh <- ch.readLoop(ctx, conn, client) }()
	go func() { errCh <- client.writeLoop(ctx, conn) }()

	if err = <-errCh; err != nil {
		status, reason := closeStatus(err)
		var closedBy string
		serverReason, closedByServer := client.serverCloseReason()
		switch {
		case closedByServer:
			// checked first: the peer answers our close frame, so the read loop may report the error first
			reason = serverReason
			closedBy = "server"
		case status == websocket.StatusNormalClosure,
			status == websocket.StatusGoingAway,
			status == websocket.StatusNoStatusRcvd:
			// client closed
			closedBy = "client"
		case errors.Is(err, io.EOF),
			strings.Contains(err.Error(), "EOF"):
			// client disconnected without sending a close frame
			// CloseNow() does this — it's not an error
			reason = "connection disconnected"
			closedBy = "client"
		default:
			_ = conn.Close(websocket.StatusUnsupportedData, err.Error())
			log.Errorf("websocket connection error on channel %s: %v", ch.Name, err)
			reason = err.Error()
			closedBy = "server"
		}

		cl := closeLog(ch, client, reason, closedBy)
		cl.Actions = ch.runCloseEvent(client, reason, closedBy)
		s.log(cl, events.NewTraits().
			With("channel", ch.Name).
			With("type", "connection").
			With("clientId", client.Id))
	}
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

func (s *Store) log(log events.EventData, traits events.Traits) {
	s.m.RLock()
	name := s.cfg.Info.Name
	s.m.RUnlock()

	t := traits.WithNamespace("websocket").
		WithName(name)
	_ = s.eh.Push(log, t)
}

func parseQuery(r *http.Request, ch *Channel) (map[string]any, error) {
	cfg := ch.Config()
	if cfg.Bindings == nil {
		return map[string]any{}, nil
	}
	s := cfg.Bindings.Websocket.Query
	if s == nil {
		return map[string]any{}, nil
	}
	raw := map[string]string{}

	// AsyncAPI spec definition:
	// A Schema object containing the definitions for each query parameter.
	// This schema MUST be of the type object and have a property key.
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
	cfg := ch.Config()
	if cfg.Bindings == nil {
		return map[string]any{}, nil
	}
	s := cfg.Bindings.Websocket.Headers
	if s == nil {
		return map[string]any{}, nil
	}
	raw := map[string]string{}

	// AsyncAPI spec definition:
	// A Schema object containing the definitions for each query parameter.
	// This schema MUST be of the type object and have a property key.
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

func closeStatus(err error) (websocket.StatusCode, string) {
	if closeErr, ok := errors.AsType[websocket.CloseError](err); ok {
		return closeErr.Code, closeErr.Reason
	}
	return -1, ""
}
