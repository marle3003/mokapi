package asyncapi3_test

import (
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/providers/asyncapi3"
	"testing"
)

func TestCorrelationId(t *testing.T) {
	b := []byte(`asyncapi: 3.0.0
channels:
  foo:
    messages:
      bar:
        correlationId:
          description: Data from message payload used as correlation ID
          location: $message.payload#/sentAt
`)
	var cfg *asyncapi3.Config
	err := yaml.Unmarshal(b, &cfg)
	require.NoError(t, err)

	msg := cfg.Channels["foo"].Value.Messages["bar"].Value
	require.Equal(t, "Data from message payload used as correlation ID", msg.CorrelationId.Value.Description)
	require.Equal(t, "$message.payload#/sentAt", msg.CorrelationId.Value.Location)
}

func TestCorrelationId_Ref(t *testing.T) {
	b := []byte(`asyncapi: 3.0.0
channels:
  foo:
    messages:
      bar:
        correlationId:
          $ref: '#/components/correlationIds/test'
components:
  correlationIds:
    test:
      description: Data from message payload used as correlation ID
      location: $message.payload#/sentAt
`)
	var cfg *asyncapi3.Config
	err := yaml.Unmarshal(b, &cfg)
	require.NoError(t, err)

	err = cfg.Parse(&dynamic.Config{Data: cfg}, &dynamictest.Reader{})
	require.NoError(t, err)

	msg := cfg.Channels["foo"].Value.Messages["bar"].Value
	require.Equal(t, "Data from message payload used as correlation ID", msg.CorrelationId.Value.Description)
	require.Equal(t, "$message.payload#/sentAt", msg.CorrelationId.Value.Location)
}
