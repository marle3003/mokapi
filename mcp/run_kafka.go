package mcp

import (
	"fmt"
	"mokapi/kafka"
	"mokapi/media"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/kafka/store"
	"mokapi/runtime"
	"slices"
	"strings"
)

type Kafka struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Brokers []Broker `json:"brokers"`

	info   *runtime.KafkaInfo
	client *store.Client
}

type Broker struct {
	Name        string `json:"name"`
	Host        string `json:"url"`
	Description string `json:"description,omitempty"`
}

type TopicSummary struct {
	Name    string `json:"name"`
	Title   string `json:"title,omitempty"`
	Summary string `json:"description,omitempty"`
}

type Topic struct {
	TopicSummary
	Description string            `json:"description,omitempty"`
	Partitions  []*KafkaPartition `json:"partitions"`

	Operations []KafkaOperation `json:"operations,omitempty"`

	info   *runtime.KafkaInfo
	client *store.Client
}

type KafkaPartition struct {
	Index  int   `json:"index"`
	Offset int64 `json:"offset"`
}

type KafkaOperation struct {
	Action      string         `json:"action"`
	Title       string         `json:"title"`
	Summary     string         `json:"summary,omitempty"`
	Description string         `json:"description,omitempty"`
	Messages    []KafkaMessage `json:"messages,omitempty"`
}

type KafkaMessage struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Summary     string `json:"summary,omitempty"`
	Description string `json:"description,omitempty"`
	ContentType string `json:"contentType"`
	Payload     any    `json:"payload,omitempty"`
	Key         any    `json:"key,omitempty"`
	Headers     any    `json:"headers,omitempty"`
}

type KafkaRecord struct {
	Offset  int64             `json:"offset"`
	Key     string            `json:"key"`
	Value   string            `json:"value"`
	Headers map[string]string `json:"headers,omitempty"`
}

func (m *mokapi) getKafkaApi(name string) any {
	for _, api := range m.app.Kafka.List() {
		if api.Info.Name == name {
			client := store.NewClient(api.Store, m.app.Monitor.Kafka)
			client.ClientId = "mokapi-mcp"

			result := &Kafka{
				Name:   name,
				Type:   "kafka",
				info:   api,
				client: client,
			}
			for it := api.Servers.Iter(); it.Next(); {
				b := it.Value()
				if b.Value == nil {
					continue
				}
				result.Brokers = append(result.Brokers, Broker{
					Name:        it.Key(),
					Host:        b.Value.Host,
					Description: b.Value.Description,
				})
			}

			return result
		}
	}
	return nil
}

func (k *Kafka) GetTopics() []TopicSummary {
	var topics []TopicSummary
	for name, c := range k.info.Channels {
		if c.Value == nil {
			continue
		}
		topics = append(topics, TopicSummary{
			Name:    name,
			Title:   c.Value.Title,
			Summary: c.Value.Summary,
		})
	}
	slices.SortStableFunc(topics, func(a, b TopicSummary) int {
		return strings.Compare(a.Name, b.Name)
	})
	return topics
}

func (k *Kafka) GetTopic(name string) (Topic, error) {
	ch, ok := k.info.Channels[name]
	if !ok || ch.Value == nil {
		return Topic{}, fmt.Errorf("topic '%s' not found", name)
	}

	t := Topic{
		TopicSummary: TopicSummary{
			Name:    name,
			Title:   ch.Value.Title,
			Summary: ch.Value.Summary,
		},
		Description: ch.Value.Description,
		info:        k.info,
		client:      k.client,
	}

	topic := k.info.Store.Topic(name)
	if topic == nil {
		return Topic{}, fmt.Errorf("topic '%s' not found", name)
	}
	for _, p := range topic.Partitions {
		t.Partitions = append(t.Partitions, &KafkaPartition{
			Index:  p.Index,
			Offset: p.Offset(),
		})
	}

	for _, op := range k.info.Operations {
		if op.Value == nil {
			continue
		}
		if op.Value.Channel.Value != ch.Value {
			continue
		}

		result := KafkaOperation{
			Action:      op.Value.Action,
			Title:       op.Value.Title,
			Summary:     op.Value.Summary,
			Description: op.Value.Description,
		}

		if len(op.Value.Messages) > 0 {
			for _, msg := range op.Value.Messages {
				if msg.Value == nil {
					continue
				}
				result.Messages = append(result.Messages, getKafkaMessages(msg))
			}
		} else {
			for _, msg := range ch.Value.Messages {
				if msg.Value == nil {
					continue
				}
				result.Messages = append(result.Messages, getKafkaMessages(msg))
			}
		}

		t.Operations = append(t.Operations, result)
	}

	slices.SortStableFunc(t.Operations, func(a, b KafkaOperation) int {
		r := strings.Compare(a.Action, b.Action)
		if r != 0 {
			return r
		}
		return strings.Compare(a.Title, b.Title)
	})

	return t, nil
}

func (t *Topic) Produce(partition int, value any, key string, headers map[string]string) error {
	var h []store.RecordHeader
	for hk, hv := range headers {
		h = append(h, store.RecordHeader{Name: hk, Value: hv})
	}

	result, err := t.client.Write(t.Name, []store.Record{
		{
			Key:       key,
			Value:     value,
			Partition: partition,
			Headers:   h,
		},
	}, media.ContentType{})

	if err != nil {
		return err
	}

	if len(result) > 0 && result[0].Error != "" {
		return fmt.Errorf("%s\nTo create a valid payload:\n1. Select a message from operation.messages\n2. Generate example data:\n\n   const value = mokapi.fake(message.payload)\n\n3. Modify only the required fields if needed.", result[0].Error)
	}

	// update JS topic and partition
	topic := t.info.Store.Topic(t.Name)
	if topic == nil {
		return fmt.Errorf("topic '%s' not found", t.Name)
	}
	p := topic.Partition(partition)
	if p == nil {
		return fmt.Errorf("partition '%s' not found", t.Name)
	}
	for _, pt := range t.Partitions {
		if pt.Index == p.Index {
			pt.Offset = p.Offset()
		}
	}

	return nil
}

func (t *Topic) Consume(partition int, startOffset int64, limit int) ([]KafkaRecord, error) {
	topic := t.info.Store.Topic(t.Name)
	if topic == nil {
		return nil, fmt.Errorf("topic '%s' not found", t.Name)
	}
	p := topic.Partition(partition)
	if p == nil {
		return nil, fmt.Errorf("partition '%d' not found", partition)
	}

	var records []KafkaRecord
	offset := startOffset
	n := 0
	for {
		if offset >= p.Tail || n >= limit {
			return records, nil
		}
		seg := p.GetSegment(offset)
		if seg == nil {
			return records, nil
		}

		for seg.Contains(offset) {
			r := seg.Record(offset)

			result := KafkaRecord{
				Offset: offset,
				Key:    kafka.BytesToString(r.Key),
				Value:  kafka.BytesToString(r.Value),
			}
			if r.Headers != nil {
				result.Headers = make(map[string]string)
				for _, h := range r.Headers {
					result.Headers[h.Key] = string(h.Value)
				}
			}

			records = append(records, result)

			n++
			offset++

			if n >= limit {
				return records, nil
			}
		}
	}
}

func getKafkaMessages(msg *asyncapi3.MessageRef) KafkaMessage {
	m := KafkaMessage{
		Name:        msg.Value.Name,
		Title:       msg.Value.Title,
		Summary:     msg.Value.Summary,
		Description: msg.Value.Description,
		ContentType: msg.Value.ContentType,
		Headers:     msg.Value.Headers,
	}
	if msg.Value.Payload != nil {
		m.Payload = msg.Value.Payload.Value
	}
	if msg.Value.Bindings.Kafka.Key != nil {
		m.Key = msg.Value.Bindings.Kafka.Key
	}
	return m
}
