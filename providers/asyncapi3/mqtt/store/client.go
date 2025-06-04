package store

import (
	"math"
	"mokapi/mqtt"
)

type Client struct {
	Id           string
	Clean        bool
	Subscription map[string]Subscription
	ctx          *mqtt.ClientContext
}

type Subscription struct {
	// may contain special topic wildcard characters
	Name string
	QoS  byte
}

func (c *Client) publish(msg *Message) {
	for _, sub := range c.Subscription {
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

func (c *Client) Subscribe(topic string, qos byte) {
	if c.Subscription == nil {
		c.Subscription = map[string]Subscription{}
	}

	c.Subscription[topic] = Subscription{
		Name: topic,
		QoS:  qos,
	}
}
