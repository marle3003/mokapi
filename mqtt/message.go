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
	Write(e *Encoder, h *Header)
	Read(d *Decoder, h *Header)
}

type MessageOptions func(*Message)

func (m *Message) WithContext(ctx context.Context) *Message {
	m.Context = ctx
	return m
}

func (m *Message) Write(w io.Writer, ctx *ClientContext) error {
	b := buffer.NewPageBuffer()
	defer b.Unref()

	e := NewEncoder(b, ctx.ProtocolVersion)
	if m.Payload == nil {
		return fmt.Errorf("mqtt: message has no payload")
	}
	m.Payload.Write(e, m.Header)

	m.Header.Size = b.Size()
	err := m.Header.Write(w)
	if err != nil {
		return err
	}

	_, err = b.WriteTo(w)
	return err
}

func (m *Message) Read(reader io.Reader, ctx *ClientContext) error {
	d := NewDecoder(reader, 5, ctx.ProtocolVersion)
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
	case PINGREQ:
		m.Payload = &PingRequest{}
	case PINGRESP:
		m.Payload = &PingResponse{}
	case DISCONNECT:
		m.Payload = &DisconnectRequest{}
	default:
		return fmt.Errorf("unknown MQTT protocol type %d", m.Header.Type)
	}

	m.Payload.Read(d, m.Header)

	if d.err != nil {
		return d.err
	}

	if d.leftSize > 0 {
		return fmt.Errorf("mqtt: remaining length %d is not zero", d.leftSize)
	}

	if m.Header.Type == CONNECT {
		c := m.Payload.(*ConnectRequest)
		ctx.ProtocolVersion = c.Version
	}

	return nil
}
