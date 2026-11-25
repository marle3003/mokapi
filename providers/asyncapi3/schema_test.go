package asyncapi3_test

import (
	"encoding/json"
	"mokapi/config/dynamic"
	"mokapi/media"
	"mokapi/providers/asyncapi3"
	avro "mokapi/schema/avro/schema"
	"mokapi/schema/json/schema"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestSchema_Marshal(t *testing.T) {
	testcases := []struct {
		name string
		s    *asyncapi3.SchemaRef
		json string
		yaml string
		test func(t *testing.T, json, yaml string, err error)
	}{
		{
			name: "ref",
			s: &asyncapi3.SchemaRef{
				Reference: dynamic.Reference[*asyncapi3.SchemaRef]{
					Ref: "/foo",
				},
			},
			test: func(t *testing.T, json, yaml string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"$ref":"/foo"}`, json)
				require.Equal(t, "$ref: /foo\n", yaml)
			},
		},
		{
			name: "value",
			s: &asyncapi3.SchemaRef{
				Value: &schema.Schema{Description: "foo"},
			},
			test: func(t *testing.T, json, yaml string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"description":"foo"}`, json)
				require.Equal(t, "description: foo\n", yaml)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			jb, err := json.Marshal(tc.s)
			if err != nil {
				tc.test(t, "", "", err)
			}
			yb, err := yaml.Marshal(tc.s)
			tc.test(t, string(jb), string(yb), err)
		})
	}
}

func TestSchemaRef_GetSchema_Avro(t *testing.T) {
	avroSchema := &avro.Schema{Type: []interface{}{"record"}, Name: "Order"}

	testcases := []struct {
		name     string
		ref      *asyncapi3.SchemaRef
		expected asyncapi3.Schema
	}{
		{
			name:     "avro ref",
			ref:      &asyncapi3.SchemaRef{Value: &asyncapi3.AvroRef{Schema: avroSchema}},
			expected: avroSchema,
		},
		{
			name: "avro ref in multi schema format",
			ref: &asyncapi3.SchemaRef{Value: &asyncapi3.MultiSchemaFormat{
				Schema: &asyncapi3.SchemaRef{Value: &asyncapi3.AvroRef{Schema: avroSchema}},
			}},
			expected: avroSchema,
		},
		{
			name: "multi schema format without schema",
			ref:  &asyncapi3.SchemaRef{Value: &asyncapi3.MultiSchemaFormat{}},
		},
		{
			name: "avro ref without schema",
			ref:  &asyncapi3.SchemaRef{Value: &asyncapi3.AvroRef{}},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := tc.ref.GetSchema()
			require.NoError(t, err)
			require.Equal(t, tc.expected, s)
		})
	}
}

func TestSchemaRef_Marshal_AvroRef(t *testing.T) {
	ref := &asyncapi3.SchemaRef{Value: &asyncapi3.AvroRef{Schema: &avro.Schema{Type: []interface{}{"string"}}}}

	b, err := ref.Marshal("foo", media.ParseContentType("avro/binary"))
	require.NoError(t, err)
	require.Equal(t, []byte{6, 'f', 'o', 'o'}, b)
}
