package engine

import (
	"fmt"
	"math/rand"
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/asyncApi/kafka/store"
	"mokapi/engine/common"
	"mokapi/kafka"
	"mokapi/media"
	"mokapi/providers/openapi/schema"
	"mokapi/runtime"
	"time"
)

type kafkaClient struct {
	app *runtime.App
}

func newKafkaClient(app *runtime.App) *kafkaClient {
	return &kafkaClient{
		app: app,
	}
}

func (c *kafkaClient) Produce(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
	t, p, config, err := c.get(args.Cluster, args.Topic, args.Partition)
	if err != nil {
		return nil, err
	}

	ch := config.Channels[t.Name]
	if ch == nil || ch.Value == nil {
		return nil, fmt.Errorf("produce kafka message to '%v' failed: invalid topic configuration", t.Name)
	}

	rb, err := c.createRecordBatch(args.Key, args.Value, args.Headers, ch.Value)
	if err != nil {
		return nil, fmt.Errorf("produce kafka message to '%v' failed: %w", t.Name, err)
	}

	_, err = p.Write(rb)
	if err != nil {
		return nil, fmt.Errorf("produce kafka message to '%v' failed: %w", t.Name, err)
	}

	k := kafka.BytesToString(rb.Records[0].Key)
	v := kafka.BytesToString(rb.Records[0].Value)
	t.Store().UpdateMetrics(c.app.Monitor.Kafka, t, p, rb)

	h := map[string]string{}
	for _, v := range rb.Records[0].Headers {
		h[v.Key] = string(v.Value)
	}

	return &common.KafkaProduceResult{
		Cluster:   config.Info.Name,
		Topic:     t.Name,
		Partition: p.Index,
		Offset:    rb.Records[0].Offset,
		Key:       k,
		Value:     v,
		Headers:   h,
	}, nil

}

func (c *kafkaClient) write(partition *store.Partition, config *asyncApi.Channel, key, value interface{}, headers map[string]interface{}) (interface{}, interface{}, error) {
	msg := config.Publish.Message.Value
	if msg == nil {
		return nil, nil, fmt.Errorf("message configuration missing")
	}
	var err error

	if key == nil {
		key, err = schema.CreateValue(msg.Bindings.Kafka.Key)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to generate kafka key: %v", err)
		}
	}
	if value == nil {
		value, err = schema.CreateValue(msg.Payload)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to generate kafka data: %v", err)
		}
	}

	var k []byte
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

func (c *kafkaClient) get(cluster string, topic string, partition int) (t *store.Topic, p *store.Partition, config *asyncApi.Config, err error) {
	if len(cluster) == 0 {
		var topics []*store.Topic
		for _, v := range c.app.Kafka {
			if t := v.Topic(topic); t != nil {
				config = v.Config
				if len(cluster) == 0 {
					cluster = v.Info.Name
				}
				topics = append(topics, t)
			}
		}
		if len(topics) > 1 {
			err = fmt.Errorf("ambiguous topic %v. Specify the cluster", topic)
			return
		} else if len(topics) == 1 {
			t = topics[0]
		}
	} else {
		if k, ok := c.app.Kafka[cluster]; ok {
			config = k.Config
			t = k.Topic(topic)
		}
	}

	if t == nil {
		err = fmt.Errorf("kafka topic '%v' not found", topic)
		return
	}

	if partition < 0 {
		r := rand.New(rand.NewSource(time.Now().Unix()))
		partition = r.Intn(len(t.Partitions))
	} else if partition >= len(t.Partitions) {
		err = fmt.Errorf("partiton %v does not exist", partition)
		return
	}

	p = t.Partition(partition)

	return
}

func (c *kafkaClient) createRecordBatch(key, value interface{}, headers map[string]interface{}, config *asyncApi.Channel) (rb kafka.RecordBatch, err error) {
	msg := config.Publish.Message.Value
	if msg == nil {
		err = fmt.Errorf("message configuration missing")
		return
	}

	if key == nil {
		key, err = schema.CreateValue(msg.Bindings.Kafka.Key)
		if err != nil {
			return rb, fmt.Errorf("unable to generate kafka key: %v", err)
		}
	}
	if value == nil {
		value, err = schema.CreateValue(msg.Payload)
		if err != nil {
			return rb, fmt.Errorf("unable to generate kafka data: %v", err)
		}
	}

	var v []byte
	v, err = msg.Payload.Marshal(value, media.ParseContentType("application/json"))
	if err != nil {
		return
	}

	var k []byte
	if msg.Bindings.Kafka.Key != nil {
		switch msg.Bindings.Kafka.Key.Value.Type {
		case "object", "array":
			k, err = msg.Bindings.Kafka.Key.Marshal(key, media.ParseContentType("application/json"))
			if err != nil {
				return
			}
		default:
			k = []byte(fmt.Sprintf("%v", key))
		}
	}

	var recordHeaders []kafka.RecordHeader
	recordHeaders, err = getHeaders(headers, msg.Headers)
	if err != nil {
		return
	}

	rb = kafka.RecordBatch{Records: []kafka.Record{
		{
			Key:     kafka.NewBytes(k),
			Value:   kafka.NewBytes(v),
			Headers: recordHeaders,
		},
	}}
	return
}

// todo: only specified headers should be written
func getHeaders(headers map[string]interface{}, r *schema.Ref) ([]kafka.RecordHeader, error) {
	var result []kafka.RecordHeader
	for k, v := range headers {
		var headerSchema *schema.Ref
		if r != nil && r.Value != nil && r.Value.Type == "object" {
			headerSchema = r.Value.Properties.Get(k)

		}

		b, err := headerSchema.Marshal(v, media.Any)
		if err != nil {
			return nil, err
		}
		result = append(result, kafka.RecordHeader{
			Key:   k,
			Value: b,
		})
	}
	return result, nil
}
