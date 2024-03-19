package js

import (
	"encoding/json"
	"github.com/brianvoe/gofakeit/v6"
	r "github.com/stretchr/testify/require"
	"mokapi/config/static"
	"mokapi/json/generator"
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
				s, err := New(newScript("",
					`import {fake} from 'mokapi/faker'
						 export default function() {
						 	return fake({type: 'string'})
						 }`),
					host, static.JsConfig{})
				r.NoError(t, err)
				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, "XidZuoWq ", v.String())
			},
		},
		{
			"fake enum",
			func(t *testing.T, host *testHost) {
				s, err := New(newScript("",
					`import {fake} from 'mokapi/faker'
						 export default function() {
						 	return fake({type: 'string', enum: ['foo', 'bar']})
						 }`),
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
				s, err := New(newScript("",
					`import {fake} from 'mokapi/faker'
						 export default function() {
						 	return fake("")
						 }`),
					host, static.JsConfig{})
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.Error(t, err)
			},
		},
		{
			"using deprecated module",
			func(t *testing.T, host *testHost) {
				s, err := New(newScript("",
					`import {fake} from 'faker'
						 export default function() {
						 	return fake({type: 'string'})
						 }`),
					host, static.JsConfig{})
				r.NoError(t, err)
				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, "XidZuoWq ", v.String())
			},
		},
		{
			"object with properties",
			func(t *testing.T, host *testHost) {
				s, err := New(newScript("",
					`import {fake} from 'faker'
						 export default function() {
						 	return fake({type: 'object', properties: {id: {type: 'integer', format: 'int64'}}})
						 }`),
					host, static.JsConfig{})
				r.NoError(t, err)
				v, err := s.RunDefault()
				r.NoError(t, err)
				b, err := json.Marshal(v)
				r.NoError(t, err)
				r.Equal(t, `{"id":843730692693298300}`, string(b))
			},
		},
		{
			"find node",
			func(t *testing.T, host *testHost) {
				s, err := New(newScript("",
					`import {fake, findByName} from 'mokapi/faker'
						 export default function() {
						 	let root = findByName('Faker')
							root.insert(0, {
								name: 'foo',
								test: () => { return true },
								fake: (r) => {
									return {
										isString: r.lastSchema().isString(),
										schema:	r.lastSchema().string()
									}
								}
							})
							return fake({ type: 'string' })
						 }`),
					host, static.JsConfig{})
				r.NoError(t, err)
				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, map[string]interface{}{
					"isString": true,
					"schema":   "schema type=string",
				}, v.Export())
				root := generator.FindByName("")
				err = root.Remove(0)
				r.NoError(t, err)
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
