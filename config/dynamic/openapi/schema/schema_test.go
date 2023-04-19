package schema_test

import (
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic/openapi/schema"
	"mokapi/config/dynamic/openapi/schema/schematest"
	"testing"
)

func TestSchema_UnmarshalYAML(t *testing.T) {
	testcases := []struct {
		name string
		s    string
		fn   func(t *testing.T, schema *schema.Schema)
	}{
		{
			"empty",
			"",
			func(t *testing.T, schema *schema.Schema) {
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
			func(t *testing.T, schema *schema.Schema) {
				require.Equal(t, "object", schema.Type)
				require.False(t, schema.IsFreeForm(), "object should not be free form")
				require.False(t, schema.IsDictionary())
			},
		},
		{
			"additional properties true",
			`
type: object
additionalProperties: true
properties:
  name:
    type: string
`,
			func(t *testing.T, schema *schema.Schema) {
				require.Equal(t, "object", schema.Type)
				require.True(t, schema.IsFreeForm(), "object should be free form")
				require.False(t, schema.IsDictionary())
			},
		},
		{
			"additional properties",
			`
type: object
additionalProperties: {}
`,
			func(t *testing.T, schema *schema.Schema) {
				require.Equal(t, "object", schema.Type)
				require.True(t, schema.IsFreeForm(), "object should be free form")
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
			func(t *testing.T, schema *schema.Schema) {
				require.Equal(t, "object", schema.Type)
				require.False(t, schema.IsFreeForm())
				require.Equal(t, "string", schema.AdditionalProperties.Value.Type)
			},
		},
		{
			"allOf",
			`
allOf:
  - type: object
`,
			func(t *testing.T, schema *schema.Schema) {
				require.Equal(t, "", schema.Type)
				require.Len(t, schema.AllOf, 1)
				require.Equal(t, "object", schema.AllOf[0].Value.Type)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := &schema.Schema{}
			err := yaml.Unmarshal([]byte(tc.s), &s)
			require.NoError(t, err)
			tc.fn(t, s)
		})
	}
}

func TestSchema_IsFreeForm(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			"number",
			func(t *testing.T) {
				s := schematest.New("number")
				require.False(t, s.IsFreeForm())
			},
		},
		{
			"object with property",
			func(t *testing.T) {
				s := schematest.New("object", schematest.WithProperty("foo", schematest.New("string")))
				require.False(t, s.IsFreeForm())
			},
		},
		{
			"object without property",
			func(t *testing.T) {
				s := schematest.New("object")
				require.True(t, s.IsFreeForm())
			},
		},
		{
			"object with empty additional properties",
			func(t *testing.T) {
				s := schematest.New("object")
				s.AdditionalProperties = &schema.AdditionalProperties{}
				require.True(t, s.IsFreeForm())
			},
		},
		{
			"object with property additional false",
			func(t *testing.T) {
				s := schematest.New("object", schematest.WithProperty("foo", schematest.New("string")), schematest.WithFreeForm(false))
				require.False(t, s.IsFreeForm())
			},
		},
	}
	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.f(t)
		})
	}
}
