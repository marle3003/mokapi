package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/json/schematest"
	"testing"
)

func TestColor(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "color name",
			req: &Request{
				Path: Path{
					&PathElement{Name: "color", Schema: schematest.NewRef("string")},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "LightYellow", v)
			},
		},
		{
			name: "hex color",
			req: &Request{
				Path: Path{
					&PathElement{Name: "color", Schema: schematest.NewRef("string", schematest.WithMaxLength(7))},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "#cea93b", v)
			},
		},
		{
			name: "rgb color",
			req: &Request{
				Path: Path{
					&PathElement{Name: "color", Schema: schematest.NewRef("array",
						schematest.WithItems("integer"),
					)},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []int{126, 91, 203}, v)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(1234567)
			Seed(1234567)

			v, err := New(tc.req)
			tc.test(t, v, err)
		})
	}
}
