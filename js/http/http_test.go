package http_test

import (
	"fmt"
	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
	"io"
	"mokapi/config/dynamic"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/eventloop"
	mod "mokapi/js/http"
	"mokapi/js/require"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestHttp(t *testing.T) {
	testcases := []struct {
		name   string
		client *enginetest.HttpClient
		test   func(t *testing.T, vm *goja.Runtime, host *enginetest.Host)
	}{
		{
			name: "empty url",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('mokapi/http')
					m.get('')
				`)
				r.EqualError(t, err, "url cannot be empty at mokapi/js/http.(*Module).Get-fm (native)")
			},
		},
		{
			name: "http client error",
			client: &enginetest.HttpClient{
				DoFunc: func(request *http.Request) (*http.Response, error) {
					return nil, fmt.Errorf("TEST")
				},
			},
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('mokapi/http')
					m.get('https://foo.bar')
				`)
				r.EqualError(t, err, "TEST at mokapi/js/http.(*Module).Get-fm (native)")
			},
		},
		{
			name: "request uses given url",
			client: &enginetest.HttpClient{
				DoFunc: func(request *http.Request) (*http.Response, error) {
					if request.URL.String() == "https://foo.bar" {
						return &http.Response{}, nil
					}
					return nil, fmt.Errorf("TEST")
				},
			},
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('mokapi/http')
					m.get('https://foo.bar')
				`)
				r.NoError(t, err)
			},
		},
		{
			name: "HTTP status code",
			client: &enginetest.HttpClient{
				DoFunc: func(request *http.Request) (*http.Response, error) {
					return &http.Response{StatusCode: http.StatusOK}, nil
				},
			},
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('mokapi/http')
					m.get('https://foo.bar').statusCode
				`)
				r.NoError(t, err)
				r.Equal(t, int64(http.StatusOK), v.Export())
			},
		},
		{
			name: "HTTP status code",
			client: &enginetest.HttpClient{
				DoFunc: func(request *http.Request) (*http.Response, error) {
					return &http.Response{StatusCode: http.StatusOK}, nil
				},
			},
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('mokapi/http')
					m.get('https://foo.bar').statusCode
				`)
				r.NoError(t, err)
				r.Equal(t, int64(http.StatusOK), v.Export())
			},
		},
		{
			name: "HTTP header",
			client: &enginetest.HttpClient{
				DoFunc: func(request *http.Request) (*http.Response, error) {
					return &http.Response{Header: map[string][]string{"foo": {"bar"}}}, nil
				},
			},
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('mokapi/http')
					m.get('https://foo.bar').headers.foo
				`)
				r.NoError(t, err)
				r.Equal(t, []string{"bar"}, v.Export())
			},
		},
		{
			name: "HTTP body",
			client: &enginetest.HttpClient{
				DoFunc: func(request *http.Request) (*http.Response, error) {
					return &http.Response{Body: io.NopCloser(strings.NewReader("foobar"))}, nil
				},
			},
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('mokapi/http')
					m.get('https://foo.bar').body
				`)
				r.NoError(t, err)
				r.Equal(t, "foobar", v.Export())
			},
		},
		{
			name: "HTTP body to json",
			client: &enginetest.HttpClient{
				DoFunc: func(request *http.Request) (*http.Response, error) {
					return &http.Response{Body: io.NopCloser(strings.NewReader(`{"foo":"bar"}`))}, nil
				},
			},
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('mokapi/http')
					m.get('https://foo.bar').json()
				`)
				r.NoError(t, err)
				r.Equal(t, map[string]interface{}{"foo": "bar"}, v.Export())
			},
		},
		{
			name: "HTTP body to json but invalid format",
			client: &enginetest.HttpClient{
				DoFunc: func(request *http.Request) (*http.Response, error) {
					return &http.Response{Body: io.NopCloser(strings.NewReader(`{"foo":"bar"`))}, nil
				},
			},
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('mokapi/http')
					m.get('https://foo.bar').json()
				`)
				r.EqualError(t, err, "response is not a valid JSON response: unexpected end of JSON input at reflect.methodValueCall (native)")
			},
		},
		{
			name: "HTTP post, convert object to json",
			client: &enginetest.HttpClient{
				DoFunc: func(request *http.Request) (*http.Response, error) {
					if request.Method != http.MethodPost {
						return nil, fmt.Errorf("expected HTTP method POST, but is %v", request.Method)
					}
					if s := request.Header["Content-Type"][0]; s != "application/json" {
						return nil, fmt.Errorf("expected Content-Type application/json, but is %v", s)
					}
					if b, _ := io.ReadAll(request.Body); string(b) != `{"foo":"bar"}` {
						return nil, fmt.Errorf("expected request body , but is %s", b)
					}
					return &http.Response{}, nil
				},
			},
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('mokapi/http')
					m.post('https://foo.bar', { 'foo': 'bar' }, {
						headers: { 'Content-Type': 'application/json' }
					})
				`)
				r.NoError(t, err)
			},
		},
		{
			name: "HTTP post, unsupported content type",
			client: &enginetest.HttpClient{
				DoFunc: func(request *http.Request) (*http.Response, error) {
					if request.Method != http.MethodPost {
						return nil, fmt.Errorf("expected HTTP method POST, but is %v", request.Method)
					}
					return &http.Response{}, nil
				},
			},
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('mokapi/http')
					m.post('https://foo.bar', { 'foo': 'bar' }, {
						headers: { 'Content-Type': 'foo/bar' }
					})
				`)
				r.EqualError(t, err, "encoding request body failed: content type 'foo/bar' is not supported at mokapi/js/http.(*Module).Post-fm (native)")
			},
		},
		{
			name: "HTTP put",
			client: &enginetest.HttpClient{
				DoFunc: func(request *http.Request) (*http.Response, error) {
					if request.Method != http.MethodPut {
						return nil, fmt.Errorf("expected HTTP method PUT, but is %v", request.Method)
					}
					return &http.Response{}, nil
				},
			},
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('mokapi/http')
					m.put('https://foo.bar')
				`)
				r.NoError(t, err)
			},
		},
		{
			name: "HTTP head",
			client: &enginetest.HttpClient{
				DoFunc: func(request *http.Request) (*http.Response, error) {
					if request.Method != http.MethodHead {
						return nil, fmt.Errorf("expected HTTP method HEAD, but is %v", request.Method)
					}
					return &http.Response{}, nil
				},
			},
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('mokapi/http')
					m.head('https://foo.bar')
				`)
				r.NoError(t, err)
			},
		},
		{
			name: "HTTP patch",
			client: &enginetest.HttpClient{
				DoFunc: func(request *http.Request) (*http.Response, error) {
					if request.Method != http.MethodPatch {
						return nil, fmt.Errorf("expected HTTP method PATCH, but is %v", request.Method)
					}
					return &http.Response{}, nil
				},
			},
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('mokapi/http')
					m.patch('https://foo.bar')
				`)
				r.NoError(t, err)
			},
		},
		{
			name: "HTTP delete",
			client: &enginetest.HttpClient{
				DoFunc: func(request *http.Request) (*http.Response, error) {
					if request.Method != http.MethodDelete {
						return nil, fmt.Errorf("expected HTTP method DELETE, but is %v", request.Method)
					}
					return &http.Response{}, nil
				},
			},
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('mokapi/http')
					m.delete('https://foo.bar')
				`)
				r.NoError(t, err)
			},
		},
		{
			name: "HTTP options",
			client: &enginetest.HttpClient{
				DoFunc: func(request *http.Request) (*http.Response, error) {
					if request.Method != http.MethodOptions {
						return nil, fmt.Errorf("expected HTTP method OPTIONS, but is %v", request.Method)
					}
					return &http.Response{}, nil
				},
			},
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('mokapi/http')
					m.options('https://foo.bar')
				`)
				r.NoError(t, err)
			},
		},
		{
			name: "fetch get request",
			client: &enginetest.HttpClient{
				DoFunc: func(request *http.Request) (*http.Response, error) {
					if request.Method != http.MethodGet {
						return nil, fmt.Errorf("expected HTTP method GET, but is %v", request.Method)
					}
					return &http.Response{StatusCode: http.StatusOK}, nil
				},
			},
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('mokapi/http')
					const p = m.fetch('https://foo.bar')
					let result;
					p.then(v => result = v).catch(err => result = err)
				`)
				r.NoError(t, err)
				time.Sleep(200 * time.Millisecond)

				v, err := vm.RunString("result")
				r.NoError(t, err)
				res, ok := v.Export().(mod.Response)
				if !ok {
					r.FailNow(t, v.String())
				}
				r.Equal(t, http.StatusOK, res.StatusCode)
			},
		},
		{
			name: "fetch post request",
			client: &enginetest.HttpClient{
				DoFunc: func(request *http.Request) (*http.Response, error) {
					if request.Method != http.MethodPost {
						return nil, fmt.Errorf("expected HTTP method POST, but is %v", request.Method)
					}
					return &http.Response{StatusCode: http.StatusOK}, nil
				},
			},
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('mokapi/http')
					const p = m.fetch('https://foo.bar', { method: 'post' })
					let result;
					p.then(v => result = v).catch(err => result = err)
				`)
				r.NoError(t, err)
				time.Sleep(200 * time.Millisecond)

				v, err := vm.RunString("result")
				r.NoError(t, err)
				res, ok := v.Export().(mod.Response)
				if !ok {
					r.FailNow(t, v.String())
				}
				r.Equal(t, http.StatusOK, res.StatusCode)
			},
		},
		{
			name: "fetch put with body",
			client: &enginetest.HttpClient{
				DoFunc: func(request *http.Request) (*http.Response, error) {
					if request.Method != http.MethodPut {
						return nil, fmt.Errorf("expected HTTP method PUT, but is %v", request.Method)
					}
					b, err := io.ReadAll(request.Body)
					if err != nil {
						return nil, fmt.Errorf("cannot read body: %w", err)
					} else if string(b) != `{"foo":"bar"}` {
						return nil, fmt.Errorf("expected body to be '{\"foo\":\"bar\"}', but is %s", b)
					}
					return &http.Response{StatusCode: http.StatusOK}, nil
				},
			},
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('mokapi/http')
					const p = m.fetch('https://foo.bar', { method: 'PUT', body: { foo: 'bar' } })
					let result;
					p.then(v => result = v).catch(err => result = err)
				`)
				r.NoError(t, err)
				time.Sleep(200 * time.Millisecond)

				v, err := vm.RunString("result")
				r.NoError(t, err)
				res, ok := v.Export().(mod.Response)
				if !ok {
					r.FailNow(t, v.String())
				}
				r.Equal(t, http.StatusOK, res.StatusCode)
			},
		},
		{
			name: "fetch delete with header",
			client: &enginetest.HttpClient{
				DoFunc: func(request *http.Request) (*http.Response, error) {
					if request.Method != http.MethodDelete {
						return nil, fmt.Errorf("expected HTTP method DELETE, but is %v", request.Method)
					}
					if request.Header["foo"][0] != "bar" {
						return nil, fmt.Errorf("expected header foo to contain 'bar', but is %v", request.Header["foo"])
					}
					if request.Header["bar"][0] != "f" || request.Header["bar"][1] != "o" || request.Header["bar"][2] != "o" {
						return nil, fmt.Errorf("expected header foo to be [f o o], but is %v", request.Header["bar"])
					}
					return &http.Response{StatusCode: http.StatusOK}, nil
				},
			},
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('mokapi/http')
					const p = m.fetch('https://foo.bar', { method: 'delete', headers: { foo: 'bar', bar: [ 'f', 'o', 'o' ] } })
					let result;
					p.then(v => result = v).catch(err => result = err)
				`)
				r.NoError(t, err)
				time.Sleep(200 * time.Millisecond)

				v, err := vm.RunString("result")
				r.NoError(t, err)
				res, ok := v.Export().(mod.Response)
				if !ok {
					r.FailNow(t, v.String())
				}
				r.Equal(t, http.StatusOK, res.StatusCode)
			},
		},
		{
			name: "fetch error",
			client: &enginetest.HttpClient{
				DoFunc: func(request *http.Request) (*http.Response, error) {
					return nil, fmt.Errorf("TEST ERROR")
				},
			},
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('mokapi/http')
					const p = m.fetch('https://foo.bar', { method: 'delete', headers: { foo: 'bar', bar: [ 'f', 'o', 'o' ] } })
					let result;
					p.then(v => result = v).catch(err => result = err)
				`)
				r.NoError(t, err)
				time.Sleep(200 * time.Millisecond)

				v, err := vm.RunString("result")
				r.NoError(t, err)
				r.Equal(t, "TEST ERROR", v.Export())
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
			req.RegisterNativeModule("mokapi/http", mod.Require)

			tc.test(t, vm, host)
		})
	}
}
