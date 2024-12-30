package schema_test

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"mokapi/schema/avro/schema"
	"testing"
)

func TestSchema_UnmarshalJSON(t *testing.T) {
	testcases := []struct {
		name  string
		input string
		test  func(t *testing.T, schema *schema.Schema, err error)
	}{
		{
			name:  "empty",
			input: "{}",
			test: func(t *testing.T, schema *schema.Schema, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:  "a JSON string, naming a defined type.",
			input: `"string"`,
			test: func(t *testing.T, schema *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, "string", schema.Type[0])
			},
		},
		{
			name:  "a JSON object",
			input: `{"type": "string"}`,
			test: func(t *testing.T, schema *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, "string", schema.Type[0])
			},
		},
		{
			name:  "a JSON array",
			input: `["string", "boolean"]`,
			test: func(t *testing.T, schema *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, "string", schema.Type[0])
				require.Equal(t, "boolean", schema.Type[1])
			},
		},
		{
			name:  "nested JSON object",
			input: `{ "type": { "type": "string" } }`,
			test: func(t *testing.T, schema *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, "string", schema.Schema[0].Type[0])
			},
		},
		{
			name:  "JSON object with array field",
			input: `{"type": "record", "fields": [{ "name": "list", "type": { "type": "array", "items": "long" } }] }`,
			test: func(t *testing.T, schema *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, "record", schema.Type[0])
				require.Equal(t, "array", schema.Fields[0].Schema[0].Type[0])
				require.Equal(t, "long", schema.Fields[0].Schema[0].Items.Type[0])
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var s *schema.Schema
			err := json.Unmarshal([]byte(tc.input), &s)
			tc.test(t, s, err)
		})
	}
}

func TestSchema(t *testing.T) {
	var s *schema.Schema
	b := jsonString
	err := json.Unmarshal([]byte(b), &s)
	require.NoError(t, err)

	require.Equal(t, "record", s.Type[0])
	require.Equal(t, "Example", s.Name)
	require.Equal(t, "A simple name (attribute) and no namespace attribute: use the null namespace (\"\"); the fullname is 'Example'.", s.Doc)
	require.Len(t, s.Fields, 3)
	require.Equal(t, "inheritNull", s.Fields[0].Name)
	require.Equal(t, "enum", s.Fields[0].Schema[0].Type[0])
	require.Equal(t, "Simple", s.Fields[0].Schema[0].Name)
	require.Equal(t, "A simple name (attribute) and no namespace attribute: inherit the null namespace of the enclosing type 'Example'. The fullname is 'Simple'.", s.Fields[0].Schema[0].Doc)
	require.Equal(t, []string{"a", "b"}, s.Fields[0].Schema[0].Symbols)

}

const jsonString = `{
  "type": "record",
  "name": "Example",
  "doc": "A simple name (attribute) and no namespace attribute: use the null namespace (\"\"); the fullname is 'Example'.",
  "fields": [
    {
      "name": "inheritNull",
      "type": {
        "type": "enum",
        "name": "Simple",
        "doc": "A simple name (attribute) and no namespace attribute: inherit the null namespace of the enclosing type 'Example'. The fullname is 'Simple'.",
        "symbols": ["a", "b"]
      }
    },
    {
      "name": "explicitNamespace",
      "type": {
        "type": "fixed",
        "name": "Simple",
        "namespace": "explicit",
        "doc": "A simple name (attribute) and a namespace (attribute); the fullname is 'explicit.Simple' (this is a different type than of the 'inheritNull' field).",
        "size": 12
      }
    }, 
    {
      "name": "fullName",
      "type": {
        "type": "record",
        "name": "a.full.Name",
        "namespace": "ignored",
        "doc": "A name attribute with a fullname, so the namespace attribute is ignored. The fullname is 'a.full.Name', and the namespace is 'a.full'.",
        "fields": [
          {
            "name": "inheritNamespace",
            "type": {
              "type": "enum",
              "name": "Understanding",
              "doc": "A simple name (attribute) and no namespace attribute: inherit the namespace of the enclosing type 'a.full.Name'. The fullname is 'a.full.Understanding'.",
              "symbols": ["d", "e"]
            }
          }
        ]
      }
    }
  ]
}`
