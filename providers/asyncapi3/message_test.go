package asyncapi3_test

import (
	"encoding/json"
	"io"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/providers/asyncapi3"
	jsonSchema "mokapi/schema/json/schema"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
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
				msf := cfg.Payload.Value.(*asyncapi3.MultiSchemaFormat)
				require.Equal(t, "application/vnd.oai.openapi;version=3.0.0", msf.Format)
				require.Equal(t, "foo.json#foo", msf.Schema.Ref)
			},
		},
		{
			name: "payload schema",
			data: `{ "payload": { "type": "string" }}`,
			test: func(t *testing.T, cfg *asyncapi3.Message, err error) {
				require.NoError(t, err)
				require.Equal(t, "string", cfg.Payload.Value.(*jsonSchema.Schema).Type.String())
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			logrus.SetOutput(io.Discard)
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
				msf := cfg.Payload.Value.(*asyncapi3.MultiSchemaFormat)
				require.Equal(t, "application/vnd.oai.openapi;version=3.0.0", msf.Format)
				require.NotNil(t, msf.Schema)
				require.Equal(t, "foo.json#foo", msf.Schema.Ref)
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
				require.Equal(t, "string", cfg.Payload.Value.(*jsonSchema.Schema).Type.String())
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
				Operations: map[string]*asyncapi3.OperationRef{
					"foo": {Value: &asyncapi3.Operation{
						Messages: []*asyncapi3.MessageRef{
							{Value: &asyncapi3.Message{}},
						},
					}},
				},
			},
			test: func(t *testing.T, cfg *asyncapi3.Config, err error) {
				require.NoError(t, err)
				op := cfg.Operations["foo"]
				require.NotNil(t, op)
				require.Len(t, op.Value.Messages, 1)
				require.Equal(t, "application/json", op.Value.Messages[0].Value.ContentType)
			},
		},
		{
			name: "use defaultContentType",
			cfg: &asyncapi3.Config{
				DefaultContentType: "avro/binary",
				Operations: map[string]*asyncapi3.OperationRef{
					"foo": {Value: &asyncapi3.Operation{
						Messages: []*asyncapi3.MessageRef{
							{Value: &asyncapi3.Message{}},
						},
					}},
				},
			},
			test: func(t *testing.T, cfg *asyncapi3.Config, err error) {
				require.NoError(t, err)
				op := cfg.Operations["foo"]
				require.NotNil(t, op)
				require.Len(t, op.Value.Messages, 1)
				require.Equal(t, "avro/binary", op.Value.Messages[0].Value.ContentType)
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

func TestMessage_Marshal(t *testing.T) {
	testcases := []struct {
		name string
		s    *asyncapi3.MessageRef
		json string
		yaml string
		test func(t *testing.T, json, yaml string, err error)
	}{
		{
			name: "ref",
			s: &asyncapi3.MessageRef{
				Reference: dynamic.Reference[*asyncapi3.MessageRef]{
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
			s: &asyncapi3.MessageRef{
				Value: &asyncapi3.Message{
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

func TestMessageTrait_Marshal(t *testing.T) {
	testcases := []struct {
		name string
		s    *asyncapi3.MessageTraitRef
		json string
		yaml string
		test func(t *testing.T, json, yaml string, err error)
	}{
		{
			name: "ref",
			s: &asyncapi3.MessageTraitRef{
				Reference: dynamic.Reference[*asyncapi3.MessageTraitRef]{
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
			s: &asyncapi3.MessageTraitRef{
				Value: &asyncapi3.MessageTrait{
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
