package engine

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"mokapi/engine/common"
	"mokapi/kafka"
	"mokapi/media"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/kafka/store"
	openapi "mokapi/providers/openapi/schema"
	"mokapi/runtime"
	avro "mokapi/schema/avro/schema"
	"mokapi/schema/encoding"
	"mokapi/schema/json/generator"
	"mokapi/schema/json/schema"
	"strings"
	"time"
)

type KafkaClient struct {
	app *runtime.App
}

func NewKafkaClient(app *runtime.App) *KafkaClient {
	return &KafkaClient{
		app: app,
	}
}

func (c *KafkaClient) Produce(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
	t, config, err := c.tryGet(args.Cluster, args.Topic, args.Retry)
	if err != nil {
		return nil, err
	}

	ch := config.Channels[t.Name]
	if ch == nil || ch.Value == nil {
		return nil, fmt.Errorf("produce kafka message to '%v' failed: invalid topic configuration", t.Name)
	}

	var produced []common.KafkaMessageResult
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

		_, records, err := p.Write(rb, options...)
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

		produced = append(produced, common.KafkaMessageResult{
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

func (c *KafkaClient) tryGet(cluster string, topic string, retry common.KafkaProduceRetry) (t *store.Topic, config *asyncapi3.Config, err error) {
	count := 0
	backoff := retry.InitialRetryTime
	for {
		t, config, err = c.get(cluster, topic)
		if err == nil {
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

func (c *KafkaClient) get(cluster string, topic string) (t *store.Topic, config *asyncapi3.Config, err error) {
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

func (c *KafkaClient) getPartition(t *store.Topic, partition int) (*store.Partition, error) {
	if partition < 0 {
		r := rand.New(rand.NewSource(time.Now().Unix()))
		partition = r.Intn(len(t.Partitions))
	} else if partition >= len(t.Partitions) {
		return nil, fmt.Errorf("partiton %v does not exist", partition)
	}

	return t.Partition(partition), nil
}

func (c *KafkaClient) createRecordBatch(key, value interface{}, headers map[string]interface{}, config *asyncapi3.Channel) (rb kafka.RecordBatch, err error) {
	n := len(config.Messages)
	if n == 0 {
		err = fmt.Errorf("message configuration missing")
		return
	}
	var msg *asyncapi3.Message
	// select first message
	for _, m := range config.Messages {
		if m.Value == nil {
			continue
		}
		msg = m.Value
		break
	}
	if msg == nil {
		err = fmt.Errorf("message configuration missing")
		return
	}

	if key == nil {
		key, err = createValue(msg.Bindings.Kafka.Key.Value.Schema.(*schema.Ref))
		if err != nil {
			return rb, fmt.Errorf("unable to generate kafka key: %v", err)
		}
	}

	s := msg.Payload.Value.Schema
	var jsSchema *schema.Ref
	switch v := s.(type) {
	case *schema.Ref:
		if value == nil {
			value, err = createValue(v)
			if err != nil {
				return rb, fmt.Errorf("unable to generate kafka data: %v", err)
			}
		}
		jsSchema = v
	case *openapi.Ref:
		if value == nil {
			value, err = openapi.CreateValue(v)
			if err != nil {
				return rb, fmt.Errorf("unable to generate kafka data: %v", err)
			}
		}
		jsSchema = openapi.ConvertToJsonSchema(v)
	case *avro.Schema:
		jsSchema = &schema.Ref{Value: v.Convert()}
		if value == nil {
			value, err = createValue(jsSchema)
			if err != nil {
				return rb, fmt.Errorf("unable to generate kafka data: %v", err)
			}
		}
	default:
		err = fmt.Errorf("schema format not supported: %v", msg.Payload.Value.Format)
		return
	}
	s, ok := msg.Payload.Value.Schema.(*schema.Ref)
	if !ok {
		if _, ok = value.([]byte); !ok {
			err = fmt.Errorf("currently only json schema supported")
			return
		}
	}

	var v []byte
	if b, ok := value.([]byte); ok {
		v = b
	} else {
		if sOpenApi, ok := msg.Payload.Value.Schema.(*openapi.Ref); ok {
			v, err = sOpenApi.Marshal(value, media.ParseContentType(msg.ContentType))
		} else {
			v, err = encoding.NewEncoder(jsSchema).Write(value, media.ParseContentType(msg.ContentType))
		}
		if err != nil {
			return
		}
	}

	var k []byte
	if msg.Bindings.Kafka.Key != nil {
		s := msg.Bindings.Kafka.Key.Value.Schema.(*schema.Ref)
		switch {
		case s.IsOneOf("object", "array"):
			k, err = encoding.NewEncoder(s).Write(key, media.ParseContentType("application/json"))
			if err != nil {
				return
			}
		default:
			k = []byte(fmt.Sprintf("%v", key))
		}
	}

	var recordHeaders []kafka.RecordHeader
	var hs *schema.Ref
	if msg.Headers != nil {
		hs, ok = msg.Headers.Value.Schema.(*schema.Ref)
		if !ok {
			err = fmt.Errorf("currently only json schema supported")
			return
		}
	}
	recordHeaders, err = getHeaders(headers, hs)
	if err != nil {
		return
	}

	rb = kafka.RecordBatch{Records: []*kafka.Record{
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
		if r != nil && r.Value != nil && r.Value.Type.IsObject() {
			headerSchema = r.Value.Properties.Get(k)

		}

		b, err := encoding.NewEncoder(headerSchema).Write(v, media.Any)
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

func createValue(s *schema.Ref) (interface{}, error) {
	return generator.New(&generator.Request{Path: generator.Path{&generator.PathElement{Schema: s}}})
}
