package websocket

import (
	"context"
	"errors"
	"io"
	engine "mokapi/engine/common"
	"mokapi/providers/asyncapi3"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"mokapi/schema/json/parser"
	"mokapi/schema/json/schema"
	"net/http"
	"strings"
	"sync"

	"github.com/coder/websocket"
	"github.com/google/uuid"
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
	for path, ref := range cfg.Channels {
		if ref.Value == nil {
			continue
		}
		c := ref.Value
		if !c.IsChannelAvailable("ws") {
			continue
		}
		if c.Address != "" {
			path = c.Address
		}

		if c.Bindings.Websocket.Method != "" {
			if strings.ToUpper(c.Bindings.Websocket.Method) != "GET" {
				log.Warnf("channel %s: mokapi only supports WebSocket method GET, ignoring method %q", path, c.Bindings.Websocket.Method)
			}
		}

		if len(c.Parameters) == 0 {
			if s.Channels == nil {
				s.Channels = make(map[string]*Channel)
			}
			ch, ok := s.Channels[path]
			if !ok {
				ch = &Channel{
					Name:    path,
					api:     s.cfg.Info.Name,
					emitter: s.emitter,
					cfg:     c,
					log:     s.log,
					monitor: s.monitor,
				}
				s.Channels[path] = ch
			}
		}
	}
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

	header, err := parseHeader(r, ch)
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

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{})
	if err != nil {
		// Accept writes the error response itself, no need to write again
		return
	}
	defer func() { _ = conn.CloseNow() }()

	client := &Client{
		Id:         uuid.New().String(),
		Query:      query,
		Header:     header,
		RemoteAddr: r.RemoteAddr,
		ServerAddr: serverAddr,
		channel:    ch,
		send:       make(chan Message, 16),
		closeCh:    make(chan struct{}),
	}
	ch.addClient(client)
	defer ch.removeClient(client)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 2)
	go func() { errCh <- ch.readLoop(ctx, conn, client) }()
	go func() { errCh <- client.writeLoop(ctx, conn) }()

	if err = <-errCh; err != nil {
		closeStatus := websocket.CloseStatus(err)
		switch {
		case closeStatus == websocket.StatusNormalClosure,
			closeStatus == websocket.StatusGoingAway:
			// client closed cleanly, nothing to log
		case errors.Is(err, context.Canceled):
			// we canceled it ourselves (e.g. spec reload), nothing to log
		case errors.Is(err, io.EOF),
			strings.Contains(err.Error(), "EOF"):
			// client disconnected without sending a close frame
			// CloseNow() does this — it's not an error
		default:
			_ = conn.Close(websocket.StatusUnsupportedData, err.Error())
			log.Errorf("websocket connection error on channel %s: %v", ch.Name, err)
		}
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
