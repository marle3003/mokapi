package kafka

import (
	"fmt"
	"mokapi/providers/pipeline/lang/types"
	"mokapi/server/kafka/protocol"
	"time"
)

type ProducerStep struct {
	types.AbstractStep
	topics map[string]*topic
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
		record := &protocol.RecordBatch{
			Records: []protocol.Record{
				{
					Offset:  0,
					Time:    time.Now(),
					Key:     nil,
					Value:   []byte(e.Message),
					Headers: nil,
				},
			},
		}
		err := t.addRecord(0, record)
		return nil, err
	}
}

//func producer(topic *topic, contentType *media.ContentType, schema string, stop chan bool) {
//	defer log.Info("producer stopped")
//
//	for {
//		select {
//		case <-stop:
//			return
//		case <-time.After(2 * time.Second):
//			data := getRandomObject(schema)
//			b, err := encode(data, schema, contentType)
//			if err != nil {
//				log.Errorf("Error in producing data: %v", err)
//			} else {
//
//				record := &protocol.RecordBatch{
//					Attributes: 0,
//					ProducerId: 0,
//					Records: []protocol.Record{
//						{
//							Offset:  0,
//							Time:    time.Now(),
//							Key:     nil,
//							Value:   b,
//							Headers: nil,
//						},
//					},
//				}
//				topic.partitions[0].log.Append(record)
//			}
//		}
//	}
//}
//
//func encode(data interface{}, schema *schemas.Schema, contentType *media.ContentType) ([]byte, error) {
//	switch contentType.Subtype {
//	case "json":
//		return encoding.MarshalJSON(data, schema)
//	case "xml", "rss+xml":
//		return encoding.MarshalXML(data, schema)
//	default:
//		if s, ok := data.(string); ok {
//			return []byte(s), nil
//		}
//		return nil, fmt.Errorf("unspupported encoding for content type %v", contentType)
//	}
//}
//
//func getRandomObject(schema *schemas.Schema) interface{} {
//	if schema.Type == "object" {
//		obj := make(map[string]interface{})
//		for name, propSchema := range schema.Properties {
//			value := getRandomObject(propSchema)
//			obj[name] = value
//		}
//		return obj
//	} else if schema.Type == "array" {
//		length := rand.Intn(5)
//		obj := make([]interface{}, length)
//		for i := range obj {
//			obj[i] = getRandomObject(schema.Items)
//		}
//		return obj
//	} else {
//		if len(schema.Faker) > 0 {
//			switch schema.Faker {
//			case "numbers.uint32":
//				return gofakeit.Uint32()
//			default:
//				return gofakeit.Generate(fmt.Sprintf("{%s}", schema.Faker))
//			}
//		} else if schema.Type == "integer" {
//			return gofakeit.Int32()
//		} else if schema.Type == "string" {
//			return gofakeit.Lexify("???????????????")
//		}
//	}
//	return nil
//}
