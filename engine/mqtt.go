package engine

import (
	"errors"
	"fmt"
	"maps"
	"mokapi/engine/common"
	"mokapi/mqtt"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/mqtt/store"
	"mokapi/runtime"
	"slices"
	"time"

	log "github.com/sirupsen/logrus"
)

type MqttClient struct {
	app *runtime.App
}

func NewMqttClient(app *runtime.App) *MqttClient {
	return &MqttClient{
		app: app,
	}
}

func (c *MqttClient) Publish(args *common.MqttPublishArgs) (*common.MqttPublishResult, error) {
	m, t, err := c.tryGet(args.Cluster, args.Topic, args.Retry)
	if err != nil {
		return nil, err
	}

	pa := store.PublishArgs{
		// an empty payload (value null or '') is the MQTT tombstone and is written as is
		SkipValidation: args.Value != nil && len(args.Value) == 0,
		Retain:         args.Retain,
		ClientId:       args.ClientId,
		ScriptFile:     args.ScriptFile,
	}

	payload, err := mqttPayload(t, args)
	if err != nil {
		return nil, err
	}

	req := &mqtt.PublishRequest{
		Topic: t.Name,
		Data:  payload,
	}

	_, err = m.Publish(req, pa)

	if err != nil {
		return nil, err
	}
	return &common.MqttPublishResult{
		Cluster: m.Info.Name,
		Topic:   t.Name,
		Value:   string(payload),
	}, nil
}

// mqttPayload returns the value as given, the encoded data, or a payload generated from a message
// of the topic. A messageId selects that message, otherwise the message is chosen by the data
// or, for a generated payload, randomly.
func mqttPayload(t *store.Topic, args *common.MqttPublishArgs) ([]byte, error) {
	ch := t.Config()
	if ch == nil {
		// a topic without channel has no messages, e.g. $SYS topics
		if args.MessageId != "" {
			return nil, fmt.Errorf("message '%v' not found in channel '%v'", args.MessageId, t.Name)
		}
		return args.Value, nil
	}

	msg, err := selectMessage(ch, args.MessageId, func() (*asyncapi3.Message, error) {
		if args.Data.IsSet {
			return ch.GetMessageByPayload(args.Data.Value)
		}
		return anyMessage(ch), nil
	})
	if err != nil {
		return nil, err
	}

	switch {
	case args.Value != nil:
		return args.Value, nil
	case args.Data.IsSet:
		return encodePayload(msg, args.Data.Value)
	}

	if payloadSchema(msg) == nil {
		return nil, nil
	}
	value, err := generatePayload(msg)
	if err != nil {
		return nil, fmt.Errorf("unable to generate mqtt payload: %v", err)
	}
	return encodePayload(msg, value)
}

func (c *MqttClient) tryGet(cluster string, topic string, retry common.RetryArgs) (m *runtime.MqttInfo, t *store.Topic, err error) {
	count := 0
	backoff := retry.InitialRetryTime
	for {
		m, t, err = c.get(cluster, topic)
		ambiguous := &ambiguousError{}
		if err == nil || errors.As(err, &ambiguous) {
			return
		}
		count++
		if count >= retry.Retries || backoff > retry.MaxRetryTime {
			return
		}
		log.Debugf("kafka topic '%v' not found. Retry in %v", topic, backoff)
		time.Sleep(backoff)
		backoff *= time.Duration(retry.Factor)
	}
}

func (c *MqttClient) get(cluster string, topic string) (m *runtime.MqttInfo, t *store.Topic, err error) {
	if len(cluster) == 0 {
		if len(topic) == 0 {
			clusters := c.app.Mqtt.List()
			if len(clusters) > 1 {
				err = newAmbiguousError("ambiguous cluster: specify the cluster")
				return
			}
			topics := clusters[0].Topics
			if len(topics) > 1 {
				err = newAmbiguousError("ambiguous topic %v. Specify the cluster", topic)
				return
			}
			if len(topics) == 0 {
				return
			}
			t = slices.Collect(maps.Values(topics))[0]
			return clusters[0], t, nil
		}

		var topics []*store.Topic
		var ok bool
		for _, v := range c.app.Mqtt.List() {
			if t, ok = v.Topic(topic); ok {
				m = v
				if len(cluster) == 0 {
					cluster = v.Info.Name
				}
				topics = append(topics, t)
			}
		}
		if len(topics) > 1 {
			err = newAmbiguousError("ambiguous topic %v. Specify the cluster", topic)
			return
		} else if len(topics) == 1 {
			t = topics[0]
		}
	} else {
		if m = c.app.Mqtt.Get(cluster); m != nil {
			t, _ = m.Topic(topic)
		} else {
			return nil, nil, fmt.Errorf("kafka cluster '%v' not found", cluster)
		}
	}

	if t == nil {
		err = fmt.Errorf("kafka topic '%v' not found", topic)
		return
	}

	return
}
