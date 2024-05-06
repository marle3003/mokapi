package js_test

import (
	"fmt"
	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/jstest"
	"mokapi/js/require"
	"net/url"
	"path/filepath"
	"testing"
	"time"
)

func TestRequire(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, host *enginetest.Host, registry *require.Registry)
	}{
		{
			name: "module not found",
			test: func(t *testing.T, host *enginetest.Host, registry *require.Registry) {
				host.OpenFileFunc = func(file, hint string) (string, string, error) {
					return "", "", fmt.Errorf("file not found")
				}
				s, err := jstest.New(jstest.WithSource(`import foo from 'foo'`), js.WithHost(host), js.WithRegistry(registry))
				r.NoError(t, err)

				_, err = s.RunDefault()
				r.EqualError(t, err, "module foo not found in test.js at mokapi/js/require.(*module).require-fm (native)")
			},
		},
		{
			name: "require module mokapi",
			test: func(t *testing.T, host *enginetest.Host, registry *require.Registry) {
				s, err := jstest.New(jstest.WithSource(`import {sleep} from 'mokapi'; export let _sleep = sleep; sleep(12); export default function() {}`), js.WithHost(host), js.WithRegistry(registry))
				r.NoError(t, err)

				r.NoError(t, s.Run())

				var exports *goja.Object
				err = s.RunFunc(func(vm *goja.Runtime) {
					exports = vm.Get("exports").ToObject(vm)
				})
				time.Sleep(500 * time.Millisecond)
				r.NoError(t, err)
				_, ok := goja.AssertFunction(exports.Get("_sleep"))
				r.True(t, ok, "sleep is not a function")
			},
		},
		{
			name: "require custom file",
			test: func(t *testing.T, host *enginetest.Host, registry *require.Registry) {
				host.OpenFileFunc = func(file, hint string) (string, string, error) {
					// first request is foo, second is foo.js
					if file == "foo" {
						return "", "", fmt.Errorf("TEST ERROR NOT FOUND")
					}
					r.Equal(t, "foo.js", file)
					return "", "export var bar = {demo: 'demo'};", nil
				}
				host.InfoFunc = func(args ...interface{}) {
					r.Equal(t, "demo", args[0])
				}
				s, err := jstest.New(jstest.WithSource(`import {bar} from 'foo'; export default function() {console.log(bar.demo);}`), js.WithHost(host), js.WithRegistry(registry))
				r.NoError(t, err)

				r.NoError(t, s.Run())
			},
		},
		{
			name: "require custom typescript file",
			test: func(t *testing.T, host *enginetest.Host, registry *require.Registry) {
				host.OpenFileFunc = func(file, hint string) (string, string, error) {
					if file == "foo.ts" {
						return "", "export var bar = {demo: 'demo'};", nil
					}
					return "", "", fmt.Errorf("TEST ERROR NOT FOUND")
				}
				host.InfoFunc = func(args ...interface{}) {
					r.Equal(t, "demo", args[0])
				}
				s, err := jstest.New(jstest.WithSource(`import {bar} from 'foo'; export default function() {console.log(bar.demo);}`), js.WithHost(host), js.WithRegistry(registry))
				r.NoError(t, err)

				r.NoError(t, s.Run())
			},
		},
		{
			name: "require custom relative file",
			test: func(t *testing.T, host *enginetest.Host, registry *require.Registry) {
				host.OpenFileFunc = func(file, hint string) (string, string, error) {
					file = filepath.ToSlash(file) // if on windows
					r.Equal(t, "/foo/bar/foo.js", file)
					return "", "export var bar = {demo: 'demo'};", nil
				}
				host.InfoFunc = func(args ...interface{}) {
					r.Equal(t, "demo", args[0])
				}
				s, err := jstest.New(jstest.WithPathSource("/foo/bar/test.js", `import {bar} from './foo.js'; export default function() {return bar}`), js.WithHost(host), js.WithRegistry(registry))
				r.NoError(t, err)

				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, map[string]interface{}{"demo": "demo"}, v.Export())
			},
		},
		{
			name: "require json file",
			test: func(t *testing.T, host *enginetest.Host, registry *require.Registry) {
				host.OpenFileFunc = func(file, hint string) (string, string, error) {
					return "", `{"foo":"bar"}`, nil
				}
				s, err := jstest.New(jstest.WithSource(`import bar from 'foo.json'; export default function() {return bar.foo;}`), js.WithHost(host), js.WithRegistry(registry))
				r.NoError(t, err)

				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, "bar", v.Export())
			},
		},
		{
			name: "require yaml file",
			test: func(t *testing.T, host *enginetest.Host, registry *require.Registry) {
				host.OpenFileFunc = func(file, hint string) (string, string, error) {
					return "", `foo: bar`, nil
				}
				s, err := jstest.New(jstest.WithSource(`import x from 'foo.yaml'; export default function() {return x.foo;}`), js.WithHost(host), js.WithRegistry(registry))
				r.NoError(t, err)

				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, "bar", v.Export())
			},
		},
		{
			name: "require http",
			test: func(t *testing.T, host *enginetest.Host, registry *require.Registry) {
				host.OpenFileFunc = func(file, hint string) (string, string, error) {
					r.Equal(t, "https://foo.bar", file)
					return "", `export var bar = {demo: 'demo'}`, nil
				}
				s, err := jstest.New(jstest.WithSource(`import {bar} from 'https://foo.bar'; export default function() {return bar}`), js.WithHost(host), js.WithRegistry(registry))
				r.NoError(t, err)

				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, map[string]interface{}{"demo": "demo"}, v.Export())
			},
		},
		{
			name: "require http but script error",
			test: func(t *testing.T, host *enginetest.Host, registry *require.Registry) {
				host.OpenFileFunc = func(file, hint string) (string, string, error) {
					return "", `foo`, nil
				}
				s, err := jstest.New(jstest.WithSource(`import bar from 'https://foo.bar'`), js.WithHost(host), js.WithRegistry(registry))
				r.NoError(t, err)

				_, err = s.RunDefault()
				r.EqualError(t, err, "module https://foo.bar not found in test.js at mokapi/js/require.(*module).require-fm (native)")
			},
		},
		{
			name: "require node module with package.json and main",
			test: func(t *testing.T, host *enginetest.Host, registry *require.Registry) {
				host.OpenFileFunc = func(file, hint string) (string, string, error) {
					file = filepath.ToSlash(file) // if on windows
					switch {
					case file == "/foo/bar/node_modules/uuid/package.json":
						return file, `{"main": "./dist/index.js"}`, nil
					case file == "/foo/bar/node_modules/uuid/dist/index.js":
						return file, "export function v4() { return 'abc-def' }", nil
					}
					return "", "", fmt.Errorf("not found")
				}
				s, err := jstest.New(jstest.WithPathSource(`/foo/bar/test.js`, `import {v4 as uuidv4} from 'uuid'; export default () => uuidv4()`), js.WithHost(host), js.WithRegistry(registry))
				r.NoError(t, err)

				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, "abc-def", v.Export())
			},
		},
		{
			name: "require node module with index.js",
			test: func(t *testing.T, host *enginetest.Host, registry *require.Registry) {
				host.OpenFileFunc = func(file, hint string) (string, string, error) {
					file = filepath.ToSlash(file) // if on windows
					switch {
					case file == "/foo/bar/node_modules/uuid/index.js":
						return file, "export function v4() { return 'abc-def' }", nil
					}
					return "", "", fmt.Errorf("not found")
				}
				s, err := jstest.New(jstest.WithPathSource(`/foo/bar/test.js`, `import {v4 as uuidv4} from 'uuid'; export default () => uuidv4()`), js.WithHost(host), js.WithRegistry(registry))
				r.NoError(t, err)

				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, "abc-def", v.Export())
			},
		},
		{
			name: "require custom module with index.js",
			test: func(t *testing.T, host *enginetest.Host, registry *require.Registry) {
				host.OpenFileFunc = func(file, hint string) (string, string, error) {
					file = filepath.ToSlash(file) // if on windows
					switch {
					case file == "/foo/users/index.js":
						return file, "export const users = ['bob', 'alice']", nil
					}
					return "", "", fmt.Errorf("not found")
				}
				s, err := jstest.New(jstest.WithPathSource(`/foo/bar/test.js`, `import { users } from '../users'; export default () => users`), js.WithHost(host), js.WithRegistry(registry))
				r.NoError(t, err)

				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, []interface{}{"bob", "alice"}, v.Export())
			},
		},
		{
			name: "require node module in parent folder",
			test: func(t *testing.T, host *enginetest.Host, registry *require.Registry) {
				host.OpenFileFunc = func(file, hint string) (string, string, error) {
					file = filepath.ToSlash(file) // if on windows
					switch {
					case file == "/foo/node_modules/uuid/index.js":
						return file, "export function v4() { return 'abc-def' }", nil
					}
					return "", "", fmt.Errorf("not found")
				}
				s, err := jstest.New(jstest.WithPathSource(`/foo/bar/test.js`, `import {v4 as uuidv4} from 'uuid'; export default () => uuidv4()`), js.WithHost(host), js.WithRegistry(registry))
				r.NoError(t, err)

				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, "abc-def", v.Export())
			},
		},
		{
			name: "require file with same name but different folder",
			test: func(t *testing.T, host *enginetest.Host, registry *require.Registry) {
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

				host.OpenFileFunc = func(file, hint string) (string, string, error) {
					file = filepath.ToSlash(file) // if on windows
					switch {
					case file == "/data.json":
						return file, dataRoot, nil
					case file == "/foo/foo.js":
						return file, foojs, nil
					case file == "/foo/data.json":
						return file, dataChild, nil
					}
					return "", "", fmt.Errorf("not found")
				}
				s, err := jstest.New(jstest.WithPathSource(`/test.js`, testjs), js.WithHost(host), js.WithRegistry(registry))
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
			name: "require node module in parent folder with nested provider",
			test: func(t *testing.T, host *enginetest.Host, registry *require.Registry) {
				host.OpenFileFunc = func(file, hint string) (string, string, error) {
					file = filepath.ToSlash(file) // if on windows
					switch {
					case file == "/foo/node_modules/uuid/index.js":
						return file, "export function v4() { return 'abc-def' }", nil
					}
					return "", "", fmt.Errorf("not found")
				}
				f := &dynamic.Config{Info: dynamic.ConfigInfo{Url: mustParse("/foo/bar/test.js")}, Raw: []byte(`import {v4 as uuidv4} from 'uuid'; export default () => uuidv4()`)}
				dynamic.Wrap(dynamic.ConfigInfo{Provider: "git", Url: mustParse("git://foo.bar")}, f)
				s, err := jstest.New(js.WithFile(f), js.WithHost(host), js.WithRegistry(registry))
				r.NoError(t, err)

				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, "abc-def", v.Export())
			},
		},
		{
			name: "require file with nested provider",
			test: func(t *testing.T, host *enginetest.Host, registry *require.Registry) {
				foo := &dynamic.Config{Info: dynamic.ConfigInfo{Url: mustParse("/foo/foo.js")}, Raw: []byte(`import users from '../users'; export function foo() { return users }`)}
				dynamic.Wrap(dynamic.ConfigInfo{Provider: "git", Url: mustParse("https://git.bar/projects/mokapi.git?file=/foo/foo.js&ref=main")}, foo)

				users := &dynamic.Config{Info: dynamic.ConfigInfo{Url: mustParse("/users.json")}, Raw: []byte(`["user1", "user2"]`)}
				dynamic.Wrap(dynamic.ConfigInfo{Provider: "git", Url: mustParse("https://git.bar/projects/mokapi.git?file=/users.json&ref=main")}, users)

				host.OpenFunc = func(file, hint string) (*dynamic.Config, error) {
					file = filepath.ToSlash(file) // if on windows
					switch {
					case file == "/foo/foo.js":
						return foo, nil
					case file == "/users.json":
						return users, nil
					}
					return nil, fmt.Errorf("not found")
				}

				index := &dynamic.Config{Info: dynamic.ConfigInfo{Url: mustParse("/foo/index.js")}, Raw: []byte(`import { foo } from './foo'; export default () => foo()`)}
				dynamic.Wrap(dynamic.ConfigInfo{Provider: "git", Url: mustParse("https://git.bar/projects/mokapi.git?file=/foo/index.js&ref=main")}, index)
				s, err := jstest.New(js.WithFile(index), js.WithHost(host), js.WithRegistry(registry))
				r.NoError(t, err)

				v, err := s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, []interface{}{"user1", "user2"}, v.Export())
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			host := &enginetest.Host{}
			registry, err := require.NewRegistry()
			js.RegisterNativeModules(registry)
			r.NoError(t, err)

			tc.test(t, host, registry)
		})
	}
}

func mustParse(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}
