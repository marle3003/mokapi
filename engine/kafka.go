package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"math/rand"
	"mokapi/engine/common"
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
	k, t, err := c.tryGet(args.Cluster, args.Topic, args.Retry)
	if err != nil {
		return nil, err
	}

	client := store.NewClient(k.Store, c.app.Monitor.Kafka)
	var produced []common.KafkaMessageResult
	for _, m := range args.Messages {
		value := m.Data
		ct := media.ContentType{}
		if m.Value != nil {
			value = m.Value
		}

		keySchema := &asyncapi3.SchemaRef{
			Value: &asyncapi3.MultiSchemaFormat{
				Schema: &schema.Schema{Type: schema.Types{"string"}, Pattern: "[a-z]{9}"},
			},
		}
		var payload *asyncapi3.SchemaRef
		var msg *asyncapi3.Message
		msg, err = selectMessage(value, t.Config, k.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to produce message to Kafka topic '%v': %w", t.Name, err)
		}
		if msg != nil {
			if msg.Bindings.Kafka.Key != nil {
				keySchema = msg.Bindings.Kafka.Key
			}
			payload = msg.Payload
			ct = media.ParseContentType(msg.ContentType)
		} else {
			ct = media.ParseContentType(k.DefaultContentType)
		}

		if m.Key == nil {
			m.Key, err = createValue(keySchema)
			if err != nil {
				return nil, fmt.Errorf("unable to generate kafka key: %v", err)
			}
		}

		if value == nil {
			value, err = createValue(payload)
			if err != nil {
				return nil, fmt.Errorf("unable to generate kafka value: %v", err)
			}
		}

		if m.Value == nil && payload != nil {
			value, err = payload.Value.Marshal(value, ct)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal kafka message: %w", err)
			}
		}

		var headers []store.RecordHeader
		for hk, hv := range m.Headers {
			headers = append(headers, store.RecordHeader{Name: hk, Value: hv})
		}

		rec, err := client.Write(t.Name, []store.Record{{
			Key:            m.Key,
			Value:          value,
			Headers:        headers,
			Partition:      m.Partition,
			SkipValidation: m.Value != nil,
		}}, ct)
		if err != nil {
			return nil, fmt.Errorf("failed to produce message to Kafka topic '%v': %w", t.Name, err)
		}

		if rec[0].Error != "" {
			return nil, fmt.Errorf("failed to produce message to Kafka topic '%v': %s", t.Name, rec[0].Error)
		}

		produced = append(produced, common.KafkaMessageResult{
			Key:       string(rec[0].Key),
			Value:     string(rec[0].Value),
			Offset:    rec[0].Offset,
			Headers:   m.Headers,
			Partition: rec[0].Partition,
		})
	}

	return &common.KafkaProduceResult{
		Cluster:  k.Info.Name,
		Topic:    t.Name,
		Messages: produced,
	}, nil

}

func (c *KafkaClient) tryGet(cluster string, topic string, retry common.KafkaProduceRetry) (k *runtime.KafkaInfo, t *store.Topic, err error) {
	count := 0
	backoff := retry.InitialRetryTime
	for {
		k, t, err = c.get(cluster, topic)
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

func (c *KafkaClient) get(cluster string, topic string) (k *runtime.KafkaInfo, t *store.Topic, err error) {
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
			return clusters[0], topics[0], nil
		}

		var topics []*store.Topic
		for _, v := range c.app.Kafka.List() {
			if t := v.Topic(topic); t != nil {
				k = v
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
		if k = c.app.Kafka.Get(cluster); k != nil {
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
		jsSchema := avro.ConvertToJsonSchema(v)
		value, err = generator.New(&generator.Request{Schema: jsSchema})
	default:
		err = fmt.Errorf("schema format not supported: %v", r.Value.Format)
	}

	return
}

func selectMessage(value any, topic *asyncapi3.Channel, cfg *asyncapi3.Config) (*asyncapi3.Message, error) {
	noOperationDefined := true
	var validationErr error

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
				if validationErr = valueMatchMessagePayload(value, msg.Value); validationErr == nil {
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
				if validationErr = valueMatchMessagePayload(value, msg.Value); validationErr == nil {
					return msg.Value, nil
				}
			}
		}
	}

	if noOperationDefined {
		return nil, fmt.Errorf("no 'send' or 'receive' operation defined in specification")
	}

	if value != nil {
		switch value.(type) {
		case string, []byte:
			break
		default:
			b, err := json.Marshal(value)
			if err == nil {
				value = string(b)
			}
		}
		return nil, fmt.Errorf("no matching message configuration found for the given value: %v\nhint:\n%w\n", value, validationErr)
	}
	return nil, fmt.Errorf("no message ")
}

func valueMatchMessagePayload(value any, msg *asyncapi3.Message) error {
	if value == nil || msg.Payload == nil {
		return nil
	}
	ct := media.ParseContentType(msg.ContentType)

	switch v := msg.Payload.Value.Schema.(type) {
	case *schema.Schema:
		_, err := encoding.NewEncoder(v).Write(value, ct)
		return err
	case *openapi.Schema:
		_, err := v.Marshal(value, ct)
		return err
	case *avro.Schema:
		jsSchema := avro.ConvertToJsonSchema(v)
		_, err := encoding.NewEncoder(jsSchema).Write(value, ct)
		return err
	default:
		return nil
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
