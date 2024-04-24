package js

import (
	r "github.com/stretchr/testify/require"
	"mokapi/config/static"
	"mokapi/engine/common"
	"testing"
)

func TestScript_Kafka_Produce(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, host *testHost)
	}{
		{
			name: "set topic",
			test: func(t *testing.T, host *testHost) {
				host.kafkaClient.produce = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					r.Equal(t, "foo", args.Topic)
					r.Equal(t, "", args.Cluster)
					return &common.KafkaProduceResult{}, nil
				}

				s, err := New(newScript("",
					`import { produce } from 'mokapi/kafka'
						 export default function() {
						  	return produce({ topic: 'foo' })
						 }`),
					host, static.JsConfig{})
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
			},
		},
		{
			name: "set topic and cluster",
			test: func(t *testing.T, host *testHost) {
				host.kafkaClient.produce = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					r.Equal(t, "foo", args.Topic)
					r.Equal(t, "bar", args.Cluster)
					return &common.KafkaProduceResult{}, nil
				}

				s, err := New(newScript("",
					`import { produce } from 'mokapi/kafka'
						 export default function() {
  							return produce({ topic: 'foo', cluster: 'bar' })
						 }`),
					host, static.JsConfig{})
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
			},
		},
		{
			name: "set cluster to null",
			test: func(t *testing.T, host *testHost) {
				host.kafkaClient.produce = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					r.Equal(t, "", args.Cluster)
					return &common.KafkaProduceResult{}, nil
				}

				s, err := New(newScript("",
					`import { produce } from 'mokapi/kafka'
						 export default function() {
  							return produce({ cluster: null })
						 }`),
					host, static.JsConfig{})
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
			},
		},
		{
			name: "set key, value and partition",
			test: func(t *testing.T, host *testHost) {
				host.kafkaClient.produce = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					msg := args.Messages[0]
					r.Equal(t, "key", msg.Key)
					r.Equal(t, "value", msg.Data)
					r.Equal(t, 2, msg.Partition)
					return &common.KafkaProduceResult{}, nil
				}

				s, err := New(newScript("",
					`import { produce } from 'mokapi/kafka'
						 export default function() {
						  	return produce({ value: 'value', key: 'key', partition: 2 })
						 }`),
					host, static.JsConfig{})
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
			},
		},
		{
			name: "set key, value and partition",
			test: func(t *testing.T, host *testHost) {
				host.kafkaClient.produce = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					msg := args.Messages[0]
					r.Equal(t, "key", msg.Key)
					r.Equal(t, "value", msg.Data)
					r.Equal(t, 2, msg.Partition)
					return &common.KafkaProduceResult{}, nil
				}

				s, err := New(newScript("",
					`import { produce } from 'mokapi/kafka'
						 export default function() {
						  	return produce({ value: 'value', key: 'key', partition: 2 })
						 }`),
					host, static.JsConfig{})
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
			},
		},
		{
			name: "set headers",
			test: func(t *testing.T, host *testHost) {
				host.kafkaClient.produce = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					msg := args.Messages[0]
					r.Equal(t, map[string]interface{}{"foo": "bar"}, msg.Headers)
					return &common.KafkaProduceResult{}, nil
				}

				s, err := New(newScript("",
					`import { produce } from 'mokapi/kafka'
						 export default function() {
						  	return produce({ headers: { foo: 'bar' } })
						 }`),
					host, static.JsConfig{})
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
			},
		},
		{
			name: "set timeout",
			test: func(t *testing.T, host *testHost) {
				host.kafkaClient.produce = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					msg := args.Messages[0]
					r.Equal(t, map[string]interface{}{"foo": "bar"}, msg.Headers)
					return &common.KafkaProduceResult{}, nil
				}

				s, err := New(newScript("",
					`import { produce } from 'mokapi/kafka'
						 export default function() {
						  	return produce({ headers: { foo: 'bar' } })
						 }`),
					host, static.JsConfig{})
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
			},
		},
		{
			name: "use messages",
			test: func(t *testing.T, host *testHost) {
				host.kafkaClient.produce = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					msg := args.Messages[0]
					r.Equal(t, "key1", msg.Key)
					r.Equal(t, []byte("hello world"), msg.Value)
					r.Equal(t, map[string]interface{}{"system-id": "foo"}, msg.Headers)
					r.Equal(t, msg.Partition, 12)
					return &common.KafkaProduceResult{}, nil
				}

				s, err := New(newScript("",
					`import { produce } from 'mokapi/kafka'
						 export default function() {
						  	return produce({ messages: [{ key: 'key1', value: 'hello world', headers: { 'system-id': 'foo' }, partition: 12 }] })
						 }`),
					host, static.JsConfig{})
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
			},
		},
		{
			name: "use messages with data",
			test: func(t *testing.T, host *testHost) {
				host.kafkaClient.produce = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					msg := args.Messages[0]
					r.Equal(t, "hello world", msg.Data)
					return &common.KafkaProduceResult{}, nil
				}

				s, err := New(newScript("",
					`import { produce } from 'mokapi/kafka'
						 export default function() {
						  	return produce({ messages: [{ data: 'hello world' }] })
						 }`),
					host, static.JsConfig{})
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
			},
		},
		{
			name: "result",
			test: func(t *testing.T, host *testHost) {
				host.kafkaClient.produce = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					return &common.KafkaProduceResult{
						Cluster: "Cluster",
						Topic:   "Topic",
						Messages: []common.KafkaProducedMessage{
							{
								Key:       "foo",
								Value:     "bar",
								Offset:    3451345,
								Partition: 99,
							},
						},
					}, nil
				}

				s, err := New(newScript("",
					`import { produce } from 'mokapi/kafka'
						 export default function() {
						  	return produce()
						 }`),
					host, static.JsConfig{})
				r.NoError(t, err)
				v, err := s.RunDefault()
				r.NoError(t, err)
				result := v.Export().(*common.KafkaProduceResult)
				r.Equal(t, "Cluster", result.Cluster)
				r.Equal(t, "Topic", result.Topic)
				r.Equal(t, 99, result.Messages[0].Partition)
				r.Equal(t, int64(3451345), result.Messages[0].Offset)
				r.Equal(t, "foo", result.Messages[0].Key)
				r.Equal(t, "bar", result.Messages[0].Value)
			},
		},
		{
			name: "using deprecated module",
			test: func(t *testing.T, host *testHost) {
				host.kafkaClient.produce = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					r.Equal(t, "foo", args.Topic)
					r.Equal(t, "", args.Cluster)
					return &common.KafkaProduceResult{}, nil
				}

				s, err := New(newScript("",
					`import { produce } from 'kafka'
						 export default function() {
						  	return produce({ topic: 'foo' })
						 }`),
					host, static.JsConfig{})
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			host := &testHost{
				kafkaClient: &kafkaClient{},
			}

			tc.test(t, host)
		})
	}
}
