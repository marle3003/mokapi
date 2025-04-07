package v2

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/schema/schematest"
	"testing"
)

func TestAnyOf(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
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
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(1234567)

			v, err := New(tc.req)
			tc.test(t, v, err)
		})
	}
}
