package store

import (
	"fmt"
	"mokapi/providers/asyncapi3"
	"mokapi/schema/encoding"

	log "github.com/sirupsen/logrus"
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

func (t *Topic) validate(value []byte) (err error) {
	if t.cfg == nil {
		return
	}

	for _, msg := range t.cfg.Messages {
		if msg.Value == nil {
			continue
		}
		payload := msg.Value.Payload
		if payload == nil || payload.Value == nil {
			continue
		}

		var p encoding.Parser
		p, err = payload.GetParser(msg.Value.ContentType)
		if err != nil {
			return err
		}
		var v any
		v, err = p.Parse(value)
		log.Infof("%v", v)
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
