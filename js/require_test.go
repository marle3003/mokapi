package js

import (
	"fmt"
	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
	"testing"
)

func TestRequire(t *testing.T) {
	host := &testHost{}
	testcases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			"mokapi",
			func(t *testing.T) {
				s, err := New("test", `import {sleep} from 'mokapi'; export let _sleep = sleep; sleep(12); export default function() {}`, host)
				r.NoError(t, err)

				r.NoError(t, s.Run())

				exports := s.runtime.Get("exports").ToObject(s.runtime)
				_, ok := goja.AssertFunction(exports.Get("_sleep"))
				r.True(t, ok, "sleep is not a function")
			},
		},
		{
			"require custom file",
			func(t *testing.T) {
				host.openScript = func(file string) (string, error) {
					// first request is foo, second is foo.js
					if file == "foo" {
						return "", fmt.Errorf("TEST ERROR NOT FOUND")
					}
					r.Equal(t, "foo.js", file)
					return "export var bar = {demo: 'demo'};", nil
				}
				host.info = func(args ...interface{}) {
					r.Equal(t, "demo", args[0])
				}
				s, err := New("test", `import {bar} from 'foo'; export default function() {console.log(bar.demo);}`, host)
				r.NoError(t, err)

				r.NoError(t, s.Run())
			},
		},
		{
			"require json file",
			func(t *testing.T) {
				host.openScript = func(file string) (string, error) {
					return `{"foo":"bar"}`, nil
				}
				s, err := New("test", `import bar from 'foo.json'; export default function() {return bar.foo;}`, host)
				r.NoError(t, err)

				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, "bar", v.Export())
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
