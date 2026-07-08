package websocket

import (
	"context"
	"mokapi/media"
	"mokapi/runtime/events"
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
}

func (c *Client) sendMessage(message any) error {
	var err error
	var data []byte
	var messageId string
	for id, m := range c.channel.cfg.Messages {
		if m.Value == nil || m.Value.Payload == nil || m.Value.Payload.Value == nil {
			continue
		}
		ct := media.ParseContentType(m.Value.ContentType)
		data, err = m.Value.Payload.Marshal(message, ct)
		if err == nil {
			messageId = id
			break
		}
	}

	l := messageLog(c.channel, data, messageId, c, Receive)

	channelName := c.channel.cfg.ResolveAddress()
	c.channel.log(l, events.NewTraits().With("channel", channelName).With("clientId", c.Id))

	labels := []string{c.channel.api, channelName}
	c.channel.monitor.Messages.WithLabel(labels...).Add(1)
	c.channel.monitor.LastMessage.WithLabel(labels...).Set(float64(time.Now().Unix()))

	if err != nil {
		c.channel.monitor.MessagesError.WithLabel(labels...).Add(1)
		l.Error = err.Error()
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
	for _, ch := range s.Channels {
		for _, c := range ch.clients {
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
			conn.Close(websocket.StatusNormalClosure, "server closing")
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
