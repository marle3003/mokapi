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
	QoS            byte
	Retain         bool
	ClientId       string
	ScriptFile     string
	SkipValidation bool
}

// Publish validates and emits the message to the JavaScript handlers before it
// is retained and delivered. A handler may publish again (mqtt.publish), so
// s.m is never held while the emitter runs.
func (s *Store) Publish(publish *mqtt.PublishRequest, args PublishArgs) (mqtt.PublishReason, error) {
	msg := &Message{
		Topic:  publish.Topic,
		Data:   publish.Data,
		QoS:    args.QoS,
		Retain: args.Retain,
	}

	s.m.Lock()
	topic, ok := s.topic(msg.Topic)
	apiName := s.cfg.Info.Name
	s.m.Unlock()
	if !ok {
		return mqtt.TopicNameInvalid, fmt.Errorf("topic %s not found", msg.Topic)
	}

	var messageId string
	var err error
	if !args.SkipValidation {
		messageId, err = topic.validate(msg.Data)
		if err != nil {
			return mqtt.PayloadFormatInvalid, fmt.Errorf("mqtt: topic validation error '%s': %s", msg.Topic, err)
		}
	}

	evt := &engine.MqttMessageEvent{
		Api:    apiName,
		Topic:  topic.Name,
		Value:  string(msg.Data),
		Retain: args.Retain,
	}
	actions := s.eventEmitter.EmitMqtt(evt)
	if actions != nil {
		if !args.SkipValidation {
			messageId, err = topic.validate(msg.Data)
			if err != nil {
				return mqtt.PayloadFormatInvalid, fmt.Errorf("mqtt: topic validation error '%s': %s", msg.Topic, err)
			}
		}
		args.Retain = evt.Retain
	}

	s.m.Lock()
	if args.Retain {
		// Update may have replaced the topic while the emitter ran
		if current, ok := s.topic(msg.Topic); ok {
			current.Retained = msg
		}
	}
	clients := s.clientList()
	s.m.Unlock()

	go func() {
		for _, client := range clients {
			client.publish(msg)
		}
	}()

	s.logMessage(apiName, messageId, topic, publish, actions, args)
	return mqtt.PublishSuccess, nil
}

func (s *Store) publish(rw mqtt.MessageWriter, publish *mqtt.PublishRequest, qos byte, retain bool, ctx *mqtt.ClientContext) {
	s.m.RLock()
	client, ok := s.clients[ctx.ClientId]
	s.m.RUnlock()
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

func (s *Store) Topic(name string) (*Topic, bool) {
	s.m.Lock()
	defer s.m.Unlock()

	return s.topic(name)
}

// topic creates topics lazily, the caller must hold s.m for writing.
func (s *Store) topic(name string) (*Topic, bool) {
	if t, ok := s.Topics[name]; ok {
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

		err := ch.Value.IsNameValid(name)
		if err != nil {
			continue
		}

		if s.Topics == nil {
			s.Topics = make(map[string]*Topic)
		}

		t := &Topic{Name: name, cfg: ch.Value}
		s.Topics[name] = t
		return t, true
	}

	return nil, false
}

func (s *Store) logMessage(apiName string, messageId string, topic *Topic, publish *mqtt.PublishRequest, actions []*engine.Action, args PublishArgs) {
	topicName := topic.cfg.ResolveAddress()
	labels := []string{apiName, topicName}
	if topic.Name != topicName {
		labels = append(labels, topic.Name)
	}

	s.monitor.Messages.WithLabel(labels...).Add(1)
	s.monitor.LastMessage.WithLabel(labels...).Set(float64(time.Now().Unix()))

	traits := events.
		NewTraits().
		WithNamespace("mqtt").
		WithName(apiName).
		With("topic", topicName).
		With("type", "message").
		With("clientId", args.ClientId)
	if topic.Name != topicName {
		traits.With("instance", topic.Name)
	}

	err := s.eh.Push(&LogMessage{
		Topic:     topic.Name,
		MessageId: messageId,
		Message: &LogValue{
			Value:  string(publish.Data),
			Binary: publish.Data,
		},
		Retain:   args.Retain,
		Api:      apiName,
		ClientId: args.ClientId,
		Actions:  actions,
	}, traits)
	if err != nil {
		log.Errorf("mqtt: failed to log message: %s", err)
	}
}
