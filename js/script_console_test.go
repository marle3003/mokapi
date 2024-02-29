package js

import (
	"github.com/brianvoe/gofakeit/v6"
	r "github.com/stretchr/testify/require"
	"mokapi/config/static"
	"testing"
)

func TestScript_Console(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T, host *testHost)
	}{
		{
			"log",
			func(t *testing.T, host *testHost) {
				var log interface{}
				host.info = func(args ...interface{}) {
					log = args[0]
				}
				s, err := New("",
					`import {fake} from 'mokapi/faker'
						 export default function() {
						 	console.log('foo')
						 }`,
					host, static.JsConfig{})
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, "foo", log)
			},
		},
		{
			"log",
			func(t *testing.T, host *testHost) {
				var log interface{}
				host.info = func(args ...interface{}) {
					log = args[0]
				}
				s, err := New("",
					`import {fake} from 'mokapi/faker'
						 export default function() {
						 	console.log({foo: 123, bar: 'mokapi'})
						 }`,
					host, static.JsConfig{})
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, `{"foo":123,"bar":"mokapi"}`, log)
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
