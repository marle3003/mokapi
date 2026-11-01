package parser_test

import (
	"encoding/json"
	"mokapi/schema/json/parser"
	"mokapi/schema/json/schema"
	"testing"

	"github.com/stretchr/testify/require"
)

const allOfWithNestedOneOf = `{
  "allOf" : [ {
    "type" : "object",
    "properties" : { "status" : { "type" : "string" } },
    "required" : [ "status" ]
  }, {
    "type" : "object",
    "oneOf" : [ {
      "type" : "object",
      "properties" : {
        "output" : {
          "type" : "object",
          "properties" : { "array" : { "type" : "array", "minItems" : 1 } }
        }
      }
    } ]
  } ]
}`

func TestParser_AllOf_NestedComposition(t *testing.T) {
	testcases := []struct {
		name   string
		schema string
		data   string
		test   func(t *testing.T, v any, err error)
	}{
		{
			name:   "allOf with nested oneOf: valid",
			schema: allOfWithNestedOneOf,
			data:   `{"status":"ok","output":{"array":[1]}}`,
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]any{"status": "ok", "output": map[string]any{"array": []any{float64(1)}}}, v)
			},
		},
		{
			name:   "allOf with nested oneOf: minItems violated",
			schema: allOfWithNestedOneOf,
			data:   `{"status":"ok","output":{"array":[]}}`,
			test: func(t *testing.T, v any, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "oneOf")
				require.Contains(t, err.Error(), "minItems")
			},
		},
		{
			name:   "allOf with nested anyOf",
			schema: `{"allOf":[{"type":"object","anyOf":[{"required":["a"]},{"required":["b"]}]}]}`,
			data:   `{"c":1}`,
			test: func(t *testing.T, v any, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "anyOf")
			},
		},
		{
			name:   "allOf with nested allOf",
			schema: `{"allOf":[{"type":"object","allOf":[{"required":["a"]}]}]}`,
			data:   `{"c":1}`,
			test: func(t *testing.T, v any, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "required")
			},
		},
		{
			name:   "anyOf with nested oneOf",
			schema: `{"anyOf":[{"type":"object","oneOf":[{"required":["a"]}]}]}`,
			data:   `{"c":1}`,
			test: func(t *testing.T, v any, err error) {
				require.Error(t, err)
			},
		},
		{
			name:   "oneOf with nested anyOf",
			schema: `{"oneOf":[{"type":"object","anyOf":[{"required":["a"]}]}]}`,
			data:   `{"c":1}`,
			test: func(t *testing.T, v any, err error) {
				require.Error(t, err)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			var s *schema.Schema
			require.NoError(t, json.Unmarshal([]byte(tc.schema), &s))
			var data any
			require.NoError(t, json.Unmarshal([]byte(tc.data), &data))

			p := &parser.Parser{Schema: s}
			v, err := p.Parse(data)
			tc.test(t, v, err)
		})
	}
}
