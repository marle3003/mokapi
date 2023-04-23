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
			"fake string",
			func(t *testing.T, host *testHost) {
				s, err := New("",
					`import {fake} from 'mokapi/faker'
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
		{
			"fake enum",
			func(t *testing.T, host *testHost) {
				s, err := New("",
					`import {fake} from 'mokapi/faker'
						 export default function() {
						 	return fake({type: 'string', enum: ['foo', 'bar']})
						 }`,
					host)
				r.NoError(t, err)
				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, "bar", v.String())
			},
		},
		{
			"invalid format",
			func(t *testing.T, host *testHost) {
				s, err := New("",
					`import {fake} from 'mokapi/faker'
						 export default function() {
						 	return fake("")
						 }`,
					host)
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.Error(t, err)
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
