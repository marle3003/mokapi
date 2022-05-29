package engine

import (
	"fmt"
	"math/rand"
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/asyncApi/kafka/store"
	"mokapi/config/dynamic/openapi/schema"
	"mokapi/kafka"
	"mokapi/media"
	"mokapi/runtime"
	"time"
)

type kafkaClient struct {
	app       *runtime.App
	generator *schema.Generator
}

func newKafkaClient(app *runtime.App) *kafkaClient {
	return &kafkaClient{
		app:       app,
		generator: schema.NewGenerator(),
	}
}

func (c *kafkaClient) Produce(cluster string, topic string, partition int, key, value interface{}, headers map[string]interface{}) (interface{}, interface{}, error) {
	var t *store.Topic
	var config *asyncApi.Config
	if len(cluster) == 0 {
		var topics []*store.Topic
		for _, v := range c.app.Kafka {
			config = v.Config
			if t := v.Topic(topic); t != nil {
				if len(cluster) == 0 {
					cluster = v.Info.Name
				}
				topics = append(topics, t)
			}
		}
		if len(topics) > 1 {
			return nil, nil, fmt.Errorf("ambiguous topic %v. Specify the cluster", topic)
		} else if len(topics) == 1 {
			t = topics[0]
		}
	} else {
		if c, ok := c.app.Kafka[cluster]; ok {
			config = c.Config
			t = c.Topic(topic)
		}
	}

	if t == nil {
		return nil, nil, fmt.Errorf("kafka topic '%v' not found", topic)
	}

	if partition < 0 {
		rand.Seed(time.Now().Unix())
		partition = rand.Intn(len(t.Partitions))
	} else if partition >= len(t.Partitions) {
		return nil, nil, fmt.Errorf("partiton %v does not exist", partition)
	}

	ch := config.Channels[t.Name]
	if ch.Value == nil {
		return nil, nil, fmt.Errorf("invalid topic configuration")
	}

	k, v, err := c.write(t.Partition(partition), ch.Value, key, value, headers)
	if err != nil {
		return nil, nil, err
	}
	c.app.Monitor.Kafka.Messages.WithLabel(cluster, t.Name).Add(1)
	c.app.Monitor.Kafka.LastMessage.WithLabel(cluster, t.Name).Set(float64(time.Now().Unix()))
	return k, v, nil

}

func (c *kafkaClient) write(partition *store.Partition, config *asyncApi.Channel, key, value interface{}, headers map[string]interface{}) (interface{}, interface{}, error) {
	msg := config.Publish.Message.Value
	if msg == nil {
		return nil, nil, fmt.Errorf("message configuration missing")
	}

	if key == nil {
		key = c.generator.New(msg.Bindings.Kafka.Key)
	}
	if value == nil {
		value = c.generator.New(msg.Payload)
	}

	var k []byte
	var err error
	switch msg.Bindings.Kafka.Key.Value.Type {
	case "object", "array":
		k, err = msg.Bindings.Kafka.Key.Marshal(key, media.ParseContentType("application/json"))
		if err != nil {
			return nil, nil, err
		}
	default:
		k = []byte(fmt.Sprintf("%v", key))
	}

	v, err := msg.Payload.Marshal(value, media.ParseContentType("application/json"))
	if err != nil {
		return nil, nil, err
	}

	_, err = partition.Write(kafka.RecordBatch{Records: []kafka.Record{
		{
			Key:     kafka.NewBytes(k),
			Value:   kafka.NewBytes(v),
			Headers: nil,
		},
	}})

	if err != nil {
		return nil, nil, err
	}

	return key, value, nil
}
