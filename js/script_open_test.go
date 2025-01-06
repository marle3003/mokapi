package js_test

import (
	"fmt"
	r "github.com/stretchr/testify/require"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/jstest"
	"testing"
)

func TestScript_Open(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, host *enginetest.Host)
	}{
		{
			name: "open",
			test: func(t *testing.T, host *enginetest.Host) {
				host.OpenFileFunc = func(file, hint string) (string, string, error) {
					return "", "bar", nil
				}
				s, err := jstest.New(jstest.WithSource(
					`export default function() {
						  	return open('foo')
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, "bar", v.String())
			},
		},
		{
			name: "file not found",
			test: func(t *testing.T, host *enginetest.Host) {
				host.OpenFileFunc = func(file, hint string) (string, string, error) {
					return "", "", fmt.Errorf("test error")
				}
				s, err := jstest.New(jstest.WithSource(
					`export default function() {
						  	return open('foo')
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.Error(t, err)
			},
		},
		{
			name: "file as binary",
			test: func(t *testing.T, host *enginetest.Host) {
				host.OpenFileFunc = func(file, hint string) (string, string, error) {
					return "", "foo", nil
				}
				s, err := jstest.New(jstest.WithSource(
					`export default function() {
						  	return open('foo', { as: 'binary' })
						 }`),
					js.WithHost(host))
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
			tc.test(t, &enginetest.Host{})
		})
	}
}
