package asyncapi3_test

import (
	"encoding/json"
	"mokapi/providers/asyncapi3"
	"mokapi/schema/json/schema/schematest"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestChannelBindings_Websocket_Marshal(t *testing.T) {
	testcases := []struct {
		name string
		s    *asyncapi3.ChannelBindings
		json string
		yaml string
		test func(t *testing.T, json, yaml string, err error)
	}{
		{
			name: "websocket default",
			s: &asyncapi3.ChannelBindings{
				// required to set default to be empty
				Kafka:     asyncapi3.TopicBindings{Partitions: 1, KeySchemaValidation: true, ValueSchemaValidation: true},
				Websocket: asyncapi3.WebsocketChannelBindings{},
			},
			test: func(t *testing.T, json, yaml string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{}`, json)
				require.Equal(t, "{}\n", yaml)
			},
		},
		{
			name: "method",
			s: &asyncapi3.ChannelBindings{
				// required to set default to be empty
				Kafka: asyncapi3.TopicBindings{Partitions: 1, KeySchemaValidation: true, ValueSchemaValidation: true},
				Websocket: asyncapi3.WebsocketChannelBindings{
					Method: "POST",
				},
			},
			test: func(t *testing.T, json, yaml string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"websocket":{"method":"POST"}}`, json)
				require.Equal(t, "websocket:\n    method: POST\n", yaml)
			},
		},
		{
			name: "query",
			s: &asyncapi3.ChannelBindings{
				// required to set default to be empty
				Kafka: asyncapi3.TopicBindings{Partitions: 1, KeySchemaValidation: true, ValueSchemaValidation: true},
				Websocket: asyncapi3.WebsocketChannelBindings{
					Query: schematest.New("string"),
				},
			},
			test: func(t *testing.T, json, yaml string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"websocket":{"query":{"type":"string"}}}`, json)
				require.Equal(t, "websocket:\n    query:\n        type: string\n", yaml)
			},
		},
		{
			name: "header",
			s: &asyncapi3.ChannelBindings{
				// required to set default to be empty
				Kafka: asyncapi3.TopicBindings{Partitions: 1, KeySchemaValidation: true, ValueSchemaValidation: true},
				Websocket: asyncapi3.WebsocketChannelBindings{
					Headers: schematest.New("string"),
				},
			},
			test: func(t *testing.T, json, yaml string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"websocket":{"headers":{"type":"string"}}}`, json)
				require.Equal(t, "websocket:\n    headers:\n        type: string\n", yaml)
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
