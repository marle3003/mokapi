package js

import (
	"github.com/brianvoe/gofakeit/v6"
	r "github.com/stretchr/testify/require"
	"testing"
)

func TestScript_Faker(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T, host *testHost)
	}{
		{
			"set topic",
			func(t *testing.T, host *testHost) {
				s, err := New("",
					`
import {fake} from 'faker'
export default function() {
  return fake({type: 'string'})
}`,
					host)
				r.NoError(t, err)
				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, "gbRMaRxHkiJBPta", v.String())
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(11)
			host := &testHost{}

			tc.f(t, host)
		})
	}
}
