package schema_test

import (
	"encoding/json"
	"mokapi/config/dynamic"
	"mokapi/providers/openapi/schema"
	"mokapi/providers/openapi/schema/schematest"
	jsonSchema "mokapi/schema/json/schema"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestSchema_MarshalJSON(t *testing.T) {
	testcases := []struct {
		name   string
		schema *schema.Schema
		exp    string
	}{
		{
			name:   "null",
			schema: nil,
			exp:    `null`,
		},
		{
			name:   "property schema is null",
			schema: schematest.New("object", schematest.WithProperty("foo", nil)),
			exp:    `{"type":"object","properties":{"foo":null}}`,
		},
		{
			name:   "$ref",
			schema: &schema.Schema{Reference: dynamic.Reference[*schema.Schema]{Ref: "#/components/schemas/Foo"}},
			exp:    `{"$ref":"#/components/schemas/Foo"}`,
		},
		{
			name:   "false",
			schema: &schema.Schema{Boolean: toBoolP(false)},
			exp:    `false`,
		},
		{
			name:   "type",
			schema: &schema.Schema{Type: jsonSchema.Types{"string"}},
			exp:    `{"type":"string"}`,
		},
		{
			name: "ref",
			schema: schematest.New("object",
				schematest.WithProperty("foo", schematest.New("string")),
				schematest.WithRef("#/components/schemas/Foo"),
			),
			exp: `{"$ref":"#/components/schemas/Foo","type":"object","properties":{"foo":{"type":"string"}}}`,
		},
		{
			name:   "Sub should not be marshalled",
			schema: &schema.Schema{Sub: schematest.New("string")},
			exp:    `{}`,
		},
		{
			name:   "exclusiveMinimum",
			schema: schematest.New("integer", schematest.WithExclusiveMinimum(1)),
			exp:    `{"type":"integer","exclusiveMinimum":1}`,
		},
		{
			name:   "exclusiveMinimum",
			schema: schematest.New("integer", schematest.WithExclusiveMinimumBool(true)),
			exp:    `{"type":"integer","exclusiveMinimum":true}`,
		},
		{
			name: "integer",
			schema: schematest.New("integer",
				schematest.WithFormat("int64"),
				schematest.WithExample(10),
			),
			exp: `{"type":"integer","format":"int64","example":10}`,
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s, err := json.Marshal(tc.schema)
			require.NoError(t, err)
			require.Equal(t, tc.exp, string(s))
		})
	}
}

func TestSchema_MarshalYAML(t *testing.T) {
	testcases := []struct {
		name string
		s    *schema.Schema
		test func(t *testing.T, s string, err error)
	}{
		{
			name: "$ref",
			s:    &schema.Schema{Reference: dynamic.Reference[*schema.Schema]{Ref: "#/components/schemas/foo"}},
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, "$ref: '#/components/schemas/foo'\n", s)
			},
		},
		{
			name: "no $ref",
			s:    &schema.Schema{},
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, "{}\n", s)
			},
		},
		{
			name: "integer",
			s: schematest.New("integer",
				schematest.WithFormat("int64"),
				schematest.WithExample(10),
			),
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `type: integer
format: int64
example: 10
`, s)
			},
		},
		{
			name: "array",
			s: schematest.New("array",
				schematest.WithItems(
					"string",
					schematest.WithMinLength(5),
				),
			),
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `type: array
items:
    type: string
    minLength: 5
`, s)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			b, err := yaml.Marshal(tc.s)
			tc.test(t, string(b), err)
		})
	}
}

func TestCircularRef(t *testing.T) {
	s := &schema.Schema{}
	s.Properties = &schema.Schemas{}
	s.Properties.Set("foo", s)

	b, err := json.Marshal(s)
	require.NoError(t, err)
	require.Equal(t, "{\"properties\":{\"foo\":{\"description\":\"circular reference\"}}}", string(b))

	// with ref
	s.Ref = "#/components/schemas/Foo"
	b, err = json.Marshal(s)
	require.NoError(t, err)
	require.Equal(t, "{\"$ref\":\"#/components/schemas/Foo\",\"properties\":{\"foo\":{\"$ref\":\"#/components/schemas/Foo\",\"description\":\"circular reference\"}}}", string(b))

	// multi-level circular refs
	s = &schema.Schema{Properties: &schema.Schemas{}}

	bar := &schema.Schema{Properties: &schema.Schemas{}}
	bar.Properties.Set("foo", s)

	s.Properties.Set("bar", bar)
	b, err = json.Marshal(s)
	require.NoError(t, err)
	require.Equal(t, "{\"properties\":{\"bar\":{\"properties\":{\"foo\":{\"description\":\"circular reference\"}}}}}", string(b))
}
