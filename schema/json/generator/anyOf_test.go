package generator

import (
	"mokapi/config/static"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schema/schematest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAnyOf(t *testing.T) {
	testcases := []struct {
		name               string
		req                *Request
		optionalProperties string
		test               func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "anyOf empty",
			req: &Request{
				Schema: schematest.NewAny(),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(-3652171958352792229), v)
			},
		},
		{
			name: "anyOf string or number",
			req: &Request{
				Schema: schematest.NewAny(
					schematest.New("string", schematest.WithMaxLength(5)),
					schematest.New("number", schematest.WithMinimum(0)),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "l", v)
			},
		},
		{
			name: "any with enum",
			req: &Request{
				Schema: schematest.NewAny(
					schematest.New("string", schematest.WithEnumValues("foo", "bar")),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", v)
			},
		},
		{
			name: "any nested",
			req: &Request{
				Schema: schematest.NewAny(
					schematest.NewAny(
						schematest.New("string", schematest.WithEnumValues("foo", "bar")),
					),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", v)
			},
		},
		{
			name: "object with anyOf and properties",
			req: &Request{
				Schema: schematest.NewTypes(nil,
					schematest.WithProperty("foo", schematest.New("string")),
					schematest.Any(
						schematest.NewTypes(nil,
							schematest.WithProperty("bar", schematest.New("string")),
							schematest.WithRequired("bar"),
						),
						schematest.NewTypes(nil,
							schematest.WithProperty("yuh", schematest.New("string")),
							schematest.WithRequired("yuh"),
						),
					),
				),
			},
			optionalProperties: "1",
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]any{"bar": "zashbEO6", "foo": "FqwCrwMfkOjojx"}, v)
			},
		},
		{
			name: "object with anyOf and properties but maxProperties=1",
			req: &Request{
				Schema: schematest.NewTypes(nil,
					schematest.WithProperty("foo", schematest.New("string")),
					schematest.WithMaxProperties(1),
					schematest.Any(
						schematest.NewTypes(nil,
							schematest.WithProperty("bar", schematest.New("string")),
							schematest.WithRequired("bar"),
						),
						schematest.NewTypes(nil,
							schematest.WithProperty("yuh", schematest.New("string")),
							schematest.WithRequired("yuh"),
						),
					),
				),
			},
			optionalProperties: "1",
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "failed to generate valid object: reached attempt limit (10) caused by: cannot apply one schema of 'anyOf': reached attempt limit (10) caused by: reached maximum of value maxProperties=1")
			},
		},
		{
			name: "implication (not A) or B",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithProperty("restaurantType", &schema.Schema{Enum: []any{"sit-down", "fast-food"}}),
					schematest.WithProperty("total", schematest.New("number")),
					schematest.WithProperty("tip", schematest.New("number",
						schematest.WithMinimum(0), schematest.WithMaximum(10000),
					)),
					schematest.WithRequired("restaurantType"),
					schematest.Any(
						schematest.NewTypes(nil, schematest.WithNot(
							schematest.NewTypes(nil,
								schematest.WithProperty("restaurantType", schematest.NewTypes(nil, schematest.WithConst("sit-down"))),
								schematest.WithRequired("restaurantType"),
							)),
						),
						schematest.NewTypes(nil, schematest.WithRequired("tip")),
					),
				),
			},
			optionalProperties: "0",
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]interface{}{
						"restaurantType": "sit-down",
						"tip":            9710.164740571692,
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
