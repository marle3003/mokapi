package store_test

import (
	"encoding/base64"
	"mokapi/engine/enginetest"
	"mokapi/media"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/providers/asyncapi3/kafka/store"
	"mokapi/runtime/events/eventstest"
	"mokapi/runtime/monitor"
	"mokapi/schema/json/schema/schematest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	testcases := []struct {
		name string
		cfg  *asyncapi3.Config
		test func(t *testing.T, s *store.Store, monitor *monitor.Kafka)
	}{
		{
			name: "topic does not exist",
			cfg:  &asyncapi3.Config{},
			test: func(t *testing.T, s *store.Store, monitor *monitor.Kafka) {
				c := store.NewClient(s, monitor)
				ct := media.ParseContentType("application/json")

				result, err := c.Write("foo", []store.Record{}, ct)
				require.EqualError(t, err, "topic not found")
				require.Nil(t, result)

				_, err = c.Read("foo", 1, 0, &ct)
				require.EqualError(t, err, "topic not found")
			},
		},
		{
			name: "partition does not exist",
			cfg: asyncapi3test.NewConfig(
				asyncapi3test.WithChannel("foo",
					asyncapi3test.WithKafkaChannelBinding(asyncapi3.TopicBindings{Partitions: 1}),
				),
			),
			test: func(t *testing.T, s *store.Store, monitor *monitor.Kafka) {
				c := store.NewClient(s, monitor)
				ct := media.ParseContentType("application/json")
				result, err := c.Write("foo", []store.Record{{
					Partition: 1,
				}}, ct)
				require.EqualError(t, err, "partition not found")
				require.Nil(t, result)

				_, err = c.Read("foo", 1, 0, &ct)
				require.EqualError(t, err, "partition not found")
			},
		},
		{
			name: "value as json and random partition",
			cfg: func() *asyncapi3.Config {
				msg := asyncapi3test.NewMessage(
					asyncapi3test.WithContentType("application/json"),
					asyncapi3test.WithPayload(
						schematest.New("object",
							schematest.WithProperty("foo", schematest.New("string")),
						),
					),
				)
				ch := asyncapi3test.NewChannel(asyncapi3test.UseMessage("foo", &asyncapi3.MessageRef{Value: msg}))

				return asyncapi3test.NewConfig(
					asyncapi3test.AddChannel("foo", ch),
					asyncapi3test.WithOperation("sendAction",
						asyncapi3test.WithOperationAction("send"),
						asyncapi3test.WithOperationChannel(ch),
						asyncapi3test.UseOperationMessage(msg),
					),
				)
			}(),
			test: func(t *testing.T, s *store.Store, monitor *monitor.Kafka) {
				c := store.NewClient(s, monitor)
				ct := media.ParseContentType("application/json")
				result, err := c.Write("foo", []store.Record{{
					Partition: -1,
					Value: map[string]interface{}{
						"foo": "foo",
					},
				}}, ct)
				require.NoError(t, err)
				require.Len(t, result, 1)
				require.Equal(t, 0, result[0].Partition)
				require.Equal(t, int64(0), result[0].Offset)
				require.Nil(t, result[0].Key)
				require.Equal(t, "{\"foo\":\"foo\"}", string(result[0].Value))
				require.Equal(t, "", result[0].Error)
			},
		},
		{
			name: "value unspecified",
			cfg: func() *asyncapi3.Config {
				msg := asyncapi3test.NewMessage(
					asyncapi3test.WithContentType("application/json"),
					asyncapi3test.WithPayload(
						schematest.New("object",
							schematest.WithProperty("foo", schematest.New("string")),
						),
					),
				)
				ch := asyncapi3test.NewChannel(asyncapi3test.UseMessage("foo", &asyncapi3.MessageRef{Value: msg}))

				return asyncapi3test.NewConfig(
					asyncapi3test.AddChannel("foo", ch),
					asyncapi3test.WithOperation("sendAction",
						asyncapi3test.WithOperationAction("send"),
						asyncapi3test.WithOperationChannel(ch),
						asyncapi3test.UseOperationMessage(msg),
					),
				)
			}(),
			test: func(t *testing.T, s *store.Store, monitor *monitor.Kafka) {
				c := store.NewClient(s, monitor)
				ct := media.ParseContentType("")
				result, err := c.Write("foo", []store.Record{{
					Value: map[string]interface{}{
						"foo": "foo",
					},
				}}, ct)
				require.NoError(t, err)
				require.Len(t, result, 1)
				require.Equal(t, "", result[0].Error)
				require.Equal(t, 0, result[0].Partition)
				require.Equal(t, int64(0), result[0].Offset)
				require.Nil(t, result[0].Key)
				require.Equal(t, "{\"foo\":\"foo\"}", string(result[0].Value))
			},
		},
		{
			name: "use []byte key and value",
			cfg: asyncapi3test.NewConfig(
				asyncapi3test.WithChannel("foo",
					asyncapi3test.WithMessage("foo",
						asyncapi3test.WithContentType("application/json"),
						asyncapi3test.WithPayload(
							schematest.New("object",
								schematest.WithProperty("foo", schematest.New("string")),
							),
						),
					),
				),
			),
			test: func(t *testing.T, s *store.Store, monitor *monitor.Kafka) {
				c := store.NewClient(s, monitor)
				ct := media.ParseContentType("")
				result, err := c.Write("foo", []store.Record{{
					Key:   []byte("12345"),
					Value: []byte(`{"foo":"bar"}`),
				}}, ct)
				require.NoError(t, err)
				require.Len(t, result, 1)
				require.Equal(t, "", result[0].Error)
				require.Equal(t, 0, result[0].Partition)
				require.Equal(t, int64(0), result[0].Offset)
				require.Equal(t, "12345", string(result[0].Key))
				require.Equal(t, `{"foo":"bar"}`, string(result[0].Value))

				ct = media.ParseContentType("application/json")
				records, err := c.Read("foo", 0, 0, &ct)
				require.NoError(t, err)
				require.Equal(t, []store.Record{
					{
						Key:       "12345",
						Value:     map[string]interface{}{"foo": "bar"},
						Headers:   nil,
						Partition: 0,
					},
				}, records)

				ct = media.ParseContentType("application/vnd.mokapi.kafka.binary+json")
				records, err = c.Read("foo", 0, 0, &ct)
				require.NoError(t, err)
				require.Equal(t, []store.Record{
					{
						Key:       "12345",
						Value:     "eyJmb28iOiJiYXIifQ==",
						Headers:   nil,
						Partition: 0,
					},
				}, records)
				b, err := base64.StdEncoding.DecodeString(records[0].Value.(string))
				require.NoError(t, err)
				require.Equal(t, `{"foo":"bar"}`, string(b))
			},
		},
		{
			name: "use string key and value",
			cfg: func() *asyncapi3.Config {
				msg := asyncapi3test.NewMessage(
					asyncapi3test.WithContentType("application/json"),
					asyncapi3test.WithPayload(
						schematest.New("object",
							schematest.WithProperty("foo", schematest.New("string")),
						),
					),
				)
				ch := asyncapi3test.NewChannel(asyncapi3test.UseMessage("foo", &asyncapi3.MessageRef{Value: msg}))

				return asyncapi3test.NewConfig(
					asyncapi3test.AddChannel("foo", ch),
					asyncapi3test.WithOperation("sendAction",
						asyncapi3test.WithOperationAction("send"),
						asyncapi3test.WithOperationChannel(ch),
						asyncapi3test.UseOperationMessage(msg),
					),
				)
			}(),
			test: func(t *testing.T, s *store.Store, monitor *monitor.Kafka) {
				c := store.NewClient(s, monitor)
				result, err := c.Write("foo", []store.Record{{
					Key:   "12345",
					Value: `{"foo":"bar"}`,
				}}, media.ParseContentType("application/vnd.mokapi.kafka.json+json"))
				require.NoError(t, err)
				require.Len(t, result, 1)
				require.Equal(t, "", result[0].Error)
				require.Equal(t, 0, result[0].Partition)
				require.Equal(t, int64(0), result[0].Offset)
				require.Equal(t, "12345", string(result[0].Key))
				require.Equal(t, `{"foo":"bar"}`, string(result[0].Value))
			},
		},
		{
			name: "use value as encoded base64",
			cfg: asyncapi3test.NewConfig(
				asyncapi3test.WithChannel("foo",
					asyncapi3test.WithMessage("foo",
						asyncapi3test.WithContentType("application/json"),
						asyncapi3test.WithPayload(
							schematest.New("object",
								schematest.WithProperty("foo", schematest.New("string")),
							),
						),
					),
				),
			),
			test: func(t *testing.T, s *store.Store, monitor *monitor.Kafka) {
				c := store.NewClient(s, monitor)
				ct := media.ParseContentType("application/vnd.mokapi.kafka.binary+json")
				result, err := c.Write("foo", []store.Record{{
					Key:   "12345",
					Value: "eyJmb28iOiJiYXIifQ==",
				}}, ct)
				require.NoError(t, err)
				require.Len(t, result, 1)
				require.Equal(t, 0, result[0].Partition)
				require.Equal(t, int64(0), result[0].Offset)
				require.Equal(t, "12345", string(result[0].Key))
				require.Equal(t, `{"foo":"bar"}`, string(result[0].Value))
				require.Equal(t, "", result[0].Error)
			},
		},
		{
			name: "key as number",
			cfg: func() *asyncapi3.Config {
				msg := asyncapi3test.NewMessage(
					asyncapi3test.WithContentType("application/json"),
					asyncapi3test.WithPayload(
						schematest.New("object",
							schematest.WithProperty("foo", schematest.New("string")),
						),
					),
				)
				ch := asyncapi3test.NewChannel(asyncapi3test.UseMessage("foo", &asyncapi3.MessageRef{Value: msg}))

				return asyncapi3test.NewConfig(
					asyncapi3test.AddChannel("foo", ch),
					asyncapi3test.WithOperation("sendAction",
						asyncapi3test.WithOperationAction("send"),
						asyncapi3test.WithOperationChannel(ch),
						asyncapi3test.UseOperationMessage(msg),
					),
				)
			}(),
			test: func(t *testing.T, s *store.Store, monitor *monitor.Kafka) {
				c := store.NewClient(s, monitor)
				result, err := c.Write("foo", []store.Record{{
					Key:   1234,
					Value: `{"foo":"bar"}`,
				}}, media.ParseContentType("application/vnd.mokapi.kafka.json+json"))
				require.NoError(t, err)
				require.Len(t, result, 1)
				require.Equal(t, "", result[0].Error)
				require.Equal(t, 0, result[0].Partition)
				require.Equal(t, int64(0), result[0].Offset)
				require.Equal(t, "1234", string(result[0].Key))
				require.Equal(t, `{"foo":"bar"}`, string(result[0].Value))
			},
		},
		{
			name: "read with unknown offset (-1) using 2 partitions",
			cfg: func() *asyncapi3.Config {
				msg := asyncapi3test.NewMessage(
					asyncapi3test.WithContentType("application/json"),
					asyncapi3test.WithPayload(
						schematest.New("object",
							schematest.WithProperty("foo", schematest.New("string")),
						),
					),
				)
				ch := asyncapi3test.NewChannel(
					asyncapi3test.UseMessage("foo", &asyncapi3.MessageRef{Value: msg}),
					asyncapi3test.WithKafkaChannelBinding(asyncapi3.TopicBindings{Partitions: 2}),
				)

				return asyncapi3test.NewConfig(
					asyncapi3test.AddChannel("foo", ch),
					asyncapi3test.WithOperation("sendAction",
						asyncapi3test.WithOperationAction("send"),
						asyncapi3test.WithOperationChannel(ch),
						asyncapi3test.UseOperationMessage(msg),
					),
				)
			}(),
			test: func(t *testing.T, s *store.Store, monitor *monitor.Kafka) {
				c := store.NewClient(s, monitor)
				result, err := c.Write("foo", []store.Record{{
					Key:       1234,
					Value:     `{"foo":"bar"}`,
					Partition: 1,
				}}, media.ParseContentType("application/vnd.mokapi.kafka.json+json"))
				require.NoError(t, err)
				require.Len(t, result, 1)
				require.Equal(t, "", result[0].Error)
				require.Equal(t, 1, result[0].Partition)
				require.Equal(t, int64(0), result[0].Offset)
				require.Equal(t, "1234", string(result[0].Key))
				require.Equal(t, `{"foo":"bar"}`, string(result[0].Value))

				ct := media.ParseContentType("application/json")
				records, err := c.Read("foo", 1, -1, &ct)
				require.NoError(t, err)
				require.Equal(t, []store.Record{
					{
						Key:       "1234",
						Value:     map[string]interface{}{"foo": "bar"},
						Headers:   nil,
						Partition: 1,
					},
				}, records)
			},
		},
		{
			name: "using header",
			cfg: func() *asyncapi3.Config {
				msg := asyncapi3test.NewMessage(
					asyncapi3test.WithContentType("application/json"),
					asyncapi3test.WithPayload(
						schematest.New("object",
							schematest.WithProperty("foo", schematest.New("string")),
						),
					),
				)
				ch := asyncapi3test.NewChannel(asyncapi3test.UseMessage("foo", &asyncapi3.MessageRef{Value: msg}))

				return asyncapi3test.NewConfig(
					asyncapi3test.AddChannel("foo", ch),
					asyncapi3test.WithOperation("sendAction",
						asyncapi3test.WithOperationAction("send"),
						asyncapi3test.WithOperationChannel(ch),
						asyncapi3test.UseOperationMessage(msg),
					),
				)
			}(),
			test: func(t *testing.T, s *store.Store, monitor *monitor.Kafka) {
				c := store.NewClient(s, monitor)
				result, err := c.Write("foo", []store.Record{{
					Key:       1234,
					Value:     `{"foo":"bar"}`,
					Headers:   []store.RecordHeader{{Name: "yuh", Value: "bar"}},
					Partition: 0,
				}}, media.ParseContentType("application/vnd.mokapi.kafka.json+json"))
				require.NoError(t, err)
				require.Equal(t, "", result[0].Error)
				require.Len(t, result, 1)
				require.Equal(t, 0, result[0].Partition)
				require.Equal(t, int64(0), result[0].Offset)
				require.Equal(t, "1234", string(result[0].Key))
				require.Equal(t, `{"foo":"bar"}`, string(result[0].Value))
				require.Equal(t, []store.RecordHeader{{Name: "yuh", Value: "bar"}}, result[0].Headers)

				ct := media.ParseContentType("application/json")
				records, err := c.Read("foo", 0, -1, &ct)
				require.NoError(t, err)
				require.Equal(t, []store.Record{
					{
						Key:       "1234",
						Value:     map[string]interface{}{"foo": "bar"},
						Headers:   []store.RecordHeader{{Name: "yuh", Value: "bar"}},
						Partition: 0,
					},
				}, records)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			s := store.New(tc.cfg, enginetest.NewEngine(), &eventstest.Handler{})
			m := monitor.NewKafka()
			tc.test(t, s, m)
		})
	}
}
