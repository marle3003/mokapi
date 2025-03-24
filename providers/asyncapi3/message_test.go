package asyncapi3_test

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/openapi/schema"
	jsonSchema "mokapi/schema/json/schema"
	"testing"
)

func TestMessage_UnmarshalJSON(t *testing.T) {
	testcases := []struct {
		name string
		data string
		test func(t *testing.T, cfg *asyncapi3.Message, err error)
	}{
		{
			name: "payload OpenAPI ref",
			data: `{ "payload": { "schemaFormat": "application/vnd.oai.openapi;version=3.0.0", "schema": { "$ref": "foo.json#foo" } }}`,
			test: func(t *testing.T, cfg *asyncapi3.Message, err error) {
				require.NoError(t, err)
				require.Equal(t, "application/vnd.oai.openapi;version=3.0.0", cfg.Payload.Value.Format)
				s := cfg.Payload.Value.Schema.(*schema.Schema)
				require.NotNil(t, s)
				require.Equal(t, "foo.json#foo", s.Ref)
			},
		},
		{
			name: "payload schema",
			data: `{ "payload": { "type": "string" }}`,
			test: func(t *testing.T, cfg *asyncapi3.Message, err error) {
				require.NoError(t, err)
				require.Equal(t, "", cfg.Payload.Value.Format)
				require.Equal(t, "string", cfg.Payload.Value.Schema.(*jsonSchema.Schema).Type.String())
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var msg *asyncapi3.Message
			err := json.Unmarshal([]byte(tc.data), &msg)
			tc.test(t, msg, err)
		})
	}
}

func TestMessage_UnmarshalYAML(t *testing.T) {
	testcases := []struct {
		name string
		data string
		test func(t *testing.T, cfg *asyncapi3.Message, err error)
	}{
		{
			name: "payload OpenAPI ref",
			data: `
payload:
  schemaFormat: application/vnd.oai.openapi;version=3.0.0
  schema:
    $ref: 'foo.json#foo'
`,
			test: func(t *testing.T, cfg *asyncapi3.Message, err error) {
				require.NoError(t, err)
				require.Equal(t, "application/vnd.oai.openapi;version=3.0.0", cfg.Payload.Value.Format)
				s := cfg.Payload.Value.Schema.(*schema.Schema)
				require.NotNil(t, s)
				require.Equal(t, "foo.json#foo", s.Ref)
			},
		},
		{
			name: "payload schema",
			data: `
payload:
  type: string
`,
			test: func(t *testing.T, cfg *asyncapi3.Message, err error) {
				require.NoError(t, err)
				require.Equal(t, "", cfg.Payload.Value.Format)
				require.Equal(t, "string", cfg.Payload.Value.Schema.(*jsonSchema.Schema).Type.String())
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var msg *asyncapi3.Message
			err := yaml.Unmarshal([]byte(tc.data), &msg)
			tc.test(t, msg, err)
		})
	}
}

func TestMessage_Parse(t *testing.T) {
	testcases := []struct {
		name string
		cfg  *asyncapi3.Config
		test func(t *testing.T, cfg *asyncapi3.Config, err error)
	}{
		{
			name: "no content type set",
			cfg: &asyncapi3.Config{
				Components: &asyncapi3.Components{Messages: map[string]*asyncapi3.MessageRef{
					"foo": {Value: &asyncapi3.Message{}},
				}},
			},
			test: func(t *testing.T, cfg *asyncapi3.Config, err error) {
				require.NoError(t, err)
				require.Equal(t, "application/json", cfg.Components.Messages["foo"].Value.ContentType)
			},
		},
		{
			name: "use defaultContentType",
			cfg: &asyncapi3.Config{
				DefaultContentType: "avro/binary",
				Components: &asyncapi3.Components{Messages: map[string]*asyncapi3.MessageRef{
					"foo": {Value: &asyncapi3.Message{}},
				}},
			},
			test: func(t *testing.T, cfg *asyncapi3.Config, err error) {
				require.NoError(t, err)
				require.Equal(t, "avro/binary", cfg.Components.Messages["foo"].Value.ContentType)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.cfg.Parse(&dynamic.Config{Data: tc.cfg}, &dynamictest.Reader{})
			tc.test(t, tc.cfg, err)
		})
	}
}
