package engine

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/asyncApi/asyncapitest"
	"mokapi/config/dynamic/asyncApi/kafka/store"
	"mokapi/config/dynamic/openapi/schema/schematest"
	"mokapi/engine/enginetest"
	"mokapi/kafka"
	"mokapi/runtime"
	"testing"
)

func TestKafkaClient_Produce(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T, s *store.Store, c *kafkaClient)
	}{
		{
			"random key",
			func(t *testing.T, s *store.Store, c *kafkaClient) {
				k, v, err := c.Produce("foo", "foo", -1, nil, nil, nil)
				require.NoError(t, err)
				require.NotNil(t, k)
				require.NotNil(t, v)
				b, kerr := s.Topic("foo").Partition(0).Read(0, 1000)
				require.Equal(t, kafka.None, kerr)
				require.NotNil(t, b)
				require.Equal(t, fmt.Sprintf("%v", k), string(readBytes(b.Records[0].Key)))
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			config := asyncapitest.NewConfig(
				asyncapitest.WithInfo("foo", "", ""),
				asyncapitest.WithChannel("foo",
					asyncapitest.WithSubscribeAndPublish(
						asyncapitest.WithMessage(
							asyncapitest.WithContentType("application/json"),
							asyncapitest.WithPayload(schematest.New("string")),
							asyncapitest.WithKey(schematest.New("string"))))))
			s := store.New(config, enginetest.NewEngine())
			app := runtime.New()
			app.AddKafka(config, s)
			c := newKafkaClient(app)
			tc.f(t, s, c)
		})
	}
}

func readBytes(b kafka.Bytes) []byte {
	buf := new(bytes.Buffer)
	buf.ReadFrom(b)
	return buf.Bytes()
}
