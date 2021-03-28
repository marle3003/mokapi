package kafka

import (
	"fmt"
	"mokapi/config/dynamic/openapi"
	"mokapi/models/media"
	"mokapi/providers/encoding"
	"mokapi/providers/pipeline/lang/types"
)

type ProducerStep struct {
	types.AbstractStep
	topics     map[string]*topic
	g          *openapi.Generator
	addMessage func(*topic, []byte, []byte) error
}

func newProducerStep(topics map[string]*topic, addMessage func(*topic, []byte, []byte) error) *ProducerStep {
	return &ProducerStep{
		topics:     topics,
		g:          openapi.NewGenerator(),
		addMessage: addMessage,
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
			i := e.step.g.New(t.payload)
			contentType := media.ParseContentType(t.contentType)
			b, err := encode(i, t.payload, contentType)
			if err != nil {
				return nil, err
			}
			data = b
		} else {
			data = []byte(e.Message)
		}

		var key []byte
		if len(e.Key) == 0 {
			k := e.step.g.New(t.key)
			key = []byte(fmt.Sprintf("%v", k))
		} else {
			key = []byte(e.Key)
		}

		return nil, e.step.addMessage(t, key, data)
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
