package mqtt

import (
	"context"
	"fmt"
	"io"
	"mokapi/buffer"
)

type Message struct {
	Header  *Header
	Payload Payload
	Context context.Context
}

type Payload interface {
	Write(e *Encoder)
	Read(*Decoder)
}

type MessageOptions func(*Message)

func NewMessage(payload Payload, opts ...MessageOptions) *Message {
	m := &Message{
		Header:  &Header{},
		Payload: payload,
	}

	for _, opt := range opts {
		opt(m)
	}
	return m
}

func (m *Message) WithContext(ctx context.Context) *Message {
	m.Context = ctx
	return m
}

func (m *Message) Write(w io.Writer) error {
	b := buffer.NewPageBuffer()
	defer b.Unref()

	e := NewEncoder(b)
	m.Payload.Write(e)

	m.Header.Size = b.Size()
	err := m.Header.Write(w)
	if err != nil {
		return err
	}

	_, err = b.WriteTo(w)
	return err
}

func (m *Message) Read(reader io.Reader) error {
	d := NewDecoder(reader, 5)
	m.Header = readHeader(d)
	if d.err != nil {
		return d.err
	}
	d.leftSize = m.Header.Size

	switch m.Header.Type {
	case CONNECT:
		m.Payload = &ConnectRequest{}
	case CONNACK:
		m.Payload = &ConnectResponse{}
	case PUBLISH:
		m.Payload = &PublishRequest{}
	case PUBACK:
		m.Payload = &PublishResponse{}
	case SUBSCRIBE:
		m.Payload = &SubscribeRequest{}
	case SUBACK:
		m.Payload = &SubscribeResponse{}
	case UNSUBSCRIBE:
		m.Payload = &UnsubscribeRequest{}
	case UNSUBACK:
		m.Payload = &UnsubscribeResponse{}
	default:
		return fmt.Errorf("unknown MQTT protocol type %d", m.Header.Type)
	}

	m.Payload.Read(d)

	if d.err != nil {
		return d.err
	}

	if d.leftSize > 0 {
		return fmt.Errorf("mqtt: remaining length %d is not zero", d.leftSize)
	}

	return nil
}

func WithDup() MessageOptions {
	return func(m *Message) {
		m.Header.Dup = true
	}
}
