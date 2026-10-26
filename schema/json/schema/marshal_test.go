package schema_test

import (
	"encoding/json"
	"mokapi/config/dynamic"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schema/schematest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRef_MarshalJSON(t *testing.T) {
	testcases := []struct {
		name string
		s    *schema.Schema
		test func(t *testing.T, s string, err error)
	}{
		{
			name: "empty type",
			s:    &schema.Schema{},
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, "{}", s)
			},
		},
		{
			name: "with type",
			s:    &schema.Schema{Type: schema.Types{"string"}},
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"type":"string"}`, s)
			},
		},
		{
			name: "with sub",
			s:    &schema.Schema{Type: schema.Types{"string"}, Sub: schematest.New("string")},
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"type":"string"}`, s)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			b, err := json.Marshal(tc.s)
			tc.test(t, string(b), err)
		})
	}
}

func TestSchema_MarshalJSON_Recursion(t *testing.T) {
	testcases := []struct {
		name string
		s    func() *schema.Schema
		test func(t *testing.T, s string, err error)
	}{
		{
			name: "recursion in property",
			s: func() *schema.Schema {
				s := &schema.Schema{
					Reference:  dynamic.Reference[*schema.Schema]{Ref: "foo"},
					Properties: &schema.Schemas{},
				}
				s.Properties.Set("foo", s)
				return s
			},
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"$ref":"foo","properties":{"foo":{"$ref":"foo","description":"circular reference"}}}`, s)
			},
		},
		{
			name: "recursion in array",
			s: func() *schema.Schema {
				s := &schema.Schema{
					Reference: dynamic.Reference[*schema.Schema]{Ref: "foo"},
				}
				s.Items = s
				return s
			},
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"$ref":"foo","items":{"$ref":"foo","description":"circular reference"}}`, s)
			},
		},
		{
			name: "recursion in contains",
			s: func() *schema.Schema {
				s := &schema.Schema{
					Reference: dynamic.Reference[*schema.Schema]{Ref: "foo"},
				}
				s.Contains = s
				return s
			},
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"$ref":"foo","contains":{"$ref":"foo","description":"circular reference"}}`, s)
			},
		},
		{
			name: "definitions",
			s: func() *schema.Schema {
				var s *schema.Schema
				err := json.Unmarshal([]byte(`
{
  "definitions": {
    "Status": {
      "$id": "status",
      "description": "a description"
    }
  }
}
`), &s)
				if err != nil {
					panic(err)
				}
				return s
			},
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"definitions":{"Status":{"$id":"status","description":"a description"}}}`, s)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			b, err := json.Marshal(tc.s())
			tc.test(t, string(b), err)
		})
	}
}
