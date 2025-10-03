package generator

import (
	"mokapi/schema/json/schema/schematest"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
)

func TestAllOf(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "with one integer schema",
			req: &Request{
				Schema: schematest.NewAllOf(
					schematest.New("integer"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(122317), v)
			},
		},
		{
			name: "with two integer schema, second is more precise",
			req: &Request{
				Schema: schematest.NewAllOf(
					schematest.New("integer"),
					schematest.New("integer", schematest.WithMinimum(0), schematest.WithMaximum(10)),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(6), v)
			},
		},
		{
			name: "string and maxLength",
			req: &Request{
				Schema: schematest.NewAllOf(
					schematest.New("string"),
					schematest.NewTypes(nil, schematest.WithMaxLength(5)),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Fq", v)
			},
		},
		{
			name: "with two integer schema, first is more precise",
			req: &Request{
				Schema: schematest.NewAllOf(
					schematest.New("integer", schematest.WithMinimum(0), schematest.WithMaximum(10)),
					schematest.New("integer"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(6), v)
			},
		},
		{
			name: "first is any, second is integer",
			req: &Request{
				Schema: schematest.NewAllOf(
					schematest.NewAny(),
					schematest.New("integer"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(122317), v)
			},
		},
		{
			name: "multiple shared types",
			req: &Request{
				Schema: schematest.NewAllOf(
					schematest.NewTypes([]string{"integer", "string", "boolean"}),
					schematest.NewTypes([]string{"integer", "string", "boolean"}),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, true, v)
			},
		},
		{
			name: "one object",
			req: &Request{
				Schema: schematest.NewAllOf(
					schematest.New("object", schematest.WithProperty("foo", schematest.New("string"))),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": "FqwCrwMfkOjojx"}, v)
			},
		},
		{
			name: "two object",
			req: &Request{
				Schema: schematest.NewAllOf(
					schematest.New("object",
						schematest.WithProperty("foo", schematest.New("string")),
						schematest.WithRequired("foo"),
					),
					schematest.New("object",
						schematest.WithProperty("bar", schematest.New("string")),
						schematest.WithRequired("bar"),
					),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"bar": "Sza", "foo": "FqwCrwMfkOjojx"}, v)
			},
		},
		{
			name: "two object with required properties",
			req: &Request{
				Schema: schematest.NewAllOf(
					schematest.New("object",
						schematest.WithProperty("foo", schematest.New("string")),
						schematest.WithRequired("foo"),
					),
					schematest.New("object",
						schematest.WithProperty("bar", schematest.New("string")),
						schematest.WithRequired("bar"),
					),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"bar": "Sza", "foo": "FqwCrwMfkOjojx"}, v)
			},
		},
		{
			name: "two object, first is integer and object",
			req: &Request{
				Schema: schematest.NewAllOf(
					schematest.NewTypes([]string{"integer", "object"},
						schematest.WithProperty("foo", schematest.New("string")),
						schematest.WithRequired("foo"),
					),
					schematest.New("object",
						schematest.WithProperty("bar", schematest.New("string")),
						schematest.WithRequired("bar"),
					),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"bar": "Sza", "foo": "FqwCrwMfkOjojx"}, v)
			},
		},
		{
			name: "no shared type",
			req: &Request{
				Schema: schematest.NewAllOf(
					schematest.NewTypes([]string{"integer", "string"}),
					schematest.NewTypes([]string{"number", "boolean"}),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "generate random data for schema failed: no shared types found: [integer, string] and [number, boolean]")
			},
		},
		{
			name: "two object with example",
			req: &Request{
				Schema: schematest.NewTypes(nil,
					schematest.WithAllOf(
						schematest.New("object", schematest.WithProperty("foo", schematest.New("string"))),
						schematest.New("object", schematest.WithProperty("bar", schematest.New("string"))),
					),
					schematest.WithExamples(
						map[string]any{"foo": "bar"},
						map[string]any{"bar": "foo"},
						map[string]any{"foo": "yuh", "bar": "foo"},
					),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"bar": "foo", "foo": "yuh"}, v)
			},
		},
		{
			name: "example from json-schema.org extending closed schemas",
			req: &Request{
				Schema: schematest.NewTypes(nil,
					schematest.WithAllOf(
						schematest.New("object",
							schematest.WithProperty("street_address", schematest.New("string")),
							schematest.WithProperty("city", schematest.New("string")),
							schematest.WithProperty("state", schematest.New("string")),
							schematest.WithRequired("street_address", "city", "state"),
							schematest.WithFreeForm(false),
						),
					),
					schematest.WithProperty("type", schematest.NewTypes(nil, schematest.WithEnumValues("residential", "business"))),
					schematest.WithRequired("type"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "failed to generate valid object: reached attempt limit (10) caused by: error count 2:\n\t- #/allOf: does not match all schema\n\t\t- #/allOf/0/additionalProperties: property 'type' not defined and the schema does not allow additional properties")
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(1234567)

			v, err := New(tc.req)
			tc.test(t, v, err)
		})
	}
}
