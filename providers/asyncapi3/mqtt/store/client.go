package store

import (
	"math"
	"mokapi/mqtt"
)

type Client struct {
	Id     string
	Clean  bool
	Topics map[string]*SubscribedTopic
	ctx    *mqtt.ClientContext
}

type SubscribedTopic struct {
	// may contain special topic wildcard characters
	Name string
	QoS  byte
}

func (c *Client) publish(msg *Message) {
	for _, sub := range c.Topics {
		if sub.Name == msg.Topic {
			c.ctx.Send(&mqtt.Request{
				Header: &mqtt.Header{
					Type:   mqtt.PUBLISH,
					QoS:    byte(math.Min(float64(msg.QoS), float64(sub.QoS))),
					Retain: false,
				},
				Message: &mqtt.PublishRequest{
					MessageId: c.ctx.NextMessageId(),
					Topic:     msg.Topic,
					Data:      msg.Data,
				},
			})
		}
	}
}
