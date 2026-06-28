package generator

import (
	"mokapi/schema/json/schema/schematest"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
)

func TestOneOf(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "oneOf only types",
			req: &Request{
				Schema: schematest.NewOneOf(
					schematest.New("integer"),
					schematest.New("string"),
				),
			},

			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(-275962), v)
			},
		},
		{
			name: "oneOf only types special case with integer and string",
			req: &Request{
				Schema: schematest.NewOneOf(
					schematest.New("string"),
					schematest.New("integer"),
				),
			},

			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "vMZsBhpyDmbo", v)
			},
		},
		{
			name: "oneOf only types special case with integer and number and integer min-max value",
			req: &Request{
				Schema: schematest.NewOneOf(
					schematest.New("integer",
						schematest.WithMinimum(1),
						schematest.WithMaximum(10),
					),
					schematest.New("integer",
						schematest.WithMinimum(3),
						schematest.WithMaximum(6),
					),
				),
			},

			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(2), v)
			},
		},
		{
			name: "oneOf",
			req: &Request{
				Path: []string{"price"},
				Schema: schematest.NewOneOf(
					schematest.New("number", schematest.WithMultipleOf(5)),
					schematest.New("number", schematest.WithMultipleOf(3)),
				),
			},

			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, float64(100), v)
			},
		},
		{
			name: "object with oneOf",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithProperty("foo", schematest.New("integer")),
					schematest.WithOneOf(
						schematest.NewTypes(nil,
							schematest.WithProperty("foo", schematest.NewTypes(nil, schematest.WithConst(123))),
						),
						schematest.NewTypes(nil,
							schematest.WithProperty("foo", schematest.NewTypes(nil, schematest.WithConst(789))),
						),
					),
				),
			},

			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": 123}, v)
			},
		},
		{
			name: "object with oneOf but not possible",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithProperty("foo", schematest.New("integer")),
					schematest.WithOneOf(
						schematest.NewTypes(nil,
							schematest.WithProperty("foo", schematest.NewTypes(nil, schematest.WithConst(123))),
						),
						schematest.NewTypes(nil,
							schematest.WithProperty("foo", schematest.NewTypes(nil, schematest.WithConst(123))),
						),
					),
				),
			},

			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "failed to generate valid object: reached attempt limit (10) caused by: cannot satisfy conditions")
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
