package js

import (
	"fmt"
	r "github.com/stretchr/testify/require"
	"mokapi/config/static"
	"testing"
)

func TestScript_Open(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T, host *testHost)
	}{
		{
			"open",
			func(t *testing.T, host *testHost) {
				host.openFile = func(file, hint string) (string, string, error) {
					return "", "bar", nil
				}
				s, err := New(newScript("",
					`export default function() {
						  	return open('foo')
						 }`),
					host, static.JsConfig{})
				r.NoError(t, err)
				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, "bar", v.String())
			},
		},
		{
			"file not found",
			func(t *testing.T, host *testHost) {
				host.openFile = func(file, hint string) (string, string, error) {
					return "", "", fmt.Errorf("test error")
				}
				s, err := New(newScript("",
					`export default function() {
						  	return open('foo')
						 }`),
					host, static.JsConfig{})
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.Error(t, err)
			},
		},
		{
			"file as binary",
			func(t *testing.T, host *testHost) {
				host.openFile = func(file, hint string) (string, string, error) {
					return "", "foo", nil
				}
				s, err := New(newScript("",
					`export default function() {
						  	return open('foo', { as: 'binary' })
						 }`),
					host, static.JsConfig{})
				r.NoError(t, err)
				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, []uint8{'f', 'o', 'o'}, v.Export())
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			host := &testHost{}

			tc.f(t, host)
		})
	}
}
