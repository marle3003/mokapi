package store

import (
	"mokapi/mqtt"
	"mokapi/runtime/events"

	log "github.com/sirupsen/logrus"
)

func (s *Store) publish(rw mqtt.MessageWriter, publish *mqtt.PublishRequest, qos byte, retain bool, ctx *mqtt.ClientContext) {
	msg := &Message{
		Topic:  publish.Topic,
		Data:   publish.Data,
		QoS:    qos,
		Retain: retain,
	}

	topic, ok := s.Topics[msg.Topic]
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

	err := topic.validate(msg.Data)
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

	client, _ := s.clients[ctx.ClientId]
	err = s.eh.Push(&LogMessage{
		Topic: publish.Topic,
		Value: LogValue{
			Value:  string(publish.Data),
			Binary: publish.Data,
		},
		Api:      s.cfg.Info.Name,
		ClientId: client.Id,
	}, events.NewTraits().WithNamespace("mqtt").WithName(s.cfg.Info.Name).With("topic", msg.Topic))
	if err != nil {
		log.Errorf("mqtt: failed to log message: %s", err)
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
