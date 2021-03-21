package kafka

import (
	"fmt"
	"mokapi/config/dynamic/openapi"
	"mokapi/models/media"
	"mokapi/providers/encoding"
	"mokapi/providers/pipeline/lang/types"
	"mokapi/server/kafka/protocol"
	"time"
)

type ProducerStep struct {
	types.AbstractStep
	topics map[string]*topic
	g      *openapi.Generator
}

func newProducerStep(topics map[string]*topic) *ProducerStep {
	return &ProducerStep{
		topics: topics,
		g:      openapi.NewGenerator(),
	}
}

type ProducerStepExecution struct {
	Topic   string `step:"topic,required"`
	Key     string `step:"key"`
	Message string `step:"message"`

	step *ProducerStep
}

func (s *ProducerStep) Start() types.StepExecution {
	return &ProducerStepExecution{step: s}
}

func (e *ProducerStepExecution) Run(_ types.StepContext) (interface{}, error) {
	if len(e.Topic) == 0 {
		return nil, fmt.Errorf("missing topic")
	}
	if t, ok := e.step.topics[e.Topic]; !ok {
		return nil, fmt.Errorf("topic %q not found", e.Topic)
	} else {
		var data []byte
		if len(e.Message) == 0 {
			i := e.step.g.New(t.config.Payload)
			contentType := media.ParseContentType(t.config.ContentType)
			b, err := encode(i, t.config.Payload, contentType)
			if err != nil {
				return nil, err
			}
			data = b
		} else {
			data = []byte(e.Message)
		}

		var key []byte
		if len(e.Key) == 0 {
			k := e.step.g.New(t.config.Bindings.Kafka.Key)
			key = []byte(fmt.Sprintf("%v", k))
		} else {
			key = []byte(e.Key)
		}

		record := &protocol.RecordBatch{
			Records: []protocol.Record{
				{
					Offset:  0,
					Time:    time.Now(),
					Key:     key,
					Value:   data,
					Headers: nil,
				},
			},
		}
		err := t.addRecord(0, record)
		return nil, err
	}
}

func encode(data interface{}, schema *openapi.SchemaRef, contentType *media.ContentType) ([]byte, error) {
	switch contentType.Subtype {
	case "json":
		return encoding.MarshalJSON(data, schema)
	case "xml", "rss+xml":
		return encoding.MarshalXML(data, schema)
	default:
		if s, ok := data.(string); ok {
			return []byte(s), nil
		}
		return nil, fmt.Errorf("unspupported encoding for content type %v", contentType)
	}
}
