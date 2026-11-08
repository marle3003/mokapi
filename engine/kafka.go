package engine

import (
	"errors"
	"fmt"
	"math/rand"
	"mokapi/engine/common"
	"mokapi/media"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/kafka/store"
	openapi "mokapi/providers/openapi/schema"
	"mokapi/runtime"
	avro "mokapi/schema/avro/schema"
	"mokapi/schema/json/generator"
	"mokapi/schema/json/schema"
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
	k, ch, err := c.tryGet(args.Cluster, args.Topic, args.Retry)
	if err != nil {
		return nil, err
	}
	if args.Topic == "" {
		args.Topic = ch.ResolveAddress()
	}

	client := store.NewClient(k.Store, c.app.Monitor.Kafka)
	client.AllowAutoTopicCreation = args.AllowAutoTopicCreation
	client.ClientId = args.ClientId
	client.ScriptFile = args.ScriptFile

	var produced []common.KafkaMessageResult
	for _, m := range args.Messages {
		msg, err := selectKafkaMessage(ch, m)
		if err != nil {
			return nil, fmt.Errorf("failed to produce message to Kafka topic '%v': %w", args.Topic, err)
		}

		value, err := kafkaValue(msg, m)
		if err != nil {
			return nil, fmt.Errorf("failed to produce message to Kafka topic '%v': %w", args.Topic, err)
		}

		if m.Key == nil {
			m.Key, err = createValue(kafkaKeySchema(msg))
			if err != nil {
				return nil, fmt.Errorf("unable to generate kafka key: %v", err)
			}
		}

		var headers []store.RecordHeader
		for hk, hv := range m.Headers {
			headers = append(headers, store.RecordHeader{Name: hk, Value: hv})
		}

		rec, err := client.Write(args.Topic, []store.Record{{
			Key:            m.Key,
			Value:          value,
			Headers:        headers,
			Partition:      m.Partition,
			SkipValidation: m.Value.IsSet,
		}}, media.ContentType{})
		if err != nil {
			return nil, fmt.Errorf("failed to produce message to Kafka topic '%v': %w", args.Topic, err)
		}

		if rec[0].Error != "" {
			return nil, fmt.Errorf("failed to produce message to Kafka topic '%v': %s", args.Topic, rec[0].Error)
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
		Api:      k.Info.Name,
		Topic:    args.Topic,
		Messages: produced,
	}, nil

}

// selectKafkaMessage returns the message selected by messageId. Without a messageId it returns
// the first message of the channel that is valid for the given value or data.
func selectKafkaMessage(ch *asyncapi3.Channel, m common.KafkaMessage) (*asyncapi3.Message, error) {
	return selectMessage(ch, m.MessageId, func() (*asyncapi3.Message, error) {
		var value any
		if m.Value.IsSet {
			value = m.Value.Value
		} else if m.Data.IsSet {
			value = m.Data.Value
		}

		msg, err := ch.GetMessageByPayload(value)
		if m.Value.IsSet {
			// a value is not validated
			return msg, nil
		}
		return msg, err
	})
}

// kafkaValue returns the value as given, the encoded data, or a payload generated from the message.
// Data is encoded here, so the store does not choose another message.
func kafkaValue(msg *asyncapi3.Message, m common.KafkaMessage) (any, error) {
	switch {
	case m.Value.IsSet:
		return m.Value.Value, nil
	case m.Data.IsSet:
		return encodePayload(msg, m.Data.Value)
	case m.MessageId == "":
		return generateKafkaValue(msg, generateAnyPayload)
	case payloadSchema(msg) == nil:
		// a selected message without payload is a record without value
		return []byte(nil), nil
	default:
		return generateKafkaValue(msg, generatePayload)
	}
}

func generateKafkaValue(msg *asyncapi3.Message, generate func(*asyncapi3.Message) (any, error)) (any, error) {
	value, err := generate(msg)
	if err != nil {
		return nil, fmt.Errorf("unable to generate kafka value: %v", err)
	}
	return value, nil
}

func kafkaKeySchema(msg *asyncapi3.Message) *asyncapi3.SchemaRef {
	if msg != nil && msg.Bindings != nil && msg.Bindings.Kafka.Key != nil {
		return msg.Bindings.Kafka.Key
	}
	return &asyncapi3.SchemaRef{
		Value: &schema.Schema{Type: schema.Types{"string"}, Pattern: "[a-z]{9}"},
	}
}

func (c *KafkaClient) tryGet(cluster string, topic string, retry common.RetryArgs) (k *runtime.KafkaInfo, t *asyncapi3.Channel, err error) {
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

func (c *KafkaClient) get(cluster string, topic string) (k *runtime.KafkaInfo, ch *asyncapi3.Channel, err error) {
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
			return clusters[0], topics[0].Config(), nil
		}

		var topics []*asyncapi3.Channel
		for _, v := range c.app.Kafka.List() {
			if t := v.Topic(topic); t != nil {
				k = v
				if len(cluster) == 0 {
					cluster = v.Info.Name
				}
				topics = append(topics, t.Config())
			} else if topic != "" {
				for _, ref := range v.Channels {
					if ref.Value == nil {
						continue
					}
					if err = ref.Value.IsNameValid(topic); err == nil {
						k = v
						topics = append(topics, ref.Value)
					}
				}
			}
		}
		if len(topics) > 1 {
			err = newAmbiguousError("ambiguous topic %v. Specify the cluster", topic)
			return
		} else if len(topics) == 1 {
			ch = topics[0]
		}
	} else {
		if k = c.app.Kafka.Get(cluster); k != nil {
			t := k.Topic(topic)
			if t != nil {
				ch = t.Config()
			}
		} else {
			return nil, nil, fmt.Errorf("kafka cluster '%v' not found", cluster)
		}
	}

	if ch == nil {
		err = fmt.Errorf("kafka topic '%v' not found", topic)
		return
	}

	return
}

func (c *KafkaClient) getPartition(t *store.Topic, partition int) (*store.Partition, error) {
	if partition < 0 {
		r := rand.New(rand.NewSource(time.Now().Unix()))
		partition = r.Intn(t.NumPartitions())
	}

	p := t.Partition(partition)
	if p == nil {
		return nil, fmt.Errorf("partiton %v does not exist", partition)
	}
	return p, nil
}

func createValue(r *asyncapi3.SchemaRef) (value any, err error) {
	var s asyncapi3.Schema
	if r != nil && r.Value != nil {
		s, err = r.GetSchema()
	} else {
		s = &schema.Schema{}
	}
	if err != nil {
		return
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
		err = fmt.Errorf("schema format not supported: %T", r.Value)
	}

	return
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
