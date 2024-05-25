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
		name string
		code string
	}

	testcases := []struct {
		name    string
		sources map[string]source
		test    func(t *testing.T, host common.Host)
	}{
		{
			name: "export array",
			sources: map[string]source{
				"mod.js": {
					name: "",
					code: `export let items = ['foo']`,
				},
			},
			test: func(t *testing.T, host common.Host) {
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
					name: "",
					code: `export default { foo: 'bar' }`,
				},
			},
			test: func(t *testing.T, host common.Host) {
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
					name: "",
					code: `const c = require("child")`,
				},
				"child.js": {
					name: "",
					code: "",
				},
			},
			test: func(t *testing.T, host common.Host) {
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
								Checksum: nil,
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

			tc.test(t, host)
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
