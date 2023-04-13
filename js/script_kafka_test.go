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
				host.kafkaClient.produce = func(args *common.KafkaProduceArgs) (interface{}, interface{}, error) {
					r.Equal(t, "foo", args.Topic)
					r.Equal(t, "", args.Cluster)
					return nil, nil, nil
				}

				s, err := New("",
					`import {produce} from 'kafka'
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
				host.kafkaClient.produce = func(args *common.KafkaProduceArgs) (interface{}, interface{}, error) {
					r.Equal(t, "foo", args.Topic)
					r.Equal(t, "bar", args.Cluster)
					return nil, nil, nil
				}

				s, err := New("",
					`import {produce} from 'kafka'
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
				host.kafkaClient.produce = func(args *common.KafkaProduceArgs) (interface{}, interface{}, error) {
					r.Equal(t, "", args.Cluster)
					return nil, nil, nil
				}

				s, err := New("",
					`import {produce} from 'kafka'
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
				host.kafkaClient.produce = func(args *common.KafkaProduceArgs) (interface{}, interface{}, error) {
					r.Equal(t, "key", args.Key)
					r.Equal(t, "value", args.Value)
					r.Equal(t, 2, args.Partition)
					return nil, nil, nil
				}

				s, err := New("",
					`import {produce} from 'kafka'
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
				host.kafkaClient.produce = func(args *common.KafkaProduceArgs) (interface{}, interface{}, error) {
					r.Equal(t, "key", args.Key)
					r.Equal(t, "value", args.Value)
					r.Equal(t, 2, args.Partition)
					return nil, nil, nil
				}

				s, err := New("",
					`import {produce} from 'kafka'
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
				host.kafkaClient.produce = func(args *common.KafkaProduceArgs) (interface{}, interface{}, error) {
					r.Equal(t, map[string]interface{}{"foo": "bar"}, args.Headers)
					return nil, nil, nil
				}

				s, err := New("",
					`import {produce} from 'kafka'
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
				host.kafkaClient.produce = func(args *common.KafkaProduceArgs) (interface{}, interface{}, error) {
					r.Equal(t, map[string]interface{}{"foo": "bar"}, args.Headers)
					return nil, nil, nil
				}

				s, err := New("",
					`import {produce} from 'kafka'
						 export default function() {
						  	return produce({headers: {foo: 'bar'}})
						 }`,
					host)
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

			tc.f(t, host)
		})
	}
}
