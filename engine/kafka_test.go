package engine

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/asyncApi/asyncapitest"
	"mokapi/config/dynamic/asyncApi/kafka/store"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/kafka"
	"mokapi/providers/openapi/schema/schematest"
	"mokapi/runtime"
	"net/url"
	"testing"
)

func TestKafkaClient_Produce(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, s *store.Store, c *kafkaClient)
	}{
		{
			name: "random key",
			test: func(t *testing.T, s *store.Store, c *kafkaClient) {
				result, err := c.Produce(&common.KafkaProduceArgs{Topic: "foo", Cluster: "foo"})
				require.NoError(t, err)
				require.NotNil(t, result)
				b, kerr := s.Topic("foo").Partition(0).Read(0, 1000)
				require.Equal(t, kafka.None, kerr)
				require.NotNil(t, b)
				require.Equal(t, fmt.Sprintf("%v", result.Key), string(readBytes(b.Records[0].Key)))
			},
		},
		{
			name: "multiple clusters",
			test: func(t *testing.T, s *store.Store, c *kafkaClient) {
				for i := 0; i < 10; i++ {
					c.app.AddKafka(getConfig(asyncapitest.NewConfig(asyncapitest.WithInfo(fmt.Sprintf("x%v", i), "", ""))), enginetest.NewEngine())
				}

				_, err := c.Produce(&common.KafkaProduceArgs{Topic: "foo"})
				require.NoError(t, err)
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
			app := runtime.New()
			info := app.AddKafka(getConfig(config), enginetest.NewEngine())
			c := newKafkaClient(app)
			tc.test(t, info.Store, c)
		})
	}
}

func readBytes(b kafka.Bytes) []byte {
	buf := new(bytes.Buffer)
	buf.ReadFrom(b)
	return buf.Bytes()
}

func getConfig(c *asyncApi.Config) *dynamic.Config {
	u, _ := url.Parse("foo.bar")
	cfg := &dynamic.Config{Data: c}
	cfg.Info.Url = u
	return cfg
}
