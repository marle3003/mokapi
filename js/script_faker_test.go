package js

import (
	"encoding/json"
	"github.com/brianvoe/gofakeit/v6"
	r "github.com/stretchr/testify/require"
	"mokapi/config/static"
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
					host, static.JsConfig{})
				r.NoError(t, err)
				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, "xid1UOwQ;", v.String())
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
					host, static.JsConfig{})
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
					host, static.JsConfig{})
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.Error(t, err)
			},
		},
		{
			"using deprecated module",
			func(t *testing.T, host *testHost) {
				s, err := New("",
					`import {fake} from 'faker'
						 export default function() {
						 	return fake({type: 'string'})
						 }`,
					host, static.JsConfig{})
				r.NoError(t, err)
				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, "xid1UOwQ;", v.String())
			},
		}, {
			"object with properties",
			func(t *testing.T, host *testHost) {
				s, err := New("",
					`import {fake} from 'faker'
						 export default function() {
						 	return fake({type: 'object', properties: {id: {type: 'integer', format: 'int64'}}})
						 }`,
					host, static.JsConfig{})
				r.NoError(t, err)
				v, err := s.RunDefault()
				r.NoError(t, err)
				b, err := json.Marshal(v)
				r.NoError(t, err)
				r.Equal(t, `{"id":-8379641344161478000}`, string(b))
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
