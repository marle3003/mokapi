package asyncapi3_test

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
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
			data: `{ "payload": { "format": "application/vnd.oai.openapi;version=3.0.0", "schema": { "$ref": "foo.json#foo" } }}`,
			test: func(t *testing.T, cfg *asyncapi3.Message, err error) {
				require.NoError(t, err)
				require.Equal(t, "application/vnd.oai.openapi;version=3.0.0", cfg.Payload.Value.Format)
				s := cfg.Payload.Value.Schema.(*schema.Ref)
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
				require.Equal(t, "string", cfg.Payload.Value.Schema.(*jsonSchema.Ref).Type())
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
