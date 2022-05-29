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
			"cluster should be foo",
			func(t *testing.T, host *testHost) {
				host.kafkaClient.produce = func(cluster, topic string, partition int, key, value interface{}, headers map[string]interface{}) (interface{}, interface{}, error) {
					r.Equal(t, "foo", cluster)
					return nil, nil, nil
				}

				s, err := New("",
					`
import kafka from 'kafka'
export default function() {
  var s = kafka.produce({cluster: 'foo'})
return s
}`,
					host)
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
			},
		},
		{
			"topic should be foo",
			func(t *testing.T, host *testHost) {
				host.kafkaClient.produce = func(cluster, topic string, partition int, key, value interface{}, headers map[string]interface{}) (interface{}, interface{}, error) {
					r.Equal(t, "foo", topic)
					return nil, nil, nil
				}

				s, err := New("",
					`
import kafka from 'kafka'
export default function() {
  var s = kafka.produce({topic: 'foo'})
return s
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
