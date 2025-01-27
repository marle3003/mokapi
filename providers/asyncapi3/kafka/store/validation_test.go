package store_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/engine/enginetest"
	"mokapi/kafka"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/providers/asyncapi3/kafka/store"
	"mokapi/runtime/events"
	"mokapi/schema/json/schema/schematest"
	"testing"
)

func TestValidation(t *testing.T) {
	testcases := []struct {
		name string
		cfg  *asyncapi3.Config
		test func(t *testing.T, s *store.Store)
	}{
		{
			name: "value validation disabled",
			cfg: asyncapi3test.NewConfig(
				asyncapi3test.WithChannel("foo",
					asyncapi3test.WithMessage("foo",
						asyncapi3test.WithPayload(schematest.New("string")),
						asyncapi3test.WithContentType("application/json"),
					),
					asyncapi3test.WithKafkaChannelBinding(asyncapi3.TopicBindings{
						Partitions:            1,
						ValueSchemaValidation: false,
					}),
				),
			),
			test: func(t *testing.T, s *store.Store) {
				p := s.Topic("foo").Partition(0)
				_, batch, err := p.Write(kafka.RecordBatch{
					Records: []*kafka.Record{
						{
							Value: kafka.NewBytes([]byte("")),
						},
					},
				})
				require.NoError(t, err)
				require.Len(t, batch, 0)
			},
		},
		{
			name: "key validation disabled",
			cfg: asyncapi3test.NewConfig(
				asyncapi3test.WithChannel("foo",
					asyncapi3test.WithMessage("foo",
						asyncapi3test.WithKey(schematest.New("integer")),
					),
					asyncapi3test.WithKafkaChannelBinding(asyncapi3.TopicBindings{
						Partitions:          1,
						KeySchemaValidation: false,
					}),
				),
			),
			test: func(t *testing.T, s *store.Store) {
				p := s.Topic("foo").Partition(0)
				_, batch, err := p.Write(kafka.RecordBatch{
					Records: []*kafka.Record{
						{
							Key: kafka.NewBytes([]byte("foo")),
						},
					},
				})
				require.NoError(t, err)
				require.Len(t, batch, 0)
			},
		},
		{
			name: "schemaId in payload",
			cfg: asyncapi3test.NewConfig(
				asyncapi3test.WithChannel("foo",
					asyncapi3test.WithMessage("foo",
						asyncapi3test.WithPayload(schematest.New("string")),
						asyncapi3test.WithContentType("application/json"),
						asyncapi3test.WithKafkaMessageBinding(asyncapi3.KafkaMessageBinding{
							SchemaIdLocation:        "payload",
							SchemaIdPayloadEncoding: "4",
						}),
					),
				),
			),
			test: func(t *testing.T, s *store.Store) {
				p := s.Topic("foo").Partition(0)
				_, batch, err := p.Write(kafka.RecordBatch{
					Records: []*kafka.Record{
						{
							Value: kafka.NewBytes([]byte{0, 0, 0, 0, 1, '"', 'f', 'o', 'o', '"'}),
						},
					},
				})
				require.NoError(t, err)
				require.Len(t, batch, 0)

				e := events.GetEvents(events.NewTraits())
				require.Len(t, e, 1)
				require.Equal(t, 1, e[0].Data.(*store.KafkaLog).SchemaId)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			defer events.Reset()
			events.SetStore(5, events.NewTraits().WithNamespace("kafka"))

			s := store.New(tc.cfg, enginetest.NewEngine())
			tc.test(t, s)
		})
	}
}
