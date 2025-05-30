package asyncApi

import (
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"testing"
)

func TestKafkaBinding(t *testing.T) {
	src := `asyncapi: 2.0
info:
  title: A sample AsyncApi Kafka streaming api
  version: 1.0.0
servers:
  broker:
    url: 127.0.0.1:19092
    protocol: kafka
    bindings:
      kafka:
        log.retention.bytes: 1000
`
	c := &Config{}
	err := yaml.Unmarshal([]byte(src), &c)
	require.NoError(t, err)
	require.Equal(t, int64(1000), c.Servers["broker"].Value.Bindings.Kafka.LogRetentionBytes)
}

func TestChannelBindings(t *testing.T) {
	testcases := []struct {
		name  string
		input string
		test  func(t *testing.T, c *Config, err error)
	}{
		{
			name: "simple",
			input: `asyncapi: '2.0.0'
info:
  title: Kafka Testserver
channels:
  message:
    bindings:
      kafka:
        partitions: 1
        segment.ms: 30000
        topicConfiguration:
          retention.bytes: 1000000000
`,
			test: func(t *testing.T, c *Config, err error) {
				require.NoError(t, err)
				require.Equal(t, 1, c.Channels["message"].Value.Bindings.Kafka.Partitions)
				require.Equal(t, int64(30000), c.Channels["message"].Value.Bindings.Kafka.SegmentMs)
				require.Equal(t, int64(1000000000), c.Channels["message"].Value.Bindings.Kafka.RetentionBytes)
			},
		},
		{
			name: "default values",
			input: `asyncapi: '2.0.0'
info:
  title: Kafka Testserver
channels:
  message:
    bindings:
      kafka:
        segment.ms: 30000
        topicConfiguration:
          retention.bytes: 1000000000
`,
			test: func(t *testing.T, c *Config, err error) {
				require.NoError(t, err)
				require.Equal(t, 1, c.Channels["message"].Value.Bindings.Kafka.Partitions)
				require.Equal(t, true, c.Channels["message"].Value.Bindings.Kafka.ValueSchemaValidation)
				require.Equal(t, true, c.Channels["message"].Value.Bindings.Kafka.KeySchemaValidation)
			},
		},
		{
			name: "default values",
			input: `asyncapi: '2.0.0'
info:
  title: Kafka Testserver
components:
  messages:
    foo:
      bindings:
        kafka:
          schemaIdLocation: payload
          schemaIdPayloadEncoding: '4'
`,
			test: func(t *testing.T, c *Config, err error) {
				require.NoError(t, err)
				require.Equal(t, "payload", c.Components.Messages["foo"].Bindings.Kafka.SchemaIdLocation)
				require.Equal(t, "4", c.Components.Messages["foo"].Bindings.Kafka.SchemaIdPayloadEncoding)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c := &Config{}
			err := yaml.Unmarshal([]byte(tc.input), &c)
			tc.test(t, c, err)
		})
	}
}
