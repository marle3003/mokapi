package store

import (
	"mokapi/mqtt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type ClientState uint8

const (
	ClientConnected ClientState = iota
	ClientDisconnected
)

type Client struct {
	Id                    string
	Clean                 bool
	SessionExpiryInterval int32
	KeepAlive             int16

	// m guards everything below, publish requests reach a client from many connections
	m            sync.Mutex
	subscription map[string]Subscription
	willMessage  *Message
	lastSeen     time.Time
	state        ClientState
	ctx          *mqtt.ClientContext
	messageId    uint16
	inflight     []*InflightMessage
}

type Subscription struct {
	// may contain special topic wildcard characters
	Name string
	QoS  byte
}

type InflightMessage struct {
	MessageId uint16
	Message   *Message
	QoS       byte
	Retries   int
	SendAt    time.Time
}

func (c *Client) publish(msg *Message) {
	ctx := c.context()
	for _, sub := range c.subscriptions() {
		if topicMatches(sub.Name, msg.Topic) {
			effectiveQoS := min(msg.QoS, sub.QoS)

			id := uint16(0)
			if effectiveQoS > 0 {
				id = c.nextMessageId()
				c.appendInflight(id, msg)
			}

			err := ctx.Send(&mqtt.Message{
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
			if err != nil {
				log.Errorf("mqtt: failed to publish msg %d: %v", id, err)
			}
		}
	}
}

func (c *Client) Subscribe(topic string, qos byte) {
	c.m.Lock()
	defer c.m.Unlock()

	if c.subscription == nil {
		c.subscription = map[string]Subscription{}
	}

	c.subscription[topic] = Subscription{
		Name: topic,
		QoS:  qos,
	}
}

// Unsubscribe reports whether the client was subscribed to topic.
func (c *Client) Unsubscribe(topic string) bool {
	c.m.Lock()
	defer c.m.Unlock()

	_, ok := c.subscription[topic]
	delete(c.subscription, topic)
	return ok
}

func (c *Client) subscriptions() []Subscription {
	c.m.Lock()
	defer c.m.Unlock()

	subs := make([]Subscription, 0, len(c.subscription))
	for _, sub := range c.subscription {
		subs = append(subs, sub)
	}
	return subs
}

// Will returns the will message set on connect.
func (c *Client) Will() *Message {
	c.m.Lock()
	defer c.m.Unlock()

	return c.willMessage
}

func (c *Client) SetWill(msg *Message) {
	c.m.Lock()
	defer c.m.Unlock()

	c.willMessage = msg
}

// attach binds the client to the connection it (re)connected with.
func (c *Client) attach(ctx *mqtt.ClientContext, keepAlive int16) {
	c.m.Lock()
	defer c.m.Unlock()

	c.ctx = ctx
	c.KeepAlive = keepAlive
	c.lastSeen = time.Now()
}

func (c *Client) checkSession(now time.Time) {
	c.m.Lock()
	defer c.m.Unlock()

	if now.After(c.lastSeen.Add(time.Duration(c.KeepAlive) * time.Second)) {
		c.state = ClientConnected
	}
}

func (c *Client) context() *mqtt.ClientContext {
	c.m.Lock()
	defer c.m.Unlock()

	return c.ctx
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

func (c *Client) appendInflight(id uint16, msg *Message) {
	c.m.Lock()
	defer c.m.Unlock()

	c.inflight = append(c.inflight, &InflightMessage{
		QoS:       msg.QoS,
		MessageId: id,
		Message:   msg,
		SendAt:    time.Now(),
	})
}

func (c *Client) nextMessageId() uint16 {
	c.m.Lock()
	defer c.m.Unlock()

	c.messageId++
	return c.messageId
}

func (c *Client) Addr() string {
	return c.context().Addr
}

func (c *Client) ServerAddress() string {
	return c.context().ServerAddress
}

func (c *Client) ProtocolVersion() byte {
	return c.context().ProtocolVersion
}

func (c *Client) Alive() {
	c.m.Lock()
	defer c.m.Unlock()

	c.lastSeen = time.Now()
}
