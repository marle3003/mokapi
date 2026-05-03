package store

import (
	"fmt"
	"mokapi/media"
	"mokapi/providers/asyncapi3"
	"mokapi/schema/encoding"
)

type Message struct {
	Topic  string
	Data   []byte
	QoS    byte
	Retain bool
}

type Topic struct {
	Name     string
	Retained *Message

	cfg *asyncapi3.Channel
}

func (t *Topic) validate(value []byte) (messageId string, err error) {
	if t.cfg == nil {
		return
	}

	for id, msg := range t.cfg.Messages {
		if msg.Value == nil {
			continue
		}
		messageId = id
		payload := msg.Value.Payload
		if payload == nil || payload.Value == nil {
			continue
		}

		var p encoding.Parser
		p, err = payload.GetParser(msg.Value.ContentType)
		if err != nil {
			return
		}

		_, err = encoding.Decode(value, encoding.WithContentType(media.ParseContentType(msg.Value.ContentType)), encoding.WithParser(p))
		if err == nil {
			return
		}
	}
	return
}

func (s *Store) addSysTopic(name string, val string) {
	t := &Topic{
		Name: name,
		Retained: &Message{
			Topic:  name,
			Data:   []byte(val),
			Retain: true,
		},
	}
	if s.Topics == nil {
		s.Topics = map[string]*Topic{}
	}
	s.Topics[name] = t
}

func (s *Store) updateSysTopic(name string, val string) {
	t, ok := s.Topics[name]
	if !ok {
		panic(fmt.Sprintf("mqtt: sys topic '%s' not found", name))
	}
	t.Retained.Data = []byte(val)
}
