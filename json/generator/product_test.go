package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestProduct(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "name",
			req: &Request{
				Names: []string{"product", "name"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Dash Cool Fitness Tracker", v)
			},
		},
		{
			name: "description",
			req: &Request{
				Names: []string{"product", "description"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Daily regularly you for recline her choir insufficient poised tribe. Mustering will might annually backwards abroad finger say.", v)
			},
		},
		{
			name: "description",
			req: &Request{
				Names: []string{"material"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "stainless", v)
			},
		},
		{
			name: "category",
			req: &Request{
				Names: []string{"category"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "musical instruments", v)
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
