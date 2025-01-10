package dynamic_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"testing"
)

func TestScope(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "name not defined",
			test: func(t *testing.T) {
				s := dynamic.NewScope("https://mokapi.io/foo")
				_, err := s.GetLexical("foo")
				require.EqualError(t, err, "name 'foo' not found in scope 'https://mokapi.io/foo'")
			},
		},
		{
			name: "name defined",
			test: func(t *testing.T) {
				s := dynamic.NewScope("https://mokapi.io/foo")
				err := s.SetLexical("foo", 1)
				require.NoError(t, err)

				v, err := s.GetLexical("foo")
				require.NoError(t, err)
				require.Equal(t, 1, v)
			},
		},
		{
			name: "name already defined",
			test: func(t *testing.T) {
				s := dynamic.NewScope("https://mokapi.io/foo")
				err := s.SetLexical("foo", 1)
				require.NoError(t, err)

				err = s.SetLexical("foo", 2)
				require.EqualError(t, err, "name 'foo' already defined in scope 'https://mokapi.io/foo'")
			},
		},
		{
			name: "name not defined in current scope",
			test: func(t *testing.T) {
				s := dynamic.NewScope("https://mokapi.io/foo")

				err := s.SetLexical("foo", 1)
				require.NoError(t, err)

				s.Open("https://mokapi.io/bar")

				_, err = s.GetLexical("foo")
				require.EqualError(t, err, "name 'foo' not found in scope 'https://mokapi.io/bar'")
			},
		},
		{
			name: "set name not defined in current scope",
			test: func(t *testing.T) {
				s := dynamic.NewScope("https://mokapi.io/foo")

				err := s.SetLexical("foo", 1)
				require.NoError(t, err)

				s.Open("https://mokapi.io/bar")

				err = s.SetLexical("foo", 2)
				require.NoError(t, err)
			},
		},
		{
			name: "name should be found after open/close a child scope",
			test: func(t *testing.T) {
				s := dynamic.NewScope("https://mokapi.io/foo")
				err := s.SetLexical("foo", 1)
				require.NoError(t, err)

				s.Open("https://mokapi.io/bar")
				s.Close()

				v, err := s.GetLexical("foo")
				require.NoError(t, err)
				require.Equal(t, 1, v)
			},
		},
		{
			name: "name not defined in current scope but in global scope",
			test: func(t *testing.T) {
				s := dynamic.NewScope("https://mokapi.io/foo")

				err := s.SetDynamic("foo", 1)
				require.NoError(t, err)

				s.Open("https://mokapi.io/bar")

				v, err := s.GetDynamic("foo")
				require.NoError(t, err)
				require.Equal(t, 1, v)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.test(t)
		})
	}
}
