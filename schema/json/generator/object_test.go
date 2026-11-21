package generator

import (
	"mokapi/config/static"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schema/schematest"
	"mokapi/sortedmap"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestObject(t *testing.T) {
	testcases := []struct {
		name               string
		req                *Request
		optionalProperties string
		test               func(t *testing.T, v any, err error)
	}{
		{
			name: "object",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithProperty("name", nil),
					schematest.WithRequired("name"),
				),
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]any{"name": "Ivy"}, v)
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
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]any{"name": "Ivy"}, v)
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
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]any{"name": "Ivy", "foo": false}, v)
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
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]any{"name": "Ivy"}, v)
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
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]any{"name": "Ivy"}, v)
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
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]any{"name": "Zenix"}, v)
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
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]any{
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
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Len(t, v, 10)
				require.Equal(t,
					map[string]any{"bale": "boezC", "bulb": "elk", "corruption": "", "equipment": "VnIkdsa aJFO", "gauva": "iyHPypTTndl", "issue": "sguOhkxhF", "jacket": "kWMaW", "man": "jbBbaS", "pack": " X", "woman": "0FILKfYaxbWMlsp"},
					v)
			},
		},
		{
			name: "object no properties",
			req: &Request{
				Schema: schematest.New("object"),
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Len(t, v.(map[string]any), 1)
			},
		},
		{
			name: "object no properties with examples",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithExamples(map[string]any{"foo": "bar"}),
				),
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]any{
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
					schematest.WithExamples(map[string]any{"foo": "bar"}),
				),
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]any{
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
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]any{
						"foo": false,
						"bar": []any{975538.5032831749},
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
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				m := v.(map[string]any)
				require.InDelta(t, 313469.93141089816, m["bar"], 0.000001)
				require.Equal(t, "F2cjChNLDnmqkY", m["foo"])
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
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]any{
						"I_gIU":  int64(-873413),
						"S_jHcT": "NLD",
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
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]any{
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
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]any{
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
			test: func(t *testing.T, v any, err error) {
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
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]any{
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
			test: func(t *testing.T, v any, err error) {
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
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]any{
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
			test: func(t *testing.T, v any, err error) {
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
			test: func(t *testing.T, v any, err error) {
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
			test: func(t *testing.T, v any, err error) {
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
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]any{
						"country":     "Benin",
						"postal_code": "C0O 9F0",
					},
					v)
			},
		},
		{
			name: "test person domain",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithProperty("name", schematest.New("string")),
					schematest.WithProperty("birthDate", schematest.New("string")),
					schematest.WithRequired("name", "birthDate"),
				),
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]any{
						"name":      "Aria Hernandez",
						"birthDate": "2013-03-15",
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

func TestObject_PropertyOrderDoesNotMatter(t *testing.T) {
	t.Run("email uses the person name", func(t *testing.T) {
		for _, names := range [][2]string{{"firstName", "lastName"}, {"first_name", "last_name"}} {
			first, last := names[0], names[1]
			orders := [][]string{
				{"email", first, last},
				{"email", last, first},
				{first, "email", last},
				{first, last, "email"},
				{last, first, "email"},
			}

			for _, order := range orders {
				t.Run(strings.Join(order, ","), func(t *testing.T) {
					Seed(1234567)

					s := schematest.New("object", schematest.WithRequired(order...))
					for _, name := range order {
						schematest.WithProperty(name, schematest.New("string"))(s)
					}

					v, err := New(&Request{Schema: s})

					require.NoError(t, err)
					obj := v.(map[string]any)
					prefix := strings.ToLower(obj[first].(string)) + "." + strings.ToLower(obj[last].(string)) + "@"
					require.True(t, strings.HasPrefix(obj["email"].(string), prefix), "%v", obj)
				})
			}
		}
	})

	t.Run("all email properties use the person name", func(t *testing.T) {
		Seed(1234567)

		v, err := New(&Request{
			Schema: schematest.New("object",
				schematest.WithProperty("email", schematest.New("string")),
				schematest.WithProperty("emailAddress", schematest.New("string")),
				schematest.WithProperty("firstName", schematest.New("string")),
				schematest.WithProperty("lastName", schematest.New("string")),
				schematest.WithRequired("email", "emailAddress", "firstName", "lastName"),
			),
		})

		require.NoError(t, err)
		obj := v.(map[string]any)
		prefix := strings.ToLower(obj["firstName"].(string)) + "." + strings.ToLower(obj["lastName"].(string)) + "@"
		require.True(t, strings.HasPrefix(obj["email"].(string), prefix), "%v", obj)
		require.True(t, strings.HasPrefix(obj["emailAddress"].(string), prefix), "%v", obj)
	})
}

func TestTopologicalSort(t *testing.T) {
	fakes := func(props ...any) *sortedmap.LinkedHashMap[string, *faker] {
		m := &sortedmap.LinkedHashMap[string, *faker]{}
		for i := 0; i < len(props); i += 2 {
			m.Set(props[i].(string), &faker{node: props[i+1].(*Node)})
		}
		return m
	}

	t.Run("dependency is resolved by node name, not property key", func(t *testing.T) {
		sorted, err := topologicalSort(fakes(
			"end_date", &Node{Name: "endDate", DependsOn: []string{"startDate"}},
			"begin", &Node{Name: "startDate"},
			"other", (*Node)(nil),
		))

		require.NoError(t, err)
		require.Equal(t, []string{"begin", "other", "end_date"}, sorted)
	})

	t.Run("every property of a node is ordered after the dependency", func(t *testing.T) {
		sorted, err := topologicalSort(fakes(
			"a", &Node{Name: "email", DependsOn: []string{"lastname"}},
			"b", &Node{Name: "email", DependsOn: []string{"lastname"}},
			"c", &Node{Name: "lastname"},
		))

		require.NoError(t, err)
		require.Equal(t, []string{"c", "a", "b"}, sorted)
	})

	t.Run("dependency on own node name is ignored", func(t *testing.T) {
		sorted, err := topologicalSort(fakes("a", &Node{Name: "x", DependsOn: []string{"x"}}))

		require.NoError(t, err)
		require.Equal(t, []string{"a"}, sorted)
	})

	t.Run("circular dependency", func(t *testing.T) {
		_, err := topologicalSort(fakes(
			"a", &Node{Name: "x", DependsOn: []string{"y"}},
			"b", &Node{Name: "y", DependsOn: []string{"x"}},
		))

		require.EqualError(t, err, "circular dependency detected")
	})
}

func TestObject_NestedObjectsAndContext(t *testing.T) {
	str := func() *schema.Schema { return schematest.New("string") }
	emailMatchesName := func(t *testing.T, obj map[string]any) {
		t.Helper()
		prefix := strings.ToLower(obj["firstName"].(string)) + "." + strings.ToLower(obj["lastName"].(string)) + "@"
		require.True(t, strings.HasPrefix(obj["email"].(string), prefix), "%v", obj)
	}

	t.Run("email after a nested object uses the name", func(t *testing.T) {
		Seed(1234567)

		v, err := New(&Request{Schema: schematest.New("object",
			schematest.WithProperty("firstName", str()),
			schematest.WithProperty("lastName", str()),
			schematest.WithProperty("address", schematest.New("object",
				schematest.WithProperty("city", str()),
				schematest.WithRequired("city"))),
			schematest.WithProperty("email", str()),
			schematest.WithRequired("firstName", "lastName", "address", "email"),
		)})

		require.NoError(t, err)
		emailMatchesName(t, v.(map[string]any))
	})

	t.Run("email of a nested object uses the name of the parent", func(t *testing.T) {
		Seed(1234567)

		v, err := New(&Request{Schema: schematest.New("object",
			schematest.WithProperty("firstName", str()),
			schematest.WithProperty("lastName", str()),
			schematest.WithProperty("contact", schematest.New("object",
				schematest.WithProperty("email", str()),
				schematest.WithRequired("email"))),
			schematest.WithRequired("firstName", "lastName", "contact"),
		)})

		require.NoError(t, err)
		obj := v.(map[string]any)
		obj["email"] = obj["contact"].(map[string]any)["email"]
		emailMatchesName(t, obj)
	})

	t.Run("items of an array get different names", func(t *testing.T) {
		Seed(1234567)

		v, err := New(&Request{Schema: schematest.New("array",
			schematest.WithItems("object",
				schematest.WithProperty("firstName", str()),
				schematest.WithProperty("lastName", str()),
				schematest.WithProperty("email", str()),
				schematest.WithRequired("firstName", "lastName", "email")),
			schematest.WithMinItems(5), schematest.WithMaxItems(5),
		)})

		require.NoError(t, err)
		names := map[string]bool{}
		for _, item := range v.([]any) {
			obj := item.(map[string]any)
			emailMatchesName(t, obj)
			names[obj["firstName"].(string)+" "+obj["lastName"].(string)] = true
		}
		require.Greater(t, len(names), 1)
	})

	t.Run("values of a nested object do not leak into the next one", func(t *testing.T) {
		Seed(1234567)
		address := func() *schema.Schema {
			return schematest.New("object",
				schematest.WithProperty("street", str()),
				schematest.WithProperty("city", str()),
				schematest.WithRequired("street", "city"))
		}

		v, err := New(&Request{Schema: schematest.New("object",
			schematest.WithProperty("billingAddress", address()),
			schematest.WithProperty("shippingAddress", address()),
			schematest.WithRequired("billingAddress", "shippingAddress"),
		)})

		require.NoError(t, err)
		obj := v.(map[string]any)
		require.NotEqual(t, obj["billingAddress"], obj["shippingAddress"])
	})

	t.Run("nested object with its own name gets a new name", func(t *testing.T) {
		Seed(99)
		person := func() *schema.Schema {
			return schematest.New("object",
				schematest.WithProperty("firstName", str()),
				schematest.WithProperty("lastName", str()),
				schematest.WithRequired("firstName", "lastName"))
		}

		v, err := New(&Request{Schema: schematest.New("object",
			schematest.WithProperty("firstName", str()),
			schematest.WithProperty("lastName", str()),
			schematest.WithProperty("manager", person()),
			schematest.WithRequired("firstName", "lastName", "manager"),
		)})

		require.NoError(t, err)
		obj := v.(map[string]any)
		manager := obj["manager"].(map[string]any)
		require.NotEqual(t, obj["firstName"], manager["firstName"])
		require.NotEqual(t, obj["lastName"], manager["lastName"])
	})

	t.Run("array of persons inside a person does not copy the name", func(t *testing.T) {
		Seed(1234567)

		v, err := New(&Request{Schema: schematest.New("object",
			schematest.WithProperty("firstName", str()),
			schematest.WithProperty("lastName", str()),
			schematest.WithProperty("friends", schematest.New("array",
				schematest.WithItems("object",
					schematest.WithProperty("firstName", str()),
					schematest.WithProperty("lastName", str()),
					schematest.WithProperty("email", str()),
					schematest.WithRequired("firstName", "lastName", "email")),
				schematest.WithMinItems(3), schematest.WithMaxItems(3))),
			schematest.WithRequired("firstName", "lastName", "friends"),
		)})

		require.NoError(t, err)
		obj := v.(map[string]any)
		names := map[string]bool{obj["firstName"].(string) + " " + obj["lastName"].(string): true}
		for _, f := range obj["friends"].([]any) {
			friend := f.(map[string]any)
			emailMatchesName(t, friend)
			names[friend["firstName"].(string)+" "+friend["lastName"].(string)] = true
		}
		require.Len(t, names, 4)
	})
}

func TestObject_RequestParametersAreEchoedAtTheShallowestLevel(t *testing.T) {
	const pathId = int64(7000001) // outside the range of generated ids
	str := func() *schema.Schema { return schematest.New("string") }
	integer := func() *schema.Schema { return schematest.New("integer") }
	idObject := func() *schema.Schema {
		return schematest.New("object", schematest.WithProperty("id", integer()), schematest.WithRequired("id"))
	}
	list := func(items *schema.Schema) *schema.Schema {
		return schematest.New("array", schematest.WithItemsNew(items), schematest.WithMinItems(3), schematest.WithMaxItems(3))
	}
	generate := func(t *testing.T, s *schema.Schema, params map[string]any) any {
		t.Helper()
		v, err := New(NewRequest(nil, s, params))
		require.NoError(t, err)
		return v
	}
	idParam := map[string]any{"id": pathId}
	requireIdsNotEqual := func(t *testing.T, items any) {
		t.Helper()
		require.Len(t, items, 3)
		for _, item := range items.([]any) {
			require.NotEqual(t, pathId, item.(map[string]any)["id"], "%v", items)
		}
	}

	t.Run("root object", func(t *testing.T) {
		v := generate(t, schematest.New("object",
			schematest.WithProperty("id", integer()),
			schematest.WithProperty("friends", list(idObject())),
			schematest.WithRequired("id", "friends"),
		), idParam).(map[string]any)

		require.Equal(t, pathId, v["id"])
		requireIdsNotEqual(t, v["friends"])
	})

	t.Run("wrapper", func(t *testing.T) {
		v := generate(t, schematest.New("object",
			schematest.WithProperty("data", schematest.New("object",
				schematest.WithProperty("id", integer()),
				schematest.WithProperty("friends", list(idObject())),
				schematest.WithRequired("id", "friends"))),
			schematest.WithRequired("data"),
		), idParam).(map[string]any)

		data := v["data"].(map[string]any)
		require.Equal(t, pathId, data["id"])
		requireIdsNotEqual(t, data["friends"])
	})

	t.Run("filter in a root array", func(t *testing.T) {
		item := schematest.New("object", schematest.WithProperty("status", str()), schematest.WithRequired("status"))

		v := generate(t, list(item), map[string]any{"status": "available"})

		for _, e := range v.([]any) {
			require.Equal(t, "available", e.(map[string]any)["status"])
		}
	})

	t.Run("filter in a wrapper", func(t *testing.T) {
		item := schematest.New("object", schematest.WithProperty("status", str()), schematest.WithRequired("status"))

		v := generate(t, schematest.New("object",
			schematest.WithProperty("items", list(item)),
			schematest.WithProperty("total", integer()),
			schematest.WithRequired("items", "total"),
		), map[string]any{"status": "active"}).(map[string]any)

		for _, e := range v["items"].([]any) {
			require.Equal(t, "active", e.(map[string]any)["status"])
		}
	})

	t.Run("siblings", func(t *testing.T) {
		v := generate(t, schematest.New("object",
			schematest.WithProperty("billing", idObject()),
			schematest.WithProperty("shipping", idObject()),
			schematest.WithRequired("billing", "shipping"),
		), idParam).(map[string]any)

		require.Equal(t, pathId, v["billing"].(map[string]any)["id"])
		require.Equal(t, pathId, v["shipping"].(map[string]any)["id"])
	})

	t.Run("uneven depth", func(t *testing.T) {
		v := generate(t, schematest.New("object",
			schematest.WithProperty("a", idObject()),
			schematest.WithProperty("b", schematest.New("object",
				schematest.WithProperty("c", idObject()),
				schematest.WithRequired("c"))),
			schematest.WithRequired("a", "b"),
		), idParam).(map[string]any)

		require.Equal(t, pathId, v["a"].(map[string]any)["id"])
		require.NotEqual(t, pathId, v["b"].(map[string]any)["c"].(map[string]any)["id"])
	})

	t.Run("type mismatch at the shallowest level is not moved deeper", func(t *testing.T) {
		nested := schematest.New("object", schematest.WithProperty("id", str()), schematest.WithRequired("id"))

		v := generate(t, schematest.New("object",
			schematest.WithProperty("id", integer()),
			schematest.WithProperty("friends", list(nested)),
			schematest.WithRequired("id", "friends"),
		), map[string]any{"id": "abc"}).(map[string]any)

		require.IsType(t, int64(0), v["id"])
		for _, f := range v["friends"].([]any) {
			require.NotEqual(t, "abc", f.(map[string]any)["id"])
		}
	})

	t.Run("array parameter picks an element at the shallowest level", func(t *testing.T) {
		ids := []any{pathId, pathId + 1}

		v := generate(t, schematest.New("object",
			schematest.WithProperty("id", integer()),
			schematest.WithProperty("friends", list(idObject())),
			schematest.WithRequired("id", "friends"),
		), map[string]any{"id": ids}).(map[string]any)

		require.Contains(t, ids, v["id"])
		for _, f := range v["friends"].([]any) {
			require.NotContains(t, ids, f.(map[string]any)["id"])
		}
	})

	t.Run("oneOf variant belongs to the level of the object", func(t *testing.T) {
		v := generate(t, schematest.New("object",
			schematest.WithProperty("friends", list(idObject())),
			schematest.WithRequired("friends"),
			schematest.WithOneOf(idObject()),
		), idParam).(map[string]any)

		require.Equal(t, pathId, v["id"])
		requireIdsNotEqual(t, v["friends"])
	})

	t.Run("allOf variant belongs to the level of the object", func(t *testing.T) {
		v := generate(t, schematest.New("object",
			schematest.WithProperty("friends", list(idObject())),
			schematest.WithRequired("friends"),
			schematest.WithAllOf(idObject()),
		), idParam).(map[string]any)

		require.Equal(t, pathId, v["id"])
		requireIdsNotEqual(t, v["friends"])
	})

	t.Run("recursive reference", func(t *testing.T) {
		node := schematest.New("object", schematest.WithProperty("id", integer()), schematest.WithRequired("id"))
		node.Ref = "#/components/schemas/Node"
		node.Properties.Set("children", &schema.Schema{Type: schema.Types{"array"}, Items: node})

		v := generate(t, node, idParam).(map[string]any)

		require.Equal(t, pathId, v["id"])
	})

	t.Run("values written by nodes are still reused when parameters exist", func(t *testing.T) {
		Seed(1234567)

		v := generate(t, schematest.New("object",
			schematest.WithProperty("firstName", str()),
			schematest.WithProperty("lastName", str()),
			schematest.WithProperty("email", str()),
			schematest.WithRequired("firstName", "lastName", "email"),
		), idParam).(map[string]any)

		prefix := strings.ToLower(v["firstName"].(string)) + "." + strings.ToLower(v["lastName"].(string)) + "@"
		require.True(t, strings.HasPrefix(v["email"].(string), prefix), "%v", v)
	})
}

func TestObject_ParametersAndNodes(t *testing.T) {
	str := func() *schema.Schema { return schematest.New("string") }
	// every nested object declares the name again, like friends[] in GET /users/{id}
	withNested := func(name string) *schema.Schema {
		nested := schematest.New("object", schematest.WithProperty(name, str()), schematest.WithRequired(name))
		return schematest.New("object",
			schematest.WithProperty(name, str()),
			schematest.WithProperty("friends", schematest.New("array",
				schematest.WithItemsNew(nested), schematest.WithMinItems(3), schematest.WithMaxItems(3))),
			schematest.WithRequired(name, "friends"),
		)
	}

	for _, name := range []string{"id", "name", "firstname", "lastname", "gender", "title", "pet", "category"} {
		t.Run("parameter "+name+" is only echoed at the shallowest level", func(t *testing.T) {
			const value = "from-request"

			v, err := New(NewRequest(nil, withNested(name), map[string]any{name: value}))

			require.NoError(t, err)
			obj := v.(map[string]any)
			if name != "id" { // a string id that is not an uuid
				require.Equal(t, value, obj[name])
			}
			for _, f := range obj["friends"].([]any) {
				require.NotEqual(t, value, f.(map[string]any)[name])
			}
		})
	}

	t.Run("parameter stays visible for nodes in nested objects", func(t *testing.T) {
		root := FindByName(RootName)
		root.Children = append(root.Children, &Node{
			Name:       "ownerId",
			Attributes: []string{"ownerId"},
			Fake: func(r *Request) (any, error) {
				return r.Context.Values["id"], nil
			},
		})
		defer func() { _ = root.Remove("ownerId") }()
		friend := schematest.New("object", schematest.WithProperty("id", str()), schematest.WithProperty("ownerId", str()), schematest.WithRequired("id", "ownerId"))
		s := schematest.New("object",
			schematest.WithProperty("id", str()),
			schematest.WithProperty("friends", schematest.New("array", schematest.WithItemsNew(friend), schematest.WithMinItems(2), schematest.WithMaxItems(2))),
			schematest.WithRequired("id", "friends"),
		)

		v, err := New(NewRequest(nil, s, map[string]any{"id": "42"}))

		require.NoError(t, err)
		for _, f := range v.(map[string]any)["friends"].([]any) {
			friend := f.(map[string]any)
			require.Equal(t, "42", friend["ownerId"])
			require.NotEqual(t, "42", friend["id"])
		}
	})
}

func TestObject_PatternPropertyIsGeneratedByItsName(t *testing.T) {
	Seed(1234567)
	s := schematest.New("object", schematest.WithPatternProperty("^email$", schematest.New("string")))

	// the number of pattern properties is random and can be 0
	generated := 0
	for range 20 {
		v, err := New(NewRequest(nil, s, nil))
		require.NoError(t, err)
		for _, email := range v.(map[string]any) {
			require.Contains(t, email, "@")
			generated++
		}
	}
	require.Greater(t, generated, 0)
}

func TestObject_UnanchoredPatternPropertyIsPaddedWithAlphanumerics(t *testing.T) {
	Seed(1234567)
	s := schematest.New("object", schematest.WithPatternProperty("S_", schematest.New("string")))
	name := regexp.MustCompile(`^[a-zA-Z0-9]*S_[a-zA-Z0-9]*$`)

	// the number of pattern properties is random and can be 0
	generated := 0
	for range 50 {
		v, err := New(NewRequest(nil, s, nil))
		require.NoError(t, err)
		for propName := range v.(map[string]any) {
			require.Regexp(t, name, propName)
			generated++
		}
	}
	require.Greater(t, generated, 0)
}
