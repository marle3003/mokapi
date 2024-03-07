package js

import (
	"fmt"
	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/static"
	"path/filepath"
	"testing"
	"time"
)

func TestRequire(t *testing.T) {
	host := &testHost{}
	testcases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			"module not found",
			func(t *testing.T) {
				host.openFile = func(file, hint string) (string, string, error) {
					return "", "", fmt.Errorf("file not found")
				}
				s, err := New(newScript("test", `import foo from 'foo'`), host, static.JsConfig{})
				r.NoError(t, err)

				err = s.Run()
				r.EqualError(t, err, "module foo not found: node module does not exist at mokapi/js.(*requireModule).require-fm (native)")
			},
		},
		{
			"mokapi",
			func(t *testing.T) {
				s, err := New(newScript("test", `import {sleep} from 'mokapi'; export let _sleep = sleep; sleep(12); export default function() {}`), host, static.JsConfig{})
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
				s, err := New(newScript("test", `import {bar} from 'foo'; export default function() {console.log(bar.demo);}`), host, static.JsConfig{})
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
				s, err := New(newScript("C:\\foo\\bar\\test.js", `import {bar} from './foo.js'; export default function() {return bar}`), host, static.JsConfig{})
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
				s, err := New(newScript("test", `import bar from 'foo.json'; export default function() {return bar.foo;}`), host, static.JsConfig{})
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
				s, err := New(newScript("test", `import x from 'foo.yaml'; export default function() {return x.foo;}`), host, static.JsConfig{})
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
				s, err := New(newScript("test", `import {bar} from 'http://foo.bar'; export default function() {return bar}`), host, static.JsConfig{})
				r.NoError(t, err)

				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, map[string]interface{}{"demo": "demo"}, v.Export())
			},
		},
		{
			"require http but script error",
			func(t *testing.T) {
				host.openFile = func(file, hint string) (string, string, error) {
					return "", `foo`, nil
				}
				s, err := New(newScript("test", `import bar from 'http://foo.bar'`), host, static.JsConfig{})
				r.NoError(t, err)

				err = s.Run()
				r.EqualError(t, err, "ReferenceError: foo is not defined at http://foo.bar:1:42(1)")
			},
		},
		{
			"require node module with package.json and main",
			func(t *testing.T) {
				host.openFile = func(file, hint string) (string, string, error) {
					hint = filepath.ToSlash(hint) // if on windows
					switch {
					case file == "package.json" && hint == "/foo/bar/node_modules/uuid":
						return file, `{"main": "./dist/index.js"}`, nil
					case file == "./dist/index.js" && hint == "/foo/bar/node_modules/uuid":
						return file, "export function v4() { return 'abc-def' }", nil
					}
					return "", "", fmt.Errorf("not found")
				}
				s, err := New(newScript(`/foo/bar/test.js`, `import {v4 as uuidv4} from 'uuid'; export default () => uuidv4()`), host, static.JsConfig{})
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
					hint = filepath.ToSlash(hint) // if on windows
					switch {
					case file == "index.js" && hint == "/foo/bar/node_modules/uuid":
						return file, "export function v4() { return 'abc-def' }", nil
					}
					return "", "", fmt.Errorf("not found")
				}
				s, err := New(newScript(`/foo/bar/test.js`, `import {v4 as uuidv4} from 'uuid'; export default () => uuidv4()`), host, static.JsConfig{})
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
					hint = filepath.ToSlash(hint) // if on windows
					switch {
					case file == "index.js" && hint == "/foo/node_modules/uuid":
						return file, "export function v4() { return 'abc-def' }", nil
					}
					return "", "", fmt.Errorf("not found")
				}
				s, err := New(newScript(`/foo/bar/test.js`, `import {v4 as uuidv4} from 'uuid'; export default () => uuidv4()`), host, static.JsConfig{})
				r.NoError(t, err)

				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, "abc-def", v.Export())
			},
		},
		{
			"require file with same name but different folder",
			func(t *testing.T) {
				testjs := `
import foo from './foo/foo.js'
import data from './data.json'

export default function () {
	return {
		data: data,
		foo: foo()
	}
}`
				dataRoot := `{"root": true }`
				foojs := `
import data from './data.json'
export default function () {return data}
`
				dataChild := `{"root": false }`

				host.openFile = func(file, hint string) (string, string, error) {
					hint = filepath.ToSlash(hint) // if on windows
					switch {
					case file == "./data.json" && hint == "/":
						return file, dataRoot, nil
					case file == "./foo/foo.js":
						return file, foojs, nil
					case file == "./data.json" && hint == "foo":
						return file, dataChild, nil
					}
					return "", "", fmt.Errorf("not found")
				}
				s, err := New(newScript(`/test.js`, testjs), host, static.JsConfig{})
				r.NoError(t, err)

				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, map[string]interface{}{
					"data": map[string]interface{}{"root": true},
					"foo":  map[string]interface{}{"root": false},
				}, v.Export())
			},
		},
		{
			"require node module in parent folder with nested provider",
			func(t *testing.T) {
				host.openFile = func(file, hint string) (string, string, error) {
					hint = filepath.ToSlash(hint) // if on windows
					switch {
					case file == "index.js" && hint == "/foo/node_modules/uuid":
						return file, "export function v4() { return 'abc-def' }", nil
					}
					return "", "", fmt.Errorf("not found")
				}
				f := &dynamic.Config{Info: dynamic.ConfigInfo{Url: mustParse("/foo/bar/test.js")}, Raw: []byte(`import {v4 as uuidv4} from 'uuid'; export default () => uuidv4()`)}
				dynamic.Wrap(dynamic.ConfigInfo{Provider: "git", Url: mustParse("git://foo.bar")}, f)
				s, err := New(f, host, static.JsConfig{})
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

func newScript(path string, code string) *dynamic.Config {
	return &dynamic.Config{
		Info: dynamic.ConfigInfo{
			Provider: "test",
			Url:      mustParse(path),
			Checksum: nil,
			Time:     time.Time{},
		},
		Raw:       []byte(code),
		Data:      nil,
		Refs:      dynamic.Refs{},
		Listeners: dynamic.Listeners{},
	}
}
