package engine

import (
	"errors"
	"fmt"
	"maps"
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
	"slices"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
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

	var produced []common.KafkaMessageResult
	for _, r := range args.Messages {
		p, err := c.getPartition(t, r.Partition)
		if err != nil {
			return nil, err
		}

		v := r.Data
		if r.Value != nil {
			v = r.Value
		}

		rb, err := c.createRecordBatch(r.Key, v, r.Headers, t.Config, config)
		if err != nil {
			return nil, fmt.Errorf("producing kafka message to '%v' failed: %w", t.Name, err)
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

func (c *KafkaClient) get(cluster string, topic string) (t *store.Topic, config *asyncapi3.Config, err error) {
	if len(cluster) == 0 {
		if len(topic) == 0 {
			clusters := c.app.Kafka.List()
			if len(clusters) > 1 {
				err = newAmbiguousError("ambiguous cluster: specify the cluster")
				return
			}
			topics := clusters[0].Topics()
			if len(topics) > 1 {
				err = newAmbiguousError("ambiguous topic %v. Specify the cluster", topic)
				return
			}
			return topics[0], clusters[0].Config, nil
		}

		var topics []*store.Topic
		for _, v := range c.app.Kafka.List() {
			if t := v.Topic(topic); t != nil {
				config = v.Config
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
		if k := c.app.Kafka.Get(cluster); k != nil {
			config = k.Config
			t = k.Topic(topic)
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

func (c *KafkaClient) getPartition(t *store.Topic, partition int) (*store.Partition, error) {
	if partition < 0 {
		r := rand.New(rand.NewSource(time.Now().Unix()))
		partition = r.Intn(len(t.Partitions))
	} else if partition >= len(t.Partitions) {
		return nil, fmt.Errorf("partiton %v does not exist", partition)
	}

	return t.Partition(partition), nil
}

func (c *KafkaClient) createRecordBatch(key, value interface{}, headers map[string]interface{}, topic *asyncapi3.Channel, config *asyncapi3.Config) (rb kafka.RecordBatch, err error) {
	contentType := config.DefaultContentType
	var payload *asyncapi3.SchemaRef
	keySchema := &asyncapi3.SchemaRef{
		Value: &asyncapi3.MultiSchemaFormat{
			Schema: &schema.Schema{Type: schema.Types{"string"}, Pattern: "[a-z]{9}"},
		},
	}
	var headerSchema *schema.Schema

	if len(topic.Messages) > 0 {
		var msg *asyncapi3.Message
		msg, err = selectMessage(value, topic, config)
		if err != nil {
			return
		}
		payload = msg.Payload

		if msg.Bindings.Kafka.Key != nil {
			keySchema = msg.Bindings.Kafka.Key
		}

		if msg.ContentType != "" {
			contentType = msg.ContentType
		}

		if msg.Headers != nil {
			var ok bool
			headerSchema, ok = msg.Headers.Value.Schema.(*schema.Schema)
			if !ok {
				err = fmt.Errorf("currently only json schema supported")
				return
			}
		}
	}

	if key == nil {
		key, err = createValue(keySchema)
		if err != nil {
			return rb, fmt.Errorf("unable to generate kafka key: %v", err)
		}
	}

	if value == nil {
		value, err = createValue(payload)
		if err != nil {
			return rb, fmt.Errorf("unable to generate kafka value: %v", err)
		}
	}

	if contentType == "" {
		// set default: https://github.com/asyncapi/spec/issues/319
		contentType = "application/json"
	}

	var v []byte
	if b, ok := value.([]byte); ok {
		v = b
	} else {
		v, err = marshal(value, payload, contentType)
		if err != nil {
			return
		}
	}

	var k []byte
	if b, ok := key.([]byte); ok {
		k = b
	} else {
		k, err = marshalKey(key, keySchema)
		if err != nil {
			return
		}
	}

	var recordHeaders []kafka.RecordHeader
	recordHeaders, err = getHeaders(headers, headerSchema)
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
func getHeaders(headers map[string]interface{}, r *schema.Schema) ([]kafka.RecordHeader, error) {
	var result []kafka.RecordHeader
	for k, v := range headers {
		var headerSchema *schema.Schema
		if r != nil && r.Type.IsObject() {
			headerSchema = r.Properties.Get(k)

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

func createValue(r *asyncapi3.SchemaRef) (value interface{}, err error) {
	var s asyncapi3.Schema
	if r != nil && r.Value != nil && r.Value.Schema != nil {
		s = r.Value.Schema
	} else {
		s = &schema.Schema{}
	}

	switch v := s.(type) {
	case *schema.Schema:
		value, err = generator.New(&generator.Request{Schema: v})
	case *openapi.Schema:
		value, err = openapi.CreateValue(v)
	case *avro.Schema:
		jsSchema := v.Convert()
		value, err = generator.New(&generator.Request{Schema: jsSchema})
	default:
		err = fmt.Errorf("schema format not supported: %v", r.Value.Format)
	}

	return
}

func marshal(value interface{}, r *asyncapi3.SchemaRef, contentType string) ([]byte, error) {
	var s asyncapi3.Schema
	if r != nil && r.Value != nil && r.Value.Schema != nil {
		s = r.Value.Schema
	} else {
		s = &schema.Schema{}
	}

	switch v := s.(type) {
	case *schema.Schema:
		return encoding.NewEncoder(v).Write(value, media.ParseContentType(contentType))
	case *openapi.Schema:
		return v.Marshal(value, media.ParseContentType(contentType))
	case *avro.Schema:
		jsSchema := v.Convert()
		return encoding.NewEncoder(jsSchema).Write(value, media.ParseContentType(contentType))
	default:
		return nil, fmt.Errorf("schema format not supported: %v", r.Value.Format)
	}
}

func marshalKey(key interface{}, r *asyncapi3.SchemaRef) ([]byte, error) {
	var s asyncapi3.Schema
	if r != nil && r.Value != nil && r.Value.Schema != nil {
		s = r.Value.Schema
	} else {
		s = &schema.Schema{}
	}

	switch v := s.(type) {
	case *schema.Schema:
		if v.IsObject() || v.IsArray() {
			return encoding.NewEncoder(v).Write(key, media.ParseContentType("application/json"))
		} else {
			return []byte(fmt.Sprintf("%v", key)), nil
		}
	case *openapi.Schema:
		if v.Type.IsObject() || v.Type.IsArray() {
			return v.Marshal(key, media.ParseContentType("application/json"))
		} else {
			return []byte(fmt.Sprintf("%v", key)), nil
		}
	case *avro.Schema:
		jsSchema := v.Convert()
		if jsSchema.IsObject() || jsSchema.IsArray() {
			return encoding.NewEncoder(jsSchema).Write(key, media.ParseContentType("application/json"))
		} else {
			return []byte(fmt.Sprintf("%v", key)), nil
		}
	default:
		return nil, fmt.Errorf("schema format not supported: %v", r.Value.Format)
	}
}

func selectMessage(value any, topic *asyncapi3.Channel, cfg *asyncapi3.Config) (*asyncapi3.Message, error) {
	noOperationDefined := true

	// first try to get send operation
	for _, op := range cfg.Operations {
		if op.Value == nil || op.Value.Channel.Value == nil {
			continue
		}
		if op.Value.Channel.Value == topic && op.Value.Action == "send" {
			noOperationDefined = false
			var messages []*asyncapi3.MessageRef
			if len(op.Value.Messages) == 0 {
				messages = slices.Collect(maps.Values(op.Value.Channel.Value.Messages))
			} else {
				messages = op.Value.Messages
			}
			for _, msg := range messages {
				if msg.Value == nil {
					continue
				}
				if valueMatchMessagePayload(value, msg.Value) {
					return msg.Value, nil
				}
			}
		}
	}

	// second, try to get receive operation
	for _, op := range cfg.Operations {
		if op.Value == nil || op.Value.Channel.Value == nil {
			continue
		}
		if op.Value.Channel.Value == topic && op.Value.Action == "receive" {
			noOperationDefined = false
			var messages []*asyncapi3.MessageRef
			if len(op.Value.Messages) == 0 {
				messages = slices.Collect(maps.Values(op.Value.Channel.Value.Messages))
			} else {
				messages = op.Value.Messages
			}
			for _, msg := range messages {
				if msg.Value == nil {
					continue
				}
				if valueMatchMessagePayload(value, msg.Value) {
					return msg.Value, nil
				}
			}
		}
	}

	if noOperationDefined {
		return nil, fmt.Errorf("no 'send' or 'receive' operation defined in specification")
	}

	if value != nil {
		return nil, fmt.Errorf("no message configuration matches the message value for topic '%s' and value: %v", topic.GetName(), value)
	}
	return nil, fmt.Errorf("no message ")
}

func valueMatchMessagePayload(value any, msg *asyncapi3.Message) bool {
	if value == nil || msg.Payload == nil {
		return true
	}

	switch v := msg.Payload.Value.Schema.(type) {
	case *schema.Schema:
		_, err := encoding.NewEncoder(v).Write(value, media.ParseContentType("application/json"))
		return err == nil
	case *openapi.Schema:
		_, err := v.Marshal(value, media.ParseContentType("application/json"))
		return err == nil
	case *avro.Schema:
		jsSchema := v.Convert()
		_, err := encoding.NewEncoder(jsSchema).Write(value, media.ParseContentType("application/json"))
		return err == nil
	default:
		return false
	}
}

type ambiguousError struct {
	msg string
}

func (e *ambiguousError) Error() string {
	return e.msg
}

func newAmbiguousError(format string, args ...any) *ambiguousError {
	return &ambiguousError{fmt.Sprintf(format, args...)}
}
