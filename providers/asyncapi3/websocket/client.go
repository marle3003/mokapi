package websocket

import (
	"mokapi/media"

	log "github.com/sirupsen/logrus"
)

type Client struct {
	Id         string
	Query      map[string]any
	Header     map[string]any
	RemoteAddr string
	ServerAddr string

	channel *Channel
	send    chan Message
	closeCh chan struct{}
}

func (c *Client) sendMessage(message any) error {
	log.Info("sendMessage ", message)
	var err error
	var data []byte
	for _, m := range c.channel.cfg.Messages {
		if m.Value == nil || m.Value.Payload == nil || m.Value.Payload.Value == nil {
			continue
		}
		ct := media.ParseContentType(m.Value.ContentType)
		data, err = m.Value.Payload.Marshal(message, ct)
	}
	if err != nil {
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
