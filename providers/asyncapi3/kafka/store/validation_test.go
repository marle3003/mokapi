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
			name: "value invalid",
			cfg: asyncapi3test.NewConfig(
				asyncapi3test.WithChannel("foo",
					asyncapi3test.WithMessage("foo",
						asyncapi3test.WithPayload(schematest.New("string")),
						asyncapi3test.WithContentType("application/json"),
					),
				),
			),
			test: func(t *testing.T, s *store.Store) {
				p := s.Topic("foo").Partition(0)
				_, batch, err := p.Write(kafka.RecordBatch{
					Records: []*kafka.Record{
						{
							Value: kafka.NewBytes([]byte("123")),
						},
					},
				})
				require.EqualError(t, err, "validation error: invalid message: error count 1:\n- #/type: invalid type, expected string but got number")
				require.Len(t, batch, 1)
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
		{
			name: "header not valid",
			cfg: asyncapi3test.NewConfig(
				asyncapi3test.WithChannel("foo",
					asyncapi3test.WithMessage("foo",
						asyncapi3test.WithHeaders(schematest.New("object",
							schematest.WithProperty("foo", schematest.New("integer", schematest.WithMinimum(100))))),
					),
				),
			),
			test: func(t *testing.T, s *store.Store) {
				p := s.Topic("foo").Partition(0)
				_, batch, err := p.Write(kafka.RecordBatch{
					Records: []*kafka.Record{
						{
							Headers: []kafka.RecordHeader{{Key: "foo", Value: []byte{64}}},
						},
					},
				})
				require.EqualError(t, err, "validation error: invalid key: error count 1:\n- #/minimum: integer 64 is less than minimum value of 100")
				require.Len(t, batch, 1)
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

func TestValidation_Header(t *testing.T) {
	testcases := []struct {
		name string
		cfg  *asyncapi3.Config
		test func(t *testing.T, s *store.Store)
	}{
		{
			name: "header int but schema is string",
			cfg: asyncapi3test.NewConfig(
				asyncapi3test.WithChannel("foo",
					asyncapi3test.WithMessage("foo",
						asyncapi3test.WithHeaders(schematest.New("object",
							schematest.WithProperty("foo", schematest.New("string")))),
					),
				),
			),
			test: func(t *testing.T, s *store.Store) {
				p := s.Topic("foo").Partition(0)
				_, batch, err := p.Write(kafka.RecordBatch{
					Records: []*kafka.Record{
						{
							Headers: []kafka.RecordHeader{{Key: "foo", Value: []byte{1, 0, 0, 0}}},
						},
					},
				})
				require.NoError(t, err)
				require.Len(t, batch, 0)

				e := events.GetEvents(events.NewTraits())
				require.Len(t, e, 1)
				require.Equal(t, "\u0001\u0000\u0000\u0000", e[0].Data.(*store.KafkaLog).Headers["foo"].Value)
			},
		},
		{
			name: "header int",
			cfg: asyncapi3test.NewConfig(
				asyncapi3test.WithChannel("foo",
					asyncapi3test.WithMessage("foo",
						asyncapi3test.WithHeaders(schematest.New("object",
							schematest.WithProperty("foo", schematest.New("integer")))),
					),
				),
			),
			test: func(t *testing.T, s *store.Store) {
				p := s.Topic("foo").Partition(0)
				_, batch, err := p.Write(kafka.RecordBatch{
					Records: []*kafka.Record{
						{
							Headers: []kafka.RecordHeader{{Key: "foo", Value: []byte{1, 0, 0, 0}}},
						},
					},
				})
				require.NoError(t, err)
				require.Len(t, batch, 0)

				e := events.GetEvents(events.NewTraits())
				require.Len(t, e, 1)
				require.Equal(t, "1", e[0].Data.(*store.KafkaLog).Headers["foo"].Value)
			},
		},
		{
			name: "header not defined in schema",
			cfg: asyncapi3test.NewConfig(
				asyncapi3test.WithChannel("foo",
					asyncapi3test.WithMessage("foo"),
				),
			),
			test: func(t *testing.T, s *store.Store) {
				p := s.Topic("foo").Partition(0)
				_, batch, err := p.Write(kafka.RecordBatch{
					Records: []*kafka.Record{
						{
							Headers: []kafka.RecordHeader{{Key: "foo", Value: []byte{1, 0, 0, 0}}},
						},
					},
				})
				require.NoError(t, err)
				require.Len(t, batch, 0)

				e := events.GetEvents(events.NewTraits())
				require.Len(t, e, 1)
				require.Equal(t, []byte{1, 0, 0, 0}, e[0].Data.(*store.KafkaLog).Headers["foo"].Binary)
			},
		},
		{
			name: "float",
			cfg: asyncapi3test.NewConfig(
				asyncapi3test.WithChannel("foo",
					asyncapi3test.WithMessage("foo",
						asyncapi3test.WithHeaders(schematest.New("object",
							schematest.WithProperty("foo", schematest.New("number")))),
					),
				),
			),
			test: func(t *testing.T, s *store.Store) {
				p := s.Topic("foo").Partition(0)
				_, batch, err := p.Write(kafka.RecordBatch{
					Records: []*kafka.Record{
						{
							Headers: []kafka.RecordHeader{{Key: "foo", Value: []byte{119, 16, 73, 64}}},
						},
					},
				})
				require.NoError(t, err)
				require.Len(t, batch, 0)

				e := events.GetEvents(events.NewTraits())
				require.Len(t, e, 1)
				require.Equal(t, "3.141629934310913", e[0].Data.(*store.KafkaLog).Headers["foo"].Value)
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
