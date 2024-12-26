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
			name: "unevaluatedProperties",
			data: map[string]interface{}{
				"name": "carol",
				"age":  28,
			},
			schema: schematest.NewAllOf(
				schematest.New("object",
					schematest.WithProperty("name", schematest.New("string")),
					schematest.WithUnevaluatedProperties(&schema.Ref{Boolean: toBoolP(false)}),
				),
				schematest.New("object", schematest.WithProperty("age", schematest.New("integer"))),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "found 1 error:\ndoes not match all schemas from 'allOf':\nproperty age not successfully evaluated and schema does not allow unevaluated properties\nschema path #/allOf/0/unevaluatedProperties")
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
				schematest.WithUnevaluatedProperties(&schema.Ref{Boolean: toBoolP(false)}),
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
				schematest.WithUnevaluatedProperties(&schema.Ref{Boolean: toBoolP(false)}),
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
		{
			name: "UnevaluatedProperties must be string",
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
				schematest.WithUnevaluatedProperties(schematest.NewRef("string")),
			),
			data: map[string]interface{}{
				"street_address":                "1600 Pennsylvania Avenue NW",
				"city":                          "Washington",
				"state":                         "DC",
				"type":                          "business",
				"something that doesn't belong": "hi!",
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"city": "Washington", "something that doesn't belong": "hi!", "state": "DC", "street_address": "1600 Pennsylvania Avenue NW", "type": "business"}, v)
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

func TestParser_AllOf_If_Then(t *testing.T) {
	// examples from https://json-schema.org/understanding-json-schema/reference/conditionals
	s := schematest.New("object",
		schematest.WithProperty("street_address", schematest.New("string")),
		schematest.WithProperty("country", schematest.NewTypes(nil,
			schematest.WithDefault("United States of America"),
			schematest.WithEnum([]interface{}{"United States of America", "Canada", "Netherlands"}),
		)),
		schematest.WithAllOf(
			schematest.NewTypes(nil,
				schematest.WithIf(schematest.NewRefTypes(nil,
					schematest.WithProperty("country", schematest.NewTypes(nil, schematest.WithConst("United States of America"))),
				)),
				schematest.WithThen(schematest.NewRefTypes(nil,
					schematest.WithProperty("postal_code", schematest.NewTypes(nil, schematest.WithPattern("[0-9]{5}(-[0-9]{4})?"))),
				)),
			),
			schematest.NewTypes(nil,
				schematest.WithIf(schematest.NewRefTypes(nil,
					schematest.WithProperty("country", schematest.NewTypes(nil, schematest.WithConst("Canada"))),
				)),
				schematest.WithThen(schematest.NewRefTypes(nil,
					schematest.WithProperty("postal_code", schematest.NewTypes(nil, schematest.WithPattern("[A-Z][0-9][A-Z] [0-9][A-Z][0-9]"))),
				)),
			),
			schematest.NewTypes(nil,
				schematest.WithIf(schematest.NewRefTypes(nil,
					schematest.WithProperty("country", schematest.NewTypes(nil, schematest.WithConst("Netherlands"))),
				)),
				schematest.WithThen(schematest.NewRefTypes(nil,
					schematest.WithProperty("postal_code", schematest.NewTypes(nil, schematest.WithPattern("[0-9]{4} [A-Z]{2}"))),
				)),
			),
		),
	)

	testcases := []struct {
		name string
		d    interface{}
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "USA",
			d: map[string]interface{}{
				"street_address": "1600 Pennsylvania Avenue NW",
				"country":        "United States of America",
				"postal_code":    "20500",
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "no country",
			d: map[string]interface{}{
				"street_address": "1600 Pennsylvania Avenue NW",
				"postal_code":    "20500",
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "Canada",
			d: map[string]interface{}{
				"street_address": "24 Sussex Drive",
				"country":        "Canada",
				"postal_code":    "K1M 1M4",
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "Canada error",
			d: map[string]interface{}{
				"street_address": "24 Sussex Drive",
				"country":        "Canada",
				"postal_code":    "10000",
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "found 1 error:\ndoes not match all schemas from 'allOf':\ndoes not match schema from 'then':\nstring '10000' does not match regex pattern '[A-Z][0-9][A-Z] [0-9][A-Z][0-9]'\nschema path #/allOf/1/then/postal_code/pattern")
			},
		},
		{
			name: "Netherlands",
			d: map[string]interface{}{
				"street_address": "Adriaan Goekooplaan",
				"country":        "Netherlands",
				"postal_code":    "2517 JX",
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "default error",
			d: map[string]interface{}{
				"street_address": "1600 Pennsylvania Avenue NW",
				"postal_code":    "K1M 1M4",
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "found 1 error:\ndoes not match all schemas from 'allOf':\ndoes not match schema from 'then':\nstring 'K1M 1M4' does not match regex pattern '[0-9]{5}(-[0-9]{4})?'\nschema path #/allOf/0/then/postal_code/pattern")
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := &parser.Parser{ValidateAdditionalProperties: true}
			v, err := p.Parse(tc.d, &schema.Ref{Value: s})
			tc.test(t, v, err)
		})
	}
}
