package js

import (
	r "github.com/stretchr/testify/require"
	"mokapi/config/static"
	"testing"
)

func TestYaml(t *testing.T) {
	host := &testHost{}
	testcases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			"parse",
			func(t *testing.T) {
				s, err := New(newScript("test",
					`import {parse} from 'mokapi/yaml';
						 export default function() {
						 	return parse('foo: bar')
						}`),
					host, static.JsConfig{})
				r.NoError(t, err)

				v, err := s.RunDefault()
				r.NoError(t, err)

				result := v.Export().(map[string]interface{})
				r.Equal(t, "bar", result["foo"])
			},
		},
		{
			"parse array",
			func(t *testing.T) {
				s, err := New(newScript("test",
					`import {parse} from 'mokapi/yaml';
						 export default function() {
						 	return parse('- a\n- b\n- c')
						}`),
					host, static.JsConfig{})
				r.NoError(t, err)

				v, err := s.RunDefault()
				r.NoError(t, err)

				result := v.Export()
				r.Equal(t, []interface{}{"a", "b", "c"}, result)
			},
		},
		{
			"access array element",
			func(t *testing.T) {
				s, err := New(newScript("test",
					`import {parse} from 'mokapi/yaml';
						 export default function() {
						 	const o = parse('- a\n- b\n- c')
							return o[1]
						}`),
					host, static.JsConfig{})
				r.NoError(t, err)

				v, err := s.RunDefault()
				r.NoError(t, err)

				result := v.Export()
				r.Equal(t, "b", result)
			},
		},
		{
			"stringify",
			func(t *testing.T) {
				s, err := New(newScript("test",
					`import {stringify} from 'mokapi/yaml';
						 export default function() {
						 	return stringify({foo: "bar"})
						}`),
					host, static.JsConfig{})
				r.NoError(t, err)

				v, err := s.RunDefault()
				r.NoError(t, err)

				result := v.String()
				r.Equal(t, "foo: bar\n", result)
			},
		},
		{
			"using deprecated module",
			func(t *testing.T) {
				s, err := New(newScript("test",
					`import {parse} from 'yaml';
						 export default function() {
						 	return parse('foo: bar')
						}`),
					host, static.JsConfig{})
				r.NoError(t, err)

				v, err := s.RunDefault()
				r.NoError(t, err)

				result := v.Export().(map[string]interface{})
				r.Equal(t, "bar", result["foo"])
			},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc.f(t)
		})
	}
}
