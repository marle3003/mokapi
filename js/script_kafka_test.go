package js_test

import (
	r "github.com/stretchr/testify/require"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/jstest"
	"testing"
	"time"
)

func TestScript_Kafka_Produce(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, host *enginetest.Host)
	}{
		{
			name: "set topic",
			test: func(t *testing.T, host *enginetest.Host) {
				host.KafkaClientTest.ProduceFunc = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					r.Equal(t, "foo", args.Topic)
					r.Equal(t, "", args.Cluster)
					return &common.KafkaProduceResult{}, nil
				}

				s, err := jstest.New(jstest.WithSource(
					`import { produce } from 'mokapi/kafka'
						 export default function() {
						  	return produce({ topic: 'foo' })
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
			},
		},
		{
			name: "set topic and cluster",
			test: func(t *testing.T, host *enginetest.Host) {
				host.KafkaClientTest.ProduceFunc = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					r.Equal(t, "foo", args.Topic)
					r.Equal(t, "bar", args.Cluster)
					return &common.KafkaProduceResult{}, nil
				}

				s, err := jstest.New(jstest.WithSource(
					`import { produce } from 'mokapi/kafka'
						 export default function() {
  							return produce({ topic: 'foo', cluster: 'bar' })
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
			},
		},
		{
			name: "set cluster to null",
			test: func(t *testing.T, host *enginetest.Host) {
				host.KafkaClientTest.ProduceFunc = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					r.Equal(t, "", args.Cluster)
					return &common.KafkaProduceResult{}, nil
				}

				s, err := jstest.New(jstest.WithSource(
					`import { produce } from 'mokapi/kafka'
						 export default function() {
  							return produce({ cluster: null })
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
			},
		},
		{
			name: "set key, value and partition",
			test: func(t *testing.T, host *enginetest.Host) {
				host.KafkaClientTest.ProduceFunc = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					msg := args.Messages[0]
					r.Equal(t, "key", msg.Key)
					r.Equal(t, "value", msg.Data)
					r.Equal(t, 2, msg.Partition)
					return &common.KafkaProduceResult{}, nil
				}

				s, err := jstest.New(jstest.WithSource(
					`import { produce } from 'mokapi/kafka'
						 export default function() {
						  	return produce({ value: 'value', key: 'key', partition: 2 })
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
			},
		},
		{
			name: "set key, value and partition",
			test: func(t *testing.T, host *enginetest.Host) {
				host.KafkaClientTest.ProduceFunc = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					msg := args.Messages[0]
					r.Equal(t, "key", msg.Key)
					r.Equal(t, "value", msg.Data)
					r.Equal(t, 2, msg.Partition)
					return &common.KafkaProduceResult{}, nil
				}

				s, err := jstest.New(jstest.WithSource(
					`import { produce } from 'mokapi/kafka'
						 export default function() {
						  	return produce({ value: 'value', key: 'key', partition: 2 })
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
			},
		},
		{
			name: "set headers",
			test: func(t *testing.T, host *enginetest.Host) {
				host.KafkaClientTest.ProduceFunc = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					msg := args.Messages[0]
					r.Equal(t, map[string]interface{}{"foo": "bar"}, msg.Headers)
					return &common.KafkaProduceResult{}, nil
				}

				s, err := jstest.New(jstest.WithSource(
					`import { produce } from 'mokapi/kafka'
						 export default function() {
						  	return produce({ headers: { foo: 'bar' } })
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
			},
		},
		{
			name: "use messages",
			test: func(t *testing.T, host *enginetest.Host) {
				host.KafkaClientTest.ProduceFunc = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					msg := args.Messages[0]
					r.Equal(t, "key1", msg.Key)
					r.Equal(t, []byte("hello world"), msg.Value)
					r.Equal(t, map[string]interface{}{"system-id": "foo"}, msg.Headers)
					r.Equal(t, msg.Partition, 12)
					return &common.KafkaProduceResult{}, nil
				}

				s, err := jstest.New(jstest.WithSource(
					`import { produce } from 'mokapi/kafka'
						 export default function() {
						  	return produce({ messages: [{ key: 'key1', value: 'hello world', headers: { 'system-id': 'foo' }, partition: 12 }] })
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
			},
		},
		{
			name: "use messages with data",
			test: func(t *testing.T, host *enginetest.Host) {
				host.KafkaClientTest.ProduceFunc = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					msg := args.Messages[0]
					r.Equal(t, "hello world", msg.Data)
					return &common.KafkaProduceResult{}, nil
				}

				s, err := jstest.New(jstest.WithSource(
					`import { produce } from 'mokapi/kafka'
						 export default function() {
						  	return produce({ messages: [{ data: 'hello world' }] })
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
			},
		},
		{
			name: "result",
			test: func(t *testing.T, host *enginetest.Host) {
				host.KafkaClientTest.ProduceFunc = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					return &common.KafkaProduceResult{
						Cluster: "Cluster",
						Topic:   "Topic",
						Messages: []common.KafkaMessageResult{
							{
								Key:       "foo",
								Value:     "bar",
								Offset:    3451345,
								Partition: 99,
							},
						},
					}, nil
				}

				s, err := jstest.New(jstest.WithSource(
					`import { produce } from 'mokapi/kafka'
						 export default function() {
						  	return produce()
						 }`),
					js.WithHost(host))
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
			name: "default retry",
			test: func(t *testing.T, host *enginetest.Host) {
				host.KafkaClientTest.ProduceFunc = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					r.Equal(t, 30000*time.Millisecond, args.Retry.MaxRetryTime)
					r.Equal(t, 200*time.Millisecond, args.Retry.InitialRetryTime)
					r.Equal(t, 5, args.Retry.Retries)
					return &common.KafkaProduceResult{}, nil
				}

				s, err := jstest.New(jstest.WithSource(
					`import { produce } from 'mokapi/kafka'
						 export default function() {
						  	return produce({})
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
			},
		},
		{
			name: "set retry using number",
			test: func(t *testing.T, host *enginetest.Host) {
				host.KafkaClientTest.ProduceFunc = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					r.Equal(t, 1000*time.Millisecond, args.Retry.MaxRetryTime)
					r.Equal(t, time.Duration(0), args.Retry.InitialRetryTime)
					r.Equal(t, 100, args.Retry.Retries)
					return &common.KafkaProduceResult{}, nil
				}

				s, err := jstest.New(jstest.WithSource(
					`import { produce } from 'mokapi/kafka'
						 export default function() {
						  	return produce({ retry: { maxRetryTime: 1000, initialRetryTime: 0, retries: 100 } })
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
			},
		},
		{
			name: "set retry using string",
			test: func(t *testing.T, host *enginetest.Host) {
				host.KafkaClientTest.ProduceFunc = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					r.Equal(t, 30*time.Second, args.Retry.MaxRetryTime)
					r.Equal(t, 200*time.Millisecond, args.Retry.InitialRetryTime)
					r.Equal(t, 100, args.Retry.Retries)
					return &common.KafkaProduceResult{}, nil
				}

				s, err := jstest.New(jstest.WithSource(
					`import { produce } from 'mokapi/kafka'
						 export default function() {
						  	return produce({ retry: { maxRetryTime: '30s', initialRetryTime: '200ms', retries: 100 } })
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
			},
		},
		{
			name: "set retry using invalid type",
			test: func(t *testing.T, host *enginetest.Host) {
				host.KafkaClientTest.ProduceFunc = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					return &common.KafkaProduceResult{}, nil
				}

				s, err := jstest.New(jstest.WithSource(
					`import { produce } from 'mokapi/kafka'
						 export default function() {
						  	return produce({ retry: { maxRetryTime: [12], initialRetryTime: '200ms', retries: 100 } })
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				err = s.Run()
				r.EqualError(t, err, "type []interface {} for maxRetryTime not supported at mokapi/js/kafka.(*Module).Produce-fm (native)")
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			host := &enginetest.Host{
				KafkaClientTest: &enginetest.KafkaClient{},
			}

			tc.test(t, host)
		})
	}
}

func TestKafkaModule_Produce_DeprecatedAttributes(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, host *enginetest.Host)
	}{
		{
			name: "key",
			test: func(t *testing.T, host *enginetest.Host) {
				host.KafkaClientTest.ProduceFunc = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					return &common.KafkaProduceResult{}, nil
				}
				var warn string
				host.WarnFunc = func(args ...interface{}) {
					warn = args[0].(string)
				}

				s, err := jstest.New(jstest.WithSource(
					`import { produce } from 'mokapi/kafka'
						 export default function() {
						  	return produce({ key: 'foo' })
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
				r.Equal(t, "DEPRECATED: 'key' should not be used anymore: check https://mokapi.io/docs/javascript-api/mokapi-kafka/produceargs for more info in test host", warn)
			},
		},
		{
			name: "value",
			test: func(t *testing.T, host *enginetest.Host) {
				host.KafkaClientTest.ProduceFunc = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					return &common.KafkaProduceResult{}, nil
				}
				var warn string
				host.WarnFunc = func(args ...interface{}) {
					warn = args[0].(string)
				}

				s, err := jstest.New(jstest.WithSource(
					`import { produce } from 'mokapi/kafka'
						 export default function() {
						  	return produce({ value: 'foo' })
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
				r.Equal(t, "DEPRECATED: 'value' should not be used anymore: check https://mokapi.io/docs/javascript-api/mokapi-kafka/produceargs for more info in test host", warn)
			},
		},
		{
			name: "headers",
			test: func(t *testing.T, host *enginetest.Host) {
				host.KafkaClientTest.ProduceFunc = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					return &common.KafkaProduceResult{}, nil
				}
				var warn string
				host.WarnFunc = func(args ...interface{}) {
					warn = args[0].(string)
				}

				s, err := jstest.New(jstest.WithSource(
					`import { produce } from 'mokapi/kafka'
						 export default function() {
						  	return produce({ headers: { 'foo': 'bar' } })
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
				r.Equal(t, "DEPRECATED: 'headers' should not be used anymore: check https://mokapi.io/docs/javascript-api/mokapi-kafka/produceargs for more info in test host", warn)
			},
		},
		{
			name: "partition",
			test: func(t *testing.T, host *enginetest.Host) {
				host.KafkaClientTest.ProduceFunc = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					return &common.KafkaProduceResult{}, nil
				}
				var warn string
				host.WarnFunc = func(args ...interface{}) {
					warn = args[0].(string)
				}

				s, err := jstest.New(jstest.WithSource(
					`import { produce } from 'mokapi/kafka'
						 export default function() {
						  	return produce({ partition: 1 })
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
				r.Equal(t, "DEPRECATED: 'partition' should not be used anymore: check https://mokapi.io/docs/javascript-api/mokapi-kafka/produceargs for more info in test host", warn)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			host := &enginetest.Host{
				KafkaClientTest: &enginetest.KafkaClient{},
			}

			tc.test(t, host)
		})
	}
}
