package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/schema/schematest"
	"testing"
)

func TestLanguage(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "language",
			req: &Request{
				Path:   []string{"language"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "sm", v)
			},
		},
		{
			name: "language with max=2",
			req: &Request{
				Path: []string{"language"},
				Schema: schematest.New("string",
					schematest.WithMaxLength(2),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "sm", v)
			},
		},
		{
			name: "language with max=5",
			req: &Request{
				Path:   []string{"language"},
				Schema: schematest.New("string", schematest.WithMaxLength(5)),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "nl-BE", v)
			},
		},
		{
			name: "language with max=15",
			req: &Request{
				Path:   []string{"language"},
				Schema: schematest.New("string", schematest.WithMaxLength(15)),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Samoan", v)
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
