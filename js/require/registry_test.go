package require_test

import (
	"fmt"
	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/require"
	"net/url"
	"testing"
	"time"
)

func TestRegistry(t *testing.T) {
	type source struct {
		name     string
		code     string
		checksum []byte
	}

	testcases := []struct {
		name    string
		sources map[string]source
		test    func(t *testing.T, host common.Host, sources map[string]source)
	}{
		{
			name: "module with syntax error",
			sources: map[string]source{
				"mod.js": {
					name: "mod.js",
					code: `"`,
				},
			},
			test: func(t *testing.T, host common.Host, _ map[string]source) {
				reg, err := require.NewRegistry()
				r.NoError(t, err)

				vm := goja.New()
				js.EnableInternal(vm, host, nil, &dynamic.Config{Info: dynamictest.NewConfigInfo()})
				reg.Enable(vm)

				_, err = vm.RunString(`
					const m = require("mod")
					if (m.items[0] !== 'foo') {
						throw new Error('m test failed')
					}
				`)
				r.EqualError(t, err, "loaded module mod contains error: script error: unterminated string literal: mod.js:1:1 at mokapi/js/require.(*module).require-fm (native)")
			},
		},
		{
			name: "export array",
			sources: map[string]source{
				"mod.js": {
					name: "mod.js",
					code: `export let items = ['foo']`,
				},
			},
			test: func(t *testing.T, host common.Host, _ map[string]source) {
				reg, err := require.NewRegistry()
				r.NoError(t, err)

				vm := goja.New()
				js.EnableInternal(vm, host, nil, &dynamic.Config{Info: dynamictest.NewConfigInfo()})
				reg.Enable(vm)

				_, err = vm.RunString(`
					const m = require("mod")
					if (m.items[0] !== 'foo') {
						throw new Error('m test failed')
					}
				`)
				r.NoError(t, err)
			},
		},
		{
			name: "export default object",
			sources: map[string]source{
				"mod.js": {
					name: "mod.js",
					code: `export default { foo: 'bar' }`,
				},
			},
			test: func(t *testing.T, host common.Host, _ map[string]source) {
				reg, err := require.NewRegistry()
				r.NoError(t, err)

				vm := goja.New()
				js.EnableInternal(vm, host, nil, &dynamic.Config{Info: dynamictest.NewConfigInfo()})
				reg.Enable(vm)

				_, err = vm.RunString(`
					const m = require("mod")
					if (m.default.foo !== 'bar') {
						throw new Error('m test failed')
					}
				`)
				r.NoError(t, err)
			},
		},
		{
			name: "requesting same file multiple",
			sources: map[string]source{
				"mod.js": {
					name: "mod.js",
					code: `const c = require("child")`,
				},
				"child.js": {
					name: "",
					code: "",
				},
			},
			test: func(t *testing.T, host common.Host, _ map[string]source) {
				reg, err := require.NewRegistry()
				r.NoError(t, err)

				vm := goja.New()
				js.EnableInternal(vm, host, nil, &dynamic.Config{Info: dynamictest.NewConfigInfo()})
				reg.Enable(vm)

				_, err = vm.RunString(`
					const m1 = require("mod")
					const m2 = require("child")
				`)
				r.NoError(t, err)
			},
		},
		{
			name: "update mod file",
			sources: map[string]source{
				"mod.js": {
					name:     "mod.js",
					code:     `export default 2+2`,
					checksum: []byte("1"),
				},
			},
			test: func(t *testing.T, host common.Host, sources map[string]source) {
				reg, err := require.NewRegistry()
				r.NoError(t, err)

				vm := goja.New()
				js.EnableInternal(vm, host, nil, &dynamic.Config{Info: dynamictest.NewConfigInfo()})
				reg.Enable(vm)

				v, err := vm.RunString(`
					const m = require("mod");
					m.default + 2
				`)
				r.NoError(t, err)
				r.Equal(t, int64(6), v.Export())

				// update mod
				sources["mod.js"] = source{
					name:     "",
					code:     `export default 3+3`,
					checksum: []byte("2"),
				}
				vm = goja.New()
				js.EnableInternal(vm, host, nil, &dynamic.Config{Info: dynamictest.NewConfigInfo()})
				reg.Enable(vm)

				v, err = vm.RunString(`
					const m = require("mod")
					m.default + 2
				`)
				r.NoError(t, err)
				r.Equal(t, int64(8), v.Export())
			},
		},
		{
			name: "getProgram updated file should return new result",
			test: func(t *testing.T, host common.Host, _ map[string]source) {
				reg, err := require.NewRegistry()
				r.NoError(t, err)

				p, err := reg.GetProgram(&dynamic.Config{
					Info: dynamic.ConfigInfo{
						Provider: "test",
						Url:      mustParse("foo.js"),
						Checksum: []byte("12345"),
						Time:     time.Time{},
					},
					Raw: []byte("2+2"),
				})
				r.NoError(t, err)

				// create new vm and run updated file
				vm := goja.New()
				v, err := vm.RunProgram(p)
				r.NoError(t, err)
				r.Equal(t, int64(4), v.Export())

				vm = goja.New()
				p, err = reg.GetProgram(&dynamic.Config{
					Info: dynamic.ConfigInfo{
						Provider: "test",
						Url:      mustParse("foo.js"),
						Checksum: []byte("54321"),
						Time:     time.Time{},
					},
					Raw: []byte("4+4"),
				})
				r.NoError(t, err)

				v, err = vm.RunProgram(p)
				r.Equal(t, int64(8), v.Export())
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			host := &enginetest.Host{
				OpenFunc: func(file, hint string) (*dynamic.Config, error) {
					if src, ok := tc.sources[file]; ok {
						return &dynamic.Config{
							Info: dynamic.ConfigInfo{
								Provider: "test",
								Url:      mustParse(src.name),
								Checksum: src.checksum,
								Time:     time.Time{},
							},
							Raw:       []byte(src.code),
							Data:      nil,
							Refs:      dynamic.Refs{},
							Listeners: dynamic.Listeners{},
						}, nil
					}
					return nil, fmt.Errorf("file not found")
				},
			}

			tc.test(t, host, tc.sources)
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
