package store

import (
	"mokapi/mqtt"
	"mokapi/runtime/events"
	"time"

	log "github.com/sirupsen/logrus"
)

func (s *Store) publish(rw mqtt.MessageWriter, publish *mqtt.PublishRequest, qos byte, retain bool, ctx *mqtt.ClientContext) {
	client, ok := s.clients[ctx.ClientId]
	if !ok {
		panic("client not found")
	}
	client.Alive()

	msg := &Message{
		Topic:  publish.Topic,
		Data:   publish.Data,
		QoS:    qos,
		Retain: retain,
	}

	topic, ok := s.getTopic(msg.Topic)
	if !ok {
		log.Infof("mqtt: topic not specified %s", msg.Topic)
		if qos != 0 {
			puback(rw, &mqtt.PublishResponse{
				MessageId:  publish.MessageId,
				ReasonCode: mqtt.TopicNameInvalid,
			})
		}
		return
	}

	messageId, err := topic.validate(msg.Data)
	if err != nil {
		log.Errorf("mqtt: topic validation error '%s': %s", msg.Topic, err)
		puback(rw, &mqtt.PublishResponse{
			MessageId:  publish.MessageId,
			ReasonCode: mqtt.PayloadFormatInvalid,
		})
		return
	}

	if retain {
		topic.Retained = msg
	}

	if qos == 1 {
		puback(rw, &mqtt.PublishResponse{
			MessageId: publish.MessageId,
		})
	}

	go func() {
		for _, client := range s.clients {
			client.publish(msg)
		}
	}()

	s.logMessage(messageId, topic, publish, ctx)
}

func puback(rw mqtt.MessageWriter, payload *mqtt.PublishResponse) error {
	return rw.Write(&mqtt.Message{
		Header: &mqtt.Header{
			Type: mqtt.PUBACK,
		},
		Payload: payload,
	})
}

func (s *Store) getTopic(topic string) (*Topic, bool) {
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

func (s *Store) logMessage(messageId string, topic *Topic, publish *mqtt.PublishRequest, ctx *mqtt.ClientContext) {
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
		With("clientId", ctx.ClientId)
	if len(topic.cfg.Parameters) > 0 {
		params, err := topic.cfg.ExtractParams(topic.Name)
		if err != nil {
			log.Errorf("mqtt: failed to log message: %s", err)
		}
		for k, v := range params {
			traits.With(k, v)
		}
	}

	client, _ := s.clients[ctx.ClientId]
	err := s.eh.Push(&LogMessage{
		Topic:     topic.Name,
		MessageId: messageId,
		Message: LogValue{
			Value:  string(publish.Data),
			Binary: publish.Data,
		},
		Api:      s.cfg.Info.Name,
		ClientId: client.Id,
	}, traits)
	if err != nil {
		log.Errorf("mqtt: failed to log message: %s", err)
	}
}
