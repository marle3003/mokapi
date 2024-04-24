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
	"strings"
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
	t, config, err := c.get(args.Cluster, args.Topic)
	if err != nil {
		return nil, err
	}

	ch := config.Channels[t.Name]
	if ch == nil || ch.Value == nil {
		return nil, fmt.Errorf("produce kafka message to '%v' failed: invalid topic configuration", t.Name)
	}

	var produced []common.KafkaProducedMessage
	for _, r := range args.Messages {
		var options []store.WriteOptions
		if r.Value != nil {
			options = append(options, func(args *store.WriteArgs) {
				args.SkipValidation = true
			})
		}

		p, err := c.getPartition(t, r.Partition)
		if err != nil {
			return nil, err
		}

		v := r.Data
		if r.Value != nil {
			v = r.Value
		}

		rb, err := c.createRecordBatch(r.Key, v, r.Headers, ch.Value)
		if err != nil {
			return nil, fmt.Errorf("produce kafka message to '%v' failed: %w", t.Name, err)
		}
		_, records, err := p.Write(rb)
		if err != nil {
			var sb strings.Builder
			for _, r := range records {
				sb.WriteString(fmt.Sprintf("%v: %v\n", r.BatchIndex, r.BatchIndexErrorMessage))
			}
			return nil, fmt.Errorf("produce kafka message to '%v' failed: %w \n%v", t.Name, err, sb.String())
		}
		t.Store().UpdateMetrics(c.app.Monitor.Kafka, t, p, rb)

		h := map[string]string{}
		for _, v := range rb.Records[0].Headers {
			h[v.Key] = string(v.Value)
		}

		produced = append(produced, common.KafkaProducedMessage{
			Key:       kafka.BytesToString(rb.Records[0].Key),
			Value:     kafka.BytesToString(rb.Records[0].Value),
			Offset:    rb.Records[0].Offset,
			Headers:   h,
			Partition: p.Index,
		})
	}

	return &common.KafkaProduceResult{
		Cluster:  config.Info.Name,
		Topic:    t.Name,
		Messages: produced,
	}, nil

}

func (c *kafkaClient) get(cluster string, topic string) (t *store.Topic, config *asyncApi.Config, err error) {
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

	return
}

func (c *kafkaClient) getPartition(t *store.Topic, partition int) (*store.Partition, error) {
	if partition < 0 {
		r := rand.New(rand.NewSource(time.Now().Unix()))
		partition = r.Intn(len(t.Partitions))
	} else if partition >= len(t.Partitions) {
		return nil, fmt.Errorf("partiton %v does not exist", partition)
	}

	return t.Partition(partition), nil
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
	if b, ok := value.([]byte); ok {
		v = b
	} else {
		v, err = msg.Payload.Marshal(value, media.ParseContentType("application/json"))
		if err != nil {
			return
		}
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
