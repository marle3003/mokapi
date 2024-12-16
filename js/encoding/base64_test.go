package encoding_test

import (
	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/engine/enginetest"
	"mokapi/js"
	mod "mokapi/js/encoding"
	"mokapi/js/eventloop"
	"mokapi/js/require"
	"testing"
)

func TestBase64(t *testing.T) {
	testcases := []struct {
		name   string
		client *enginetest.HttpClient
		test   func(t *testing.T, vm *goja.Runtime, host *enginetest.Host)
	}{
		{
			name: "encode empty string",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('mokapi/encoding')
					m.base64.encode('')
				`)
				r.NoError(t, err)
				r.Equal(t, "", v.Export())
			},
		},
		{
			name: "encode string abc",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('mokapi/encoding')
					m.base64.encode('abc')
				`)
				r.NoError(t, err)
				r.Equal(t, "YWJj", v.Export())
			},
		},
		{
			name: "encode []byte",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('mokapi/encoding')
					m.base64.encode([65, 65, 65])
				`)
				r.NoError(t, err)
				r.Equal(t, "QUFB", v.Export(), "AAA")
			},
		},
		{
			name: "encode ArrayBuffer",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('mokapi/encoding')
					const buf = new Uint8Array([104, 101, 108, 108, 111, 32, 119, 111, 114, 108, 100]).buffer;
					m.base64.encode(buf)
				`)
				r.NoError(t, err)
				r.Equal(t, "aGVsbG8gd29ybGQ=", v.Export(), "hello world")
			},
		},
		{
			name: "decode string YWJj",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('mokapi/encoding')
					m.base64.decode('YWJj')
				`)
				r.NoError(t, err)
				r.Equal(t, "abc", v.Export())
			},
		},
		{
			name: "decode string but error",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('mokapi/encoding')
					m.base64.decode('---')
				`)
				r.EqualError(t, err, "illegal base64 data at input byte 0 at mokapi/js/encoding.(*base64).Decode-fm (native)")
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			vm := goja.New()
			vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
			host := &enginetest.Host{HttpClientTest: tc.client}
			js.EnableInternal(vm, host, &eventloop.EventLoop{}, &dynamic.Config{})
			req, err := require.NewRegistry()
			r.NoError(t, err)
			req.Enable(vm)
			req.RegisterNativeModule("mokapi/encoding", mod.Require)

			tc.test(t, vm, host)
		})
	}
}
