package schema_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/providers/openapi/schema"
	"mokapi/providers/openapi/schema/schematest"
	jsonSchema "mokapi/schema/json/schema"
	"testing"
)

func TestConvert(t *testing.T) {
	testcases := []struct {
		name string
		s    *schema.Schema
		test func(t *testing.T, s *jsonSchema.Schema)
	}{
		{
			name: "nullable",
			s:    schematest.New("string", schematest.IsNullable(true)),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.Equal(t, jsonSchema.Types{"string", "null"}, s.Type)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := schema.ConvertToJsonSchema(&schema.Ref{Value: tc.s})
			tc.test(t, r.Value)
		})
	}
}
