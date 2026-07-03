package websocket

import (
	"mokapi/media"

	log "github.com/sirupsen/logrus"
)

type Client struct {
	channel *Channel
	query   map[string]any
	header  map[string]any

	remoteAddr string
	send       chan Message
	closeCh    chan struct{}
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
