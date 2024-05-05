package js_test

import (
	r "github.com/stretchr/testify/require"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/jstest"
	"testing"
)

func TestYaml(t *testing.T) {
	host := &enginetest.Host{}
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "parse",
			test: func(t *testing.T) {
				s, err := jstest.New(jstest.WithSource(
					`import {parse} from 'mokapi/yaml';
						 export default function() {
						 	return parse('foo: bar')
						}`),
					js.WithHost(host))
				r.NoError(t, err)

				v, err := s.RunDefault()
				r.NoError(t, err)

				result := v.Export().(map[string]interface{})
				r.Equal(t, "bar", result["foo"])
			},
		},
		{
			name: "parse array",
			test: func(t *testing.T) {
				s, err := jstest.New(jstest.WithSource(
					`import {parse} from 'mokapi/yaml';
						 export default function() {
						 	return parse('- a\n- b\n- c')
						}`),
					js.WithHost(host))
				r.NoError(t, err)

				v, err := s.RunDefault()
				r.NoError(t, err)

				result := v.Export()
				r.Equal(t, []interface{}{"a", "b", "c"}, result)
			},
		},
		{
			name: "access array element",
			test: func(t *testing.T) {
				s, err := jstest.New(jstest.WithSource(
					`import {parse} from 'mokapi/yaml';
						 export default function() {
						 	const o = parse('- a\n- b\n- c')
							return o[1]
						}`),
					js.WithHost(host))
				r.NoError(t, err)

				v, err := s.RunDefault()
				r.NoError(t, err)

				result := v.Export()
				r.Equal(t, "b", result)
			},
		},
		{
			name: "stringify",
			test: func(t *testing.T) {
				s, err := jstest.New(jstest.WithSource(
					`import {stringify} from 'mokapi/yaml';
						 export default function() {
						 	return stringify({foo: "bar"})
						}`),
					js.WithHost(host))
				r.NoError(t, err)

				v, err := s.RunDefault()
				r.NoError(t, err)

				result := v.String()
				r.Equal(t, "foo: bar\n", result)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.test(t)
		})
	}
}
