package generator

import (
	"mokapi/config/static"
	"mokapi/schema/json/schema/schematest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestObject(t *testing.T) {
	testcases := []struct {
		name               string
		req                *Request
		optionalProperties string
		test               func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "object",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithProperty("name", nil),
					schematest.WithRequired("name"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"name": "Ivy"}, v)
			},
		},
		{
			name: "no type but properties",
			req: &Request{
				Schema: schematest.NewTypes(nil,
					schematest.WithProperty("name", nil),
					schematest.WithRequired("name"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"name": "Ivy"}, v)
			},
		},
		{
			name: "with one property is required",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithProperty("name", nil),
					schematest.WithProperty("foo", nil),
					schematest.WithRequired("name"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"name": "Ivy", "foo": false}, v)
			},
		},
		{
			name: "with required and maxProperties",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithProperty("name", nil),
					schematest.WithProperty("foo", nil),
					schematest.WithProperty("bar", nil),
					schematest.WithProperty("yuh", nil),
					schematest.WithProperty("zzz", nil),
					schematest.WithRequired("name"),
					schematest.WithMaxProperties(3),
				),
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]any{"name": "Ivy", "foo": false, "zzz": false}, v)
			},
		},
		{
			name: "no additional properties",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithProperty("name", schematest.New("string")),
					schematest.WithRequired("name"),
					schematest.WithFreeForm(false),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"name": "Ivy"}, v)
			},
		},
		{
			name: "with additional properties true should not create additional properties because users do not expect this",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithProperty("name", schematest.New("string")),
					schematest.WithRequired("name"),
					schematest.WithFreeForm(true),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"name": "Ivy"}, v)
			},
		},
		{
			name: "with additional properties string but maxProperties=1 should not create additional properties",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithProperty("name", schematest.New("string")),
					schematest.WithRequired("name"),
					schematest.WithAdditionalProperties(schematest.New("string")),
					schematest.WithMaxProperties(1),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"name": "Zenix"}, v)
			},
		},
		{
			name: "additionalProperties string",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithAdditionalProperties(schematest.New("string")),
					schematest.WithMinProperties(1),
					schematest.WithMaxProperties(3),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]interface{}{
						"bale": "MZsBhpy",
					},
					v)
			},
		},
		{
			name: "with additional properties specific",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithAdditionalProperties(
						schematest.New("object",
							schematest.WithProperty("age", schematest.New("integer")),
							schematest.WithProperty("gender", schematest.New("string")),
							schematest.WithRequired("age", "gender"),
						),
					),
					schematest.WithMaxProperties(3),
				),
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]any{
					"bale": map[string]any{
						"age": int64(85), "gender": "male",
					},
				}, v)
			},
		},
		{
			name: "dictionary with min and max length",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithAdditionalProperties(
						schematest.New("string"),
					),
					schematest.WithMinProperties(10),
					schematest.WithMaxProperties(12),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Len(t, v, 10)
				require.Equal(t,
					map[string]interface{}{"bale": "boezC", "bulb": "elk", "corruption": "", "equipment": "VnIkdsa aJFO", "gauva": "iyHPypTTndl", "issue": "sguOhkxhF", "jacket": "kWMaW", "man": "jbBbaS", "pack": " X", "woman": "0FILKfYaxbWMlsp"},
					v)
			},
		},
		{
			name: "object no properties",
			req: &Request{
				Schema: schematest.New("object"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Len(t, v.(map[string]any), 1)
			},
		},
		{
			name: "object no properties with examples",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithExamples(map[string]interface{}{"foo": "bar"}),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]interface{}{
						"foo": "bar",
					},
					v)
			},
		},
		{
			name: "fallback to example",
			req: &Request{
				Path: []string{"address"},
				Schema: schematest.New("object",
					schematest.WithProperty("foo", schematest.New("string")),
					schematest.WithExamples(map[string]interface{}{"foo": "bar"}),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]interface{}{
						"foo": "bar",
					},
					v)
			},
		},
		{
			name: "object no properties with required properties",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithRequired("foo", "bar"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]interface{}{
						"foo": false,
						"bar": []interface{}{975538.5032831749},
					},
					v)
			},
		},
		{
			name: "object with properties and additional required properties",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithProperty("foo", schematest.New("string")),
					schematest.WithRequired("bar"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]interface{}{
						"foo": "F2cjChNLDnmqkY",
						"bar": 313469.93141089816,
					},
					v)
			},
		},
		{
			name: "patternProperties",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithPatternProperty("^S_", schematest.New("string")),
					schematest.WithPatternProperty("^I_", schematest.New("integer")),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]interface{}{
						"I_8wE":  int64(-873413),
						"S_l7QZ": "NLD",
					},
					v)
			},
		},
		{
			name: "dependentRequired should add required property",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithProperty("name", schematest.New("string")),
					schematest.WithProperty("credit_card", schematest.New("string")),
					schematest.WithProperty("billing_address", schematest.New("string")),
					schematest.WithRequired("name", "credit_card"),
					schematest.WithDependentRequired("credit_card", "billing_address"),
				),
			},
			optionalProperties: "0",
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]interface{}{
						"name":            "Ivy",
						"credit_card":     "2824193600225549",
						"billing_address": "EEalaJFOjl",
					},
					v)
			},
		},
		{
			name: "dependentRequired should remove optional property to satisfy maxProperties",
			req: &Request{
				Schema: schematest.New("object",
					// order of properties is relevant. generator will add name and credit_card but not
					// billing_address because maxProperties=2 and then will remove credit_card
					// because dependentRequired will exceed maxProperties
					schematest.WithProperty("name", schematest.New("string")),
					schematest.WithProperty("billing_address", schematest.New("string")),
					schematest.WithProperty("credit_card", schematest.New("string")),
					schematest.WithDependentRequired("credit_card", "billing_address"),
					schematest.WithMaxProperties(2),
				),
			},
			optionalProperties: "1",
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]interface{}{
						"name":            "Ivy",
						"billing_address": "vtLnEEalaJ",
					},
					v)
			},
		},
		{
			name: "dependentRequired not possible",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithProperty("name", schematest.New("string")),
					schematest.WithProperty("credit_card", schematest.New("string")),
					schematest.WithProperty("billing_address", schematest.New("string")),
					schematest.WithRequired("name", "credit_card"),
					schematest.WithDependentRequired("credit_card", "billing_address"),
					schematest.WithMaxProperties(2),
				),
			},
			optionalProperties: "1",
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "failed to generate valid object: reached attempt limit (10) caused by: cannot apply dependentRequired for 'credit_card': maxProperties=2 was exceeded")
			},
		},
		{
			name: "dependentSchemas should add required property",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithProperty("name", schematest.New("string")),
					schematest.WithProperty("credit_card", schematest.New("string")),
					schematest.WithRequired("name", "credit_card"),
					schematest.WithDependentSchemas("credit_card",
						schematest.NewTypes(nil,
							schematest.WithProperty("billing_address", schematest.New("string")),
							schematest.WithRequired("billing_address"),
						),
					),
				),
			},
			optionalProperties: "0",
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]interface{}{
						"name":            "Ivy",
						"credit_card":     "2824193600225549",
						"billing_address": "EEalaJFOjl",
					},
					v)
			},
		},
		{
			name: "dependentRequired not possible",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithProperty("name", schematest.New("string")),
					schematest.WithProperty("credit_card", schematest.New("string")),
					schematest.WithProperty("billing_address", schematest.New("string")),
					schematest.WithRequired("name", "credit_card"),
					schematest.WithDependentSchemas("credit_card",
						schematest.NewTypes(nil,
							schematest.WithProperty("billing_address", schematest.New("string")),
							schematest.WithRequired("billing_address"),
						),
					),
					schematest.WithMaxProperties(2),
				),
			},
			optionalProperties: "1",
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "failed to generate valid object: reached attempt limit (10) caused by: cannot apply dependentSchemas for 'credit_card': maxProperties=2 was exceeded")
			},
		},
		{
			name: "if-then",
			req: &Request{
				Path: []string{"address"},
				Schema: schematest.New("object",
					schematest.WithProperty("country", schematest.New("string")),
					schematest.WithIf(schematest.NewTypes(nil,
						schematest.WithProperty("country", schematest.New("string",
							schematest.WithConst("Benin")),
						),
						schematest.WithRequired("country"),
					)),
					schematest.WithThen(schematest.NewTypes(nil,
						schematest.WithProperty("postal_code", schematest.New("string",
							schematest.WithPattern("[0-9]{5}(-[0-9]{4})?"))),
						schematest.WithRequired("postal_code"),
					)),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]interface{}{
						"country":     "Benin",
						"postal_code": "80291",
					},
					v)
			},
		},
		{
			name: "if-then but not object",
			req: &Request{
				Path: []string{"address"},
				Schema: schematest.New("object",
					schematest.WithProperty("country", schematest.New("string", schematest.WithConst("Slovenia"))),
					schematest.WithIf(schematest.NewTypes(nil,
						schematest.WithProperty("country", schematest.New("string",
							schematest.WithConst("Slovenia")),
						),
						schematest.WithRequired("country"),
					)),
					schematest.WithThen(schematest.New("string")),
					schematest.WithMinProperties(1),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "failed to generate valid object: reached attempt limit (10) caused by: cannot satisfy conditions")
			},
		},
		{
			name: "if-then but maxProperties should take different country",
			req: &Request{
				Path: []string{"address"},
				Schema: schematest.New("object",
					schematest.WithProperty("country", schematest.New("string")),
					schematest.WithIf(schematest.NewTypes(nil,
						schematest.WithProperty("country", schematest.New("string",
							schematest.WithConst("Benin")),
						),
						schematest.WithRequired("country"),
					)),
					schematest.WithThen(schematest.NewTypes(nil,
						schematest.WithProperty("postal_code", schematest.New("string",
							schematest.WithPattern("[0-9]{5}(-[0-9]{4})?"))),
						schematest.WithRequired("postal_code"),
					)),
					schematest.WithMinProperties(1),
				),
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]any{"country": "Benin", "postal_code": "80291"},
					v)
			},
		},
		{
			name: "if with maxProperties but country is required",
			req: &Request{
				Path: []string{"address"},
				Schema: schematest.New("object",
					schematest.WithProperty("country", schematest.New("string",
						// set const value to force the error
						schematest.WithConst("Slovenia"),
					)),
					schematest.WithIf(schematest.NewTypes(nil,
						schematest.WithProperty("country", schematest.New("string",
							schematest.WithConst("Slovenia")),
						),
						schematest.WithRequired("country"),
					)),
					schematest.WithThen(schematest.NewTypes(nil,
						schematest.WithProperty("postal_code", schematest.New("string",
							schematest.WithPattern("[0-9]{5}(-[0-9]{4})?"))),
						schematest.WithRequired("postal_code"),
					)),
					schematest.WithRequired("country"),
					schematest.WithMaxProperties(1),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "failed to generate valid object: reached attempt limit (10) caused by: conditional schema could not be applied: reached attempt limit (10) caused by: reached maximum of value maxProperties=1")
			},
		},
		{
			name: "if with maxProperties but country is required first try with Slovenia (random value) fails",
			req: &Request{
				Path: []string{"address"},
				Schema: schematest.New("object",
					schematest.WithProperty("country", schematest.New("string")),
					schematest.WithIf(schematest.NewTypes(nil,
						schematest.WithProperty("country", schematest.New("string",
							schematest.WithConst("Benin")),
						),
						schematest.WithRequired("country"),
					)),
					schematest.WithThen(schematest.NewTypes(nil,
						schematest.WithProperty("postal_code", schematest.New("string",
							schematest.WithPattern("[0-9]{5}(-[0-9]{4})?"))),
						schematest.WithRequired("postal_code"),
					)),
					schematest.WithRequired("country"),
					schematest.WithMaxProperties(1),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]any{"country": "Azerbaijan"}, v)
			},
		},
		{
			name: "if-else",
			req: &Request{
				Path: []string{"address"},
				Schema: schematest.New("object",
					schematest.WithProperty("country", schematest.New("string")),
					schematest.WithIf(schematest.NewTypes(nil,
						schematest.WithProperty("country", schematest.New("string",
							schematest.WithConst("Canada")),
						),
						schematest.WithRequired("country"),
					)),
					schematest.WithElse(schematest.NewTypes(nil,
						schematest.WithProperty("postal_code", schematest.New("string",
							schematest.WithPattern("[A-Z][0-9][A-Z] [0-9][A-Z][0-9]")),
						),
					)),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]interface{}{
						"country":     "Benin",
						"postal_code": "C0O 9F0",
					},
					v)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			Seed(1234567)

			if tc.optionalProperties != "" {
				SetConfig(static.DataGen{
					OptionalProperties: tc.optionalProperties,
				})
				defer SetConfig(static.DataGen{})
			}

			v, err := New(tc.req)
			tc.test(t, v, err)
		})
	}
}
