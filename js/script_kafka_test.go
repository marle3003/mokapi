package js

import (
	r "github.com/stretchr/testify/require"
	"mokapi/engine/common"
	"testing"
)

func TestScript_Kafka_Produce(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T, host *testHost)
	}{
		{
			"set topic",
			func(t *testing.T, host *testHost) {
				host.kafkaClient.produce = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					r.Equal(t, "foo", args.Topic)
					r.Equal(t, "", args.Cluster)
					return &common.KafkaProduceResult{}, nil
				}

				s, err := New("",
					`import {produce} from 'mokapi/kafka'
						 export default function() {
						  	return produce({topic: 'foo'})
						 }`,
					host)
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
			},
		},
		{
			"set topic and cluster",
			func(t *testing.T, host *testHost) {
				host.kafkaClient.produce = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					r.Equal(t, "foo", args.Topic)
					r.Equal(t, "bar", args.Cluster)
					return &common.KafkaProduceResult{}, nil
				}

				s, err := New("",
					`import {produce} from 'mokapi/kafka'
						 export default function() {
  							return produce({topic: 'foo', cluster: 'bar'})
						 }`,
					host)
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
			},
		},
		{
			"set cluster to null",
			func(t *testing.T, host *testHost) {
				host.kafkaClient.produce = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					r.Equal(t, "", args.Cluster)
					return &common.KafkaProduceResult{}, nil
				}

				s, err := New("",
					`import {produce} from 'mokapi/kafka'
						 export default function() {
  							return produce({cluster: null})
						 }`,
					host)
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
			},
		},
		{
			"set key, value and partition",
			func(t *testing.T, host *testHost) {
				host.kafkaClient.produce = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					r.Equal(t, "key", args.Key)
					r.Equal(t, "value", args.Value)
					r.Equal(t, 2, args.Partition)
					return &common.KafkaProduceResult{}, nil
				}

				s, err := New("",
					`import {produce} from 'mokapi/kafka'
						 export default function() {
						  	return produce({value: 'value', key: 'key', partition: 2})
						 }`,
					host)
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
			},
		},
		{
			"set key, value and partition",
			func(t *testing.T, host *testHost) {
				host.kafkaClient.produce = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					r.Equal(t, "key", args.Key)
					r.Equal(t, "value", args.Value)
					r.Equal(t, 2, args.Partition)
					return &common.KafkaProduceResult{}, nil
				}

				s, err := New("",
					`import {produce} from 'mokapi/kafka'
						 export default function() {
						  	return produce({value: 'value', key: 'key', partition: 2})
						 }`,
					host)
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
			},
		},
		{
			"set headers",
			func(t *testing.T, host *testHost) {
				host.kafkaClient.produce = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					r.Equal(t, map[string]interface{}{"foo": "bar"}, args.Headers)
					return &common.KafkaProduceResult{}, nil
				}

				s, err := New("",
					`import {produce} from 'mokapi/kafka'
						 export default function() {
						  	return produce({headers: {foo: 'bar'}})
						 }`,
					host)
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
			},
		},
		{
			"set timeout",
			func(t *testing.T, host *testHost) {
				host.kafkaClient.produce = func(args *common.KafkaProduceArgs) (*common.KafkaProduceResult, error) {
					r.Equal(t, map[string]interface{}{"foo": "bar"}, args.Headers)
					return &common.KafkaProduceResult{}, nil
				}

				s, err := New("",
					`import {produce} from 'mokapi/kafka'
						 export default function() {
						  	return produce({headers: {foo: 'bar'}})
						 }`,
					host)
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
			},
		},
		{
			"result",
			func(t *testing.T, host *testHost) {
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

				s, err := New("",
					`import {produce} from 'mokapi/kafka'
						 export default function() {
						  	return produce()
						 }`,
					host)
				r.NoError(t, err)
				v, err := s.RunDefault()
				r.NoError(t, err)
				result := v.Export().(ProduceResult)
				r.Equal(t, "Cluster", result.Cluster)
				r.Equal(t, "Topic", result.Topic)
				r.Equal(t, 99, result.Partition)
				r.Equal(t, int64(3451345), result.Offset)
				r.Equal(t, "foo", result.Key)
				r.Equal(t, "bar", result.Value)
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

			tc.f(t, host)
		})
	}
}
