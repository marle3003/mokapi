package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestString(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "language",
			req:  &Request{Names: []string{"language"}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "nl-BE", v)
			},
		},
		{
			name: "languages",
			req:  &Request{Names: []string{"langs"}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"pt-BR", "pt-PT"}, v)
			},
		},
		{
			name: "error",
			req:  &Request{Names: []string{"error"}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "named cookie not present", v)
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
