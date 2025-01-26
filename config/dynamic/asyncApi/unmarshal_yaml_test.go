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
	src := `asyncapi: '2.0.0'
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
`
	c := &Config{}
	err := yaml.Unmarshal([]byte(src), &c)
	require.NoError(t, err)
	require.Equal(t, 1, c.Channels["message"].Value.Bindings.Kafka.Partitions)
	require.Equal(t, int64(30000), c.Channels["message"].Value.Bindings.Kafka.SegmentMs)
	require.Equal(t, int64(1000000000), c.Channels["message"].Value.Bindings.Kafka.RetentionBytes)

	src = `asyncapi: '2.0.0'
info:
  title: Kafka Testserver
channels:
  message:
    bindings:
      kafka:
        segment.ms: 30000
        topicConfiguration:
          retention.bytes: 1000000000
`
	c = &Config{}
	err = yaml.Unmarshal([]byte(src), &c)
	require.NoError(t, err)
	require.Equal(t, 1, c.Channels["message"].Value.Bindings.Kafka.Partitions)
}
