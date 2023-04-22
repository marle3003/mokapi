package js

import (
	r "github.com/stretchr/testify/require"
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
				s, err := New("test",
					`import {parse} from 'mokapi/yaml';
						 export default function() {
						 	return parse('foo: bar')
						}`,
					host)
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
				s, err := New("test",
					`import {parse} from 'mokapi/yaml';
						 export default function() {
						 	return parse('- a\n- b\n- c')
						}`,
					host)
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
				s, err := New("test",
					`import {parse} from 'mokapi/yaml';
						 export default function() {
						 	const o = parse('- a\n- b\n- c')
							return o[1]
						}`,
					host)
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
				s, err := New("test",
					`import {stringify} from 'mokapi/yaml';
						 export default function() {
						 	return stringify({foo: "bar"})
						}`,
					host)
				r.NoError(t, err)

				v, err := s.RunDefault()
				r.NoError(t, err)

				result := v.String()
				r.Equal(t, "foo: bar\n", result)
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
