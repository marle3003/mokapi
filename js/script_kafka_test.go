package js

import (
	r "github.com/stretchr/testify/require"
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
				host.kafkaClient.produce = func(cluster, topic string, partition int, key, value interface{}, headers map[string]interface{}) (interface{}, interface{}, error) {
					r.Equal(t, "foo", topic)
					r.Equal(t, "", cluster)
					return nil, nil, nil
				}

				s, err := New("",
					`
import {produce} from 'kafka'
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
				host.kafkaClient.produce = func(cluster, topic string, partition int, key, value interface{}, headers map[string]interface{}) (interface{}, interface{}, error) {
					r.Equal(t, "foo", topic)
					r.Equal(t, "bar", cluster)
					return nil, nil, nil
				}

				s, err := New("",
					`
import {produce} from 'kafka'
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
			"set key, value and partition",
			func(t *testing.T, host *testHost) {
				host.kafkaClient.produce = func(cluster, topic string, partition int, key, value interface{}, headers map[string]interface{}) (interface{}, interface{}, error) {
					r.Equal(t, "key", key)
					r.Equal(t, "value", value)
					r.Equal(t, 2, partition)
					return nil, nil, nil
				}

				s, err := New("",
					`
import {produce} from 'kafka'
export default function() {
  return produce({value: 'value', key: 'key', partition: 2})
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
