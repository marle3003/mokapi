package asyncapi3_test

import (
	"encoding/json"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/providers/asyncapi3"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
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

func TestCorrelationId_Marshal(t *testing.T) {
	testcases := []struct {
		name string
		s    *asyncapi3.CorrelationIdRef
		json string
		yaml string
		test func(t *testing.T, json, yaml string, err error)
	}{
		{
			name: "ref",
			s: &asyncapi3.CorrelationIdRef{
				Reference: dynamic.Reference[*asyncapi3.CorrelationIdRef]{
					Ref: "/foo",
				},
			},
			test: func(t *testing.T, json, yaml string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"$ref":"/foo"}`, json)
				require.Equal(t, "$ref: /foo\n", yaml)
			},
		},
		{
			name: "value",
			s: &asyncapi3.CorrelationIdRef{
				Value: &asyncapi3.CorrelationId{
					Description: "foo",
				},
			},
			test: func(t *testing.T, json, yaml string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"description":"foo"}`, json)
				require.Equal(t, "description: foo\n", yaml)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			jb, err := json.Marshal(tc.s)
			if err != nil {
				tc.test(t, "", "", err)
			}
			yb, err := yaml.Marshal(tc.s)
			tc.test(t, string(jb), string(yb), err)
		})
	}
}
