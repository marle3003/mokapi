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
					r.Equal(t, "key", args.Key)
					r.Equal(t, "value", args.Value)
					r.Equal(t, 2, args.Partition)
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
					r.Equal(t, "key", args.Key)
					r.Equal(t, "value", args.Value)
					r.Equal(t, 2, args.Partition)
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
					r.Equal(t, map[string]interface{}{"foo": "bar"}, args.Headers)
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
					r.Equal(t, map[string]interface{}{"foo": "bar"}, args.Headers)
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
			name: "result",
			test: func(t *testing.T, host *testHost) {
				host.kafkaClient.produce = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					return &common.KafkaProduceResult{
						Cluster:   "Cluster",
						Topic:     "Topic",
						Partition: 99,
						Offset:    3451345,
						Key:       "foo",
						Value:     "bar",
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
				r.Equal(t, 99, result.Partition)
				r.Equal(t, int64(3451345), result.Offset)
				r.Equal(t, "foo", result.Key)
				r.Equal(t, "bar", result.Value)
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
