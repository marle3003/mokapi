package schema_test

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"mokapi/providers/openapi/schema"
	"mokapi/providers/openapi/schema/schematest"
	jsonSchema "mokapi/schema/json/schema"
	"testing"
)

func TestSchema_Marshal(t *testing.T) {
	testcases := []struct {
		name   string
		schema *schema.Schema
		exp    string
	}{
		{
			name:   "$ref",
			schema: &schema.Schema{Ref: "#/components/schemas/Foo"},
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
