package js

import (
	"fmt"
	r "github.com/stretchr/testify/require"
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
				s, err := New("",
					`export default function() {
						  	return open('foo')
						 }`,
					host)
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
				s, err := New("",
					`export default function() {
						  	return open('foo')
						 }`,
					host)
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.Error(t, err)
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
