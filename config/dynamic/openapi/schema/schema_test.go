package schema

import (
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"testing"
)

func TestSchema_UnmarshalYAML(t *testing.T) {
	testcases := []struct {
		name string
		s    string
		fn   func(t *testing.T, schema *Schema)
	}{
		{
			"empty",
			"",
			func(t *testing.T, schema *Schema) {
				require.Equal(t, "", schema.Type)
			},
		},
		{
			"additional properties false",
			`
type: object
additionalProperties: false
properties:
  name:
    type: string
`,
			func(t *testing.T, schema *Schema) {
				require.Equal(t, "object", schema.Type)
				require.False(t, schema.IsFreeForm())
			},
		},
		{
			"additional properties",
			`
type: object
additionalProperties: {}
`,
			func(t *testing.T, schema *Schema) {
				require.Equal(t, "object", schema.Type)
				require.True(t, schema.IsFreeForm())
			},
		},
		{
			"additional properties",
			`
type: object
additionalProperties:
  type: string
properties:
  name:
    type: string
`,
			func(t *testing.T, schema *Schema) {
				require.Equal(t, "object", schema.Type)
				require.False(t, schema.IsFreeForm())
				require.Equal(t, "string", schema.AdditionalProperties.Value.Type)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := &Schema{}
			err := yaml.Unmarshal([]byte(tc.s), &s)
			require.NoError(t, err)
			tc.fn(t, s)
		})
	}
}
