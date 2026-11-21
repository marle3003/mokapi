package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"mokapi/media"
	"mokapi/runtime/events"
	"sync"
	"time"

	"github.com/coder/websocket"
)

type Client struct {
	Id         string
	Query      map[string]any
	Headers    map[string]any
	RemoteAddr string
	ServerAddr string

	channel *Channel
	send    chan Message
	closeCh chan struct{}
	// shared by copies of the client, see Store.Clients
	closing *closeRequest
}

// closeRequest is written once by Client.close before closeCh is closed, so everyone who has
// seen closeCh closed sees code and reason.
type closeRequest struct {
	once   sync.Once
	code   websocket.StatusCode
	reason string
}

// errClosedByServer ends writeLoop after the server closed the connection on its own initiative.
var errClosedByServer = errors.New("closed by server")

// close asks the connection to close with the given status. It is safe to call concurrently
// and repeatedly, only the first call counts.
func (c *Client) close(code websocket.StatusCode, reason string) {
	c.closing.once.Do(func() {
		c.closing.code = code
		c.closing.reason = reason
		close(c.closeCh)
	})
}

// serverCloseReason returns the reason if the server closed the connection.
func (c *Client) serverCloseReason() (string, bool) {
	select {
	case <-c.closeCh:
		return c.closing.reason, true
	default:
		return "", false
	}
}

func (c *Client) Send(message any) {
	err := c.sendMessage(message, "")
	if err != nil {
		panic(err)
	}
}

func (c *Client) sendMessage(message any, scriptFile string) error {
	var err error
	var data []byte
	var messageId string
	validationErrors := make(map[string]string)
	cfg := c.channel.Config()
	for id, m := range cfg.Messages {
		if m.Value == nil || m.Value.Payload == nil || m.Value.Payload.Value == nil {
			continue
		}
		ct := media.ParseContentType(m.Value.ContentType)
		data, err = m.Value.Payload.Marshal(message, ct)
		if err == nil {
			messageId = id
			break
		}
		validationErrors[id] = err.Error()
	}

	l := messageLog(c.channel, data, messageId, c, Receive)
	l.Script = scriptFile
	if err != nil {
		b, _ := json.Marshal(message)
		l.Message = &LogValue{
			Value:  string(b),
			Binary: b,
		}
		l.ValidationErrors = validationErrors
	}

	channelName := cfg.ResolveAddress()
	traits := events.NewTraits().
		With("channel", channelName).
		With("type", "message").
		With("clientId", c.Id)
	if c.channel.Name != channelName {
		traits = traits.With("instance", c.channel.Name)
	}

	c.channel.log(l, traits)

	labels := []string{c.channel.api, channelName}
	if c.channel.Name != channelName {
		labels = append(labels, c.channel.Name)
	}
	c.channel.monitor.Messages.WithLabel(labels...).Add(1)
	c.channel.monitor.LastMessage.WithLabel(labels...).Set(float64(time.Now().Unix()))

	if err != nil {
		c.channel.monitor.MessagesError.WithLabel(labels...).Add(1)
		return err
	}

	c.send <- Message{
		Type:    MessageTypeText,
		Payload: data,
	}

	return nil
}

func (s *Store) Clients() []Client {
	var clients []Client
	for _, ch := range s.Channels() {
		for _, c := range ch.connectedClients() {
			clients = append(clients, *c)
		}
	}
	return clients
}

func (c *Client) writeLoop(ctx context.Context, conn *websocket.Conn) error {
	for {
		select {
		case msg := <-c.send:
			wsType := toWebSocketType(msg.Type)
			if err := conn.Write(ctx, wsType, msg.Payload); err != nil {
				return err
			}
		case <-c.closeCh:
			_ = conn.Close(c.closing.code, c.closing.reason)
			return errClosedByServer
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
