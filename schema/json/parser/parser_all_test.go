package parser_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/parser"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schematest"
	"testing"
)

func TestParser_ParseAll(t *testing.T) {
	testcases := []struct {
		name   string
		data   interface{}
		schema *schema.Schema
		test   func(t *testing.T, v interface{}, err error)
	}{
		{
			name:   "AllOf empty",
			data:   12,
			schema: schematest.NewAllOf(),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(12), v)
			},
		},
		{
			name:   "AllOf with one matching type",
			data:   12,
			schema: schematest.NewAllOf(schematest.New("integer")),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(12), v)
			},
		},
		{
			name: "AllOf with two matching type",
			data: 12,
			schema: schematest.NewAllOf(
				schematest.New("integer"),
				schematest.New("integer", schematest.WithMaximum(12)),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(12), v)
			},
		},
		{
			name: "AllOf with two types one is empty",
			data: 12,
			schema: schematest.NewAllOf(
				schematest.New("integer"),
				schematest.NewTypes(nil),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(12), v)
			},
		},
		{
			name: "AllOf with two matching type but not valid",
			data: 12,
			schema: schematest.NewAllOf(
				schematest.New("integer"),
				schematest.New("integer", schematest.WithMaximum(11)),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "found 1 error:\nparse 12 failed: does not match all schemas from 'allOf': all of schema type=integer, schema type=integer maximum=11: integer 12 exceeds maximum value of 11\nschema path #/maximum")
			},
		},
		{
			name: "AllOf with two NOT matching type",
			data: 12,
			schema: schematest.NewAllOf(
				schematest.New("integer"),
				schematest.New("string"),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "found 1 error:\nparse 12 failed: does not match all schemas from 'allOf': all of schema type=integer, schema type=string: invalid type, expected string but got integer\nschema path #/type")
			},
		},
		{
			name: "object with nil schema",
			data: map[string]interface{}{"foo": "bar"},
			schema: schematest.NewAllOf(
				nil,
				schematest.New("object", schematest.WithProperty("foo", schematest.New("string"))),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": "bar"}, v)
			},
		},
		{
			name: "object with empty schema",
			data: map[string]interface{}{"foo": "bar"},
			schema: schematest.NewAllOf(
				&schema.Schema{},
				schematest.New("object", schematest.WithProperty("foo", schematest.New("string"))),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": "bar"}, v)
			},
		},
		{
			name: "object and integer",
			data: map[string]interface{}{"foo": "bar"},
			schema: schematest.NewAllOf(
				schematest.New("integer"),
				schematest.New("object", schematest.WithProperty("foo", schematest.New("string"))),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "found 1 error:\ndoes not match all schemas from 'allOf':\ninvalid type, expected integer but got object\nschema path #/allOf/0/type")
			},
		},
		{
			name: "AllOf with two objects",
			data: map[string]interface{}{
				"name": "carol",
				"age":  28,
			},
			schema: schematest.NewAllOf(
				schematest.New("object", schematest.WithProperty("name", schematest.New("string"))),
				schematest.New("object", schematest.WithProperty("age", schematest.New("integer"))),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"age": int64(28), "name": "carol"}, v)
			},
		},
		{
			name: "AllOf with two objects, one defines a property without type",
			data: map[string]interface{}{
				"name": "carol",
				"age":  28,
			},
			schema: schematest.NewAllOf(
				schematest.New("object", schematest.WithProperty("name", schematest.New("string")), schematest.WithProperty("age", schematest.New(""))),
				schematest.New("object", schematest.WithProperty("age", schematest.New("integer"))),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"age": int64(28), "name": "carol"}, v)
			},
		},
		{
			name: "AllOf with two objects free-form false",
			data: map[string]interface{}{
				"name": "foo",
				"age":  28,
			},
			schema: schematest.NewAllOf(
				schematest.New("object", schematest.WithProperty("name", schematest.New("string")), schematest.WithFreeForm(false)),
				schematest.New("object", schematest.WithProperty("age", schematest.New("integer")), schematest.WithFreeForm(false)),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "found 2 errors:\ndoes not match all schemas from 'allOf':\nproperty 'age' not defined and the schema does not allow additional properties\nschema path #/allOf/0/additionalProperties\nproperty 'name' not defined and the schema does not allow additional properties\nschema path #/allOf/1/additionalProperties")
			},
		},
		{
			name: "AllOf with two objects free-form false and errors - error message should contain free-form=false",
			data: map[string]interface{}{
				"name": 12,
				"age":  "28",
			},
			schema: schematest.NewAllOf(
				schematest.New("object", schematest.WithProperty("name", schematest.New("string")), schematest.WithFreeForm(false)),
				schematest.New("object", schematest.WithProperty("age", schematest.New("integer")), schematest.WithFreeForm(false)),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "found 2 errors:\ndoes not match all schemas from 'allOf':\ninvalid type, expected string but got integer\nschema path #/allOf/0/name/type\nproperty 'age' not defined and the schema does not allow additional properties\nschema path #/allOf/0/additionalProperties\ninvalid type, expected integer but got string\nschema path #/allOf/1/age/type\nproperty 'name' not defined and the schema does not allow additional properties\nschema path #/allOf/1/additionalProperties")
			},
		},
		{
			name: "AllOf example from Spec: extending closed schema",
			schema: schematest.New("object",
				schematest.WithAllOf(
					schematest.New("object",
						schematest.WithProperty("street_address", schematest.New("string")),
						schematest.WithProperty("city", schematest.New("string")),
						schematest.WithProperty("state", schematest.New("string")),
						schematest.WithRequired("street_address", "city", "state"),
						schematest.WithFreeForm(false),
					),
				),
				schematest.WithProperty("type", schematest.NewTypes(nil, schematest.WithEnum([]interface{}{"residential", "business"}))),
				schematest.WithRequired("type"),
			),
			data: map[string]interface{}{
				"street_address": "1600 Pennsylvania Avenue NW",
				"city":           "Washington",
				"state":          "DC",
				"type":           "business",
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "found 1 error:\ndoes not match all schemas from 'allOf':\nproperty 'type' not defined and the schema does not allow additional properties\nschema path #/allOf/0/additionalProperties")
			},
		},
		{
			name: "AllOf example from Spec: UnevaluatedProperties error",
			schema: schematest.New("object",
				schematest.WithAllOf(
					schematest.New("object",
						schematest.WithProperty("street_address", schematest.New("string")),
						schematest.WithProperty("city", schematest.New("string")),
						schematest.WithProperty("state", schematest.New("string")),
						schematest.WithRequired("street_address", "city", "state"),
					),
				),
				schematest.WithProperty("type", schematest.NewTypes(nil, schematest.WithEnum([]interface{}{"residential", "business"}))),
				schematest.WithRequired("type"),
				schematest.WithUnevaluatedProperties(false),
			),
			data: map[string]interface{}{
				"street_address":                "1600 Pennsylvania Avenue NW",
				"city":                          "Washington",
				"state":                         "DC",
				"type":                          "business",
				"something that doesn't belong": "hi!",
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "found 1 error:\nproperty something that doesn't belong not successfully evaluated and schema does not allow unevaluated properties\nschema path #/unevaluatedProperties")
			},
		},
		{
			name: "AllOf example from Spec: UnevaluatedProperties",
			schema: schematest.New("object",
				schematest.WithAllOf(
					schematest.New("object",
						schematest.WithProperty("street_address", schematest.New("string")),
						schematest.WithProperty("city", schematest.New("string")),
						schematest.WithProperty("state", schematest.New("string")),
						schematest.WithRequired("street_address", "city", "state"),
					),
				),
				schematest.WithProperty("type", schematest.NewTypes(nil, schematest.WithEnum([]interface{}{"residential", "business"}))),
				schematest.WithRequired("type"),
				schematest.WithUnevaluatedProperties(false),
			),
			data: map[string]interface{}{
				"street_address": "1600 Pennsylvania Avenue NW",
				"city":           "Washington",
				"state":          "DC",
				"type":           "business",
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := &parser.Parser{ValidateAdditionalProperties: true}
			v, err := p.Parse(tc.data, &schema.Ref{Value: tc.schema})
			tc.test(t, v, err)
		})
	}
}
