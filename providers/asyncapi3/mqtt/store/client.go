package store

import (
	"fmt"
	"mokapi/mqtt"
	"sync"
	"time"
)

type Client struct {
	Id           string
	Clean        bool
	Subscription map[string]Subscription

	ctx       *mqtt.ClientContext
	messageId int16
	inflight  []*InflightMessage
	m         sync.Mutex
}

type Subscription struct {
	// may contain special topic wildcard characters
	Name string
	QoS  byte
}

type InflightMessage struct {
	MessageId int16
	Message   *Message
	QoS       byte
	Retries   int
	SendAt    time.Time
}

func (c *Client) publish(msg *Message) {
	for _, sub := range c.Subscription {
		if sub.Name == msg.Topic {
			effectiveQoS := min(msg.QoS, sub.QoS)

			id := c.nextMessageId()
			if effectiveQoS > 0 {
				c.appendInflight(id, msg)
			}

			c.ctx.Send(&mqtt.Message{
				Header: &mqtt.Header{
					Type:   mqtt.PUBLISH,
					QoS:    effectiveQoS,
					Retain: false,
				},
				Payload: &mqtt.PublishRequest{
					MessageId: id,
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

func (c *Client) ResendInflight(duration time.Duration) {
	c.m.Lock()
	defer c.m.Unlock()

	now := time.Now()
	for _, inflight := range c.inflight {

		t := inflight.SendAt.Add(duration)
		if duration > 0 && t.After(now) {
			continue
		}

		fmt.Println("send")

		c.ctx.Send(&mqtt.Message{
			Header: &mqtt.Header{
				Type:   mqtt.PUBLISH,
				QoS:    inflight.QoS,
				Retain: inflight.Message.Retain,
			},
			Payload: &mqtt.PublishRequest{
				MessageId: inflight.MessageId,
				Topic:     inflight.Message.Topic,
				Data:      inflight.Message.Data,
			},
		})
	}
}

func (c *Client) appendInflight(id int16, msg *Message) {
	c.m.Lock()
	defer c.m.Unlock()

	c.inflight = append(c.inflight, &InflightMessage{
		MessageId: id,
		Message:   msg,
		SendAt:    time.Now(),
	})
}

func (c *Client) nextMessageId() int16 {
	c.m.Lock()
	defer c.m.Unlock()

	c.messageId++
	return c.messageId
}
