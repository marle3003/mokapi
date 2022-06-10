package js

import (
	r "github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestScript_Mokapi_Date(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T, host *testHost)
	}{
		{
			"now default",
			func(t *testing.T, host *testHost) {
				s, err := New("",
					`
import { date } from 'mokapi'
export default function() {
  return date({timestamp:  new Date(2022, 5, 9, 12, 0, 0, 0).getTime()}); // january is 0
}`,
					host)
				r.NoError(t, err)
				i, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, "2022-06-09T10:00:00Z", i.String())
			},
		},
		{
			"utc time ends with Z",
			func(t *testing.T, host *testHost) {
				s, err := New("",
					`
import { date } from 'mokapi'
export default function() {
  return date()
}`,
					host)
				r.NoError(t, err)
				i, err := s.RunDefault()
				r.NoError(t, err)
				r.True(t, strings.HasSuffix(i.String(), "Z"))
			},
		},
		{
			"utc time ends with Z",
			func(t *testing.T, host *testHost) {
				s, err := New("",
					`
import { date } from 'mokapi'
export default function() {
  return date({timestamp: Date.now()})
}`,
					host)
				r.NoError(t, err)
				i, err := s.RunDefault()
				r.NoError(t, err)
				r.True(t, strings.HasSuffix(i.String(), "Z"))
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
