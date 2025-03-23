package schema_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/providers/openapi/schema"
	jsonSchema "mokapi/schema/json/schema"
	"testing"
)

func TestSchema_Marshal(t *testing.T) {
	testcases := []struct {
		name   string
		schema schema.Schema
		exp    string
	}{
		{
			name:   "$ref",
			schema: schema.Schema{Ref: "#/components/schemas/Foo"},
			exp:    `{"$ref":"#/components/schemas/Foo"}`,
		},
		{
			name:   "false",
			schema: schema.Schema{SubSchema: &schema.SubSchema{Boolean: toBoolP(false)}},
			exp:    `false`,
		},
		{
			name:   "type",
			schema: schema.Schema{SubSchema: &schema.SubSchema{Type: jsonSchema.Types{"string"}}},
			exp:    `{"type":"string"}`,
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s, err := tc.schema.MarshalJSON()
			require.NoError(t, err)
			require.Equal(t, tc.exp, string(s))
		})
	}
}
