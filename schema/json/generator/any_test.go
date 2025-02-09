package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/schema/schematest"
	"testing"
)

func TestAny(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "any with enum",
			req: &Request{
				Path: Path{
					&PathElement{Schema: schematest.NewAny(
						schematest.New("string", schematest.WithEnumValues("foo", "bar")),
					)},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", v)
			},
		},
		{
			name: "any nested",
			req: &Request{
				Path: Path{
					&PathElement{Schema: schematest.NewAny(
						schematest.NewAny(
							schematest.New("string", schematest.WithEnumValues("foo", "bar")),
						),
					)},
				},
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
