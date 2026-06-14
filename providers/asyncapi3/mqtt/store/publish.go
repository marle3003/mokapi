package store

import (
	"fmt"
	engine "mokapi/engine/common"
	"mokapi/mqtt"
	"mokapi/runtime/events"
	"time"

	log "github.com/sirupsen/logrus"
)

type PublishArgs struct {
	QoS        byte
	Retain     bool
	ClientId   string
	ScriptFile string
}

func (s *Store) Publish(publish *mqtt.PublishRequest, args PublishArgs) (mqtt.PublishReason, error) {
	msg := &Message{
		Topic:  publish.Topic,
		Data:   publish.Data,
		QoS:    args.QoS,
		Retain: args.Retain,
	}

	topic, ok := s.Topic(msg.Topic)
	if !ok {
		return mqtt.TopicNameInvalid, fmt.Errorf("topic %s not found", msg.Topic)
	}

	messageId, err := topic.validate(msg.Data)
	if err != nil {
		return mqtt.PayloadFormatInvalid, fmt.Errorf("mqtt: topic validation error '%s': %s", msg.Topic, err)
	}

	evt := &Event{
		Api:    s.cfg.Info.Name,
		Topic:  topic.Name,
		Value:  string(msg.Data),
		Retain: args.Retain,
	}
	actions := s.eventEmitter.Emit("mqtt", evt)
	if actions != nil {
		messageId, err = topic.validate(msg.Data)
		if err != nil {
			return mqtt.PayloadFormatInvalid, fmt.Errorf("mqtt: topic validation error '%s': %s", msg.Topic, err)
		}
		args.Retain = evt.Retain
	}

	if args.Retain {
		topic.Retained = msg
	}

	go func() {
		for _, client := range s.clients {
			client.publish(msg)
		}
	}()

	s.logMessage(messageId, topic, publish, actions, args)
	return mqtt.PublishSuccess, nil
}

func (s *Store) publish(rw mqtt.MessageWriter, publish *mqtt.PublishRequest, qos byte, retain bool, ctx *mqtt.ClientContext) {
	client, ok := s.clients[ctx.ClientId]
	if !ok {
		panic("unknown client")
	}
	client.Alive()

	args := PublishArgs{
		QoS:      qos,
		Retain:   retain,
		ClientId: ctx.ClientId,
	}

	reason, err := s.Publish(publish, args)

	if qos != 0 {
		res := &mqtt.PublishResponse{
			MessageId:  publish.MessageId,
			ReasonCode: reason,
		}
		if err != nil {
			res.Properties = map[byte]any{
				mqtt.ReasonString: err.Error(),
			}
		}

		_ = puback(rw, res)
	}
}

func puback(rw mqtt.MessageWriter, payload *mqtt.PublishResponse) error {
	return rw.Write(&mqtt.Message{
		Header: &mqtt.Header{
			Type: mqtt.PUBACK,
		},
		Payload: payload,
	})
}

func (s *Store) Topic(topic string) (*Topic, bool) {
	if t, ok := s.Topics[topic]; ok {
		return t, ok
	}

	for _, ch := range s.cfg.Channels {
		if ch == nil || ch.Value == nil {
			continue
		}
		if !ch.Value.IsChannelAvailable("mqtt") {
			continue
		}
		if len(ch.Value.Parameters) == 0 {
			continue
		}

		err := ch.Value.IsNameValid(topic)
		if err != nil {
			continue
		}

		if s.Topics == nil {
			s.Topics = make(map[string]*Topic)
		}

		t := &Topic{Name: topic, cfg: ch.Value}
		s.Topics[topic] = t
		return t, true
	}

	return nil, false
}

func (s *Store) logMessage(messageId string, topic *Topic, publish *mqtt.PublishRequest, actions []*engine.Action, args PublishArgs) {
	topicName := topic.cfg.ResolveAddress()
	labels := []string{s.cfg.Info.Name, topicName}

	s.monitor.Messages.WithLabel(labels...).Add(1)
	s.monitor.LastMessage.WithLabel(labels...).Set(float64(time.Now().Unix()))

	traits := events.
		NewTraits().
		WithNamespace("mqtt").
		WithName(s.cfg.Info.Name).
		With("topic", topicName).
		With("type", "message").
		With("clientId", args.ClientId)
	if len(topic.cfg.Parameters) > 0 {
		params, err := topic.cfg.ExtractParams(topic.Name)
		if err != nil {
			log.Errorf("mqtt: failed to log message: %s", err)
		}
		for k, v := range params {
			traits.With(k, v)
		}
	}

	err := s.eh.Push(&LogMessage{
		Topic:     topic.Name,
		MessageId: messageId,
		Message: LogValue{
			Value:  string(publish.Data),
			Binary: publish.Data,
		},
		Retain:   args.Retain,
		Api:      s.cfg.Info.Name,
		ClientId: args.ClientId,
		Actions:  actions,
	}, traits)
	if err != nil {
		log.Errorf("mqtt: failed to log message: %s", err)
	}
}
