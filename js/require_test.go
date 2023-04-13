package js

import (
	"fmt"
	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
	"path/filepath"
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
				host.openFile = func(file, hint string) (string, string, error) {
					// first request is foo, second is foo.js
					if file == "foo" {
						return "", "", fmt.Errorf("TEST ERROR NOT FOUND")
					}
					r.Equal(t, "foo.js", file)
					return "", "export var bar = {demo: 'demo'};", nil
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
			"require custom relative file",
			func(t *testing.T) {
				host.openFile = func(file, hint string) (string, string, error) {
					r.Equal(t, "./foo.js", file)
					return "", "export var bar = {demo: 'demo'};", nil
				}
				host.info = func(args ...interface{}) {
					r.Equal(t, "demo", args[0])
				}
				s, err := New("test", `import {bar} from './foo.js'; export default function() {return bar}`, host)
				r.NoError(t, err)

				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, map[string]interface{}{"demo": "demo"}, v.Export())
			},
		},
		{
			"require json file",
			func(t *testing.T) {
				host.openFile = func(file, hint string) (string, string, error) {
					return "", `{"foo":"bar"}`, nil
				}
				s, err := New("test", `import bar from 'foo.json'; export default function() {return bar.foo;}`, host)
				r.NoError(t, err)

				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, "bar", v.Export())
			},
		},
		{
			"require yaml file",
			func(t *testing.T) {
				host.openFile = func(file, hint string) (string, string, error) {
					return "", `foo: bar`, nil
				}
				s, err := New("test", `import x from 'foo.yaml'; export default function() {return x.foo;}`, host)
				r.NoError(t, err)

				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, "bar", v.Export())
			},
		},
		{
			"require http",
			func(t *testing.T) {
				host.openFile = func(file, hint string) (string, string, error) {
					r.Equal(t, "http://foo.bar", file)
					return "", `export var bar = {demo: 'demo'}`, nil
				}
				s, err := New("test", `import {bar} from 'http://foo.bar'; export default function() {return bar}`, host)
				r.NoError(t, err)

				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, map[string]interface{}{"demo": "demo"}, v.Export())
			},
		},
		{
			"require node module with package.json and main",
			func(t *testing.T) {
				host.openFile = func(file, hint string) (string, string, error) {
					switch file {
					case filepath.Join("node_modules", "uuid", "package.json"):
						return file, `{"main": "./dist/index.js"}`, nil
					case filepath.Join("node_modules", "uuid", "dist", "index.js"):
						return file, "export function v4() { return 'abc-def' }", nil
					}
					return "", "", fmt.Errorf("not found")
				}
				s, err := New("test", `import {v4 as uuidv4} from 'uuid'; export default () => uuidv4()`, host)
				r.NoError(t, err)

				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, "abc-def", v.Export())
			},
		},
		{
			"require node module with index.js",
			func(t *testing.T) {
				host.openFile = func(file, hint string) (string, string, error) {
					switch file {
					case filepath.Join("node_modules", "uuid", "index.js"):
						return file, "export function v4() { return 'abc-def' }", nil
					}
					return "", "", fmt.Errorf("not found")
				}
				s, err := New("test", `import {v4 as uuidv4} from 'uuid'; export default () => uuidv4()`, host)
				r.NoError(t, err)

				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, "abc-def", v.Export())
			},
		},
		{
			"require node module in parent folder",
			func(t *testing.T) {
				host.openFile = func(file, hint string) (string, string, error) {
					switch file {
					case filepath.Join("/foo", "node_modules", "uuid", "index.js"):
						return file, "export function v4() { return 'abc-def' }", nil
					}
					return "", "", fmt.Errorf("not found")
				}
				s, err := New("/foo/bar/test", `import {v4 as uuidv4} from 'uuid'; export default () => uuidv4()`, host)
				r.NoError(t, err)

				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, "abc-def", v.Export())
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
