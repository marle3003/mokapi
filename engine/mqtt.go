package engine

import (
	"errors"
	"fmt"
	"maps"
	"mokapi/engine/common"
	"mokapi/mqtt"
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
		Retain:     args.Retain,
		ClientId:   args.ClientId,
		ScriptFile: args.ScriptFile,
	}

	_, err = m.Publish(&mqtt.PublishRequest{
		Topic: t.Name,
		Data:  []byte(args.Value),
	}, pa)

	if err != nil {
		return nil, err
	}
	return &common.MqttPublishResult{
		Cluster: m.Info.Name,
		Topic:   t.Name,
		Value:   args.Value,
	}, nil
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
