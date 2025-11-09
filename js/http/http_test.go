package http_test

import (
	"fmt"
	"io"
	"mokapi/config/dynamic"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/eventloop"
	mod "mokapi/js/http"
	"mokapi/js/require"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
)

type timeoutErr struct{}

func (timeoutErr) Timeout() bool { return true }
func (timeoutErr) Error() string { return "timeout" }

func TestHttp(t *testing.T) {
	testcases := []struct {
		name   string
		client func(options common.HttpClientOptions) common.HttpClient
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
			client: func(options common.HttpClientOptions) common.HttpClient {
				return &enginetest.HttpClient{
					DoFunc: func(request *http.Request) (*http.Response, error) {
						return nil, fmt.Errorf("TEST")
					},
				}
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
			client: func(options common.HttpClientOptions) common.HttpClient {
				return &enginetest.HttpClient{
					DoFunc: func(request *http.Request) (*http.Response, error) {
						if request.URL.String() == "https://foo.bar" {
							return &http.Response{}, nil
						}
						return nil, fmt.Errorf("TEST")
					},
				}
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
			client: func(options common.HttpClientOptions) common.HttpClient {
				return &enginetest.HttpClient{
					DoFunc: func(request *http.Request) (*http.Response, error) {
						return &http.Response{StatusCode: http.StatusOK}, nil
					},
				}
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
			client: func(options common.HttpClientOptions) common.HttpClient {
				return &enginetest.HttpClient{
					DoFunc: func(request *http.Request) (*http.Response, error) {
						return &http.Response{StatusCode: http.StatusOK}, nil
					},
				}
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
			client: func(options common.HttpClientOptions) common.HttpClient {
				return &enginetest.HttpClient{
					DoFunc: func(request *http.Request) (*http.Response, error) {
						return &http.Response{Header: map[string][]string{"foo": {"bar"}}}, nil
					},
				}
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
			client: func(options common.HttpClientOptions) common.HttpClient {
				return &enginetest.HttpClient{
					DoFunc: func(request *http.Request) (*http.Response, error) {
						return &http.Response{Body: io.NopCloser(strings.NewReader("foobar"))}, nil
					},
				}
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
			client: func(options common.HttpClientOptions) common.HttpClient {
				return &enginetest.HttpClient{
					DoFunc: func(request *http.Request) (*http.Response, error) {
						return &http.Response{Body: io.NopCloser(strings.NewReader(`{"foo":"bar"}`))}, nil
					},
				}
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
			client: func(options common.HttpClientOptions) common.HttpClient {
				return &enginetest.HttpClient{
					DoFunc: func(request *http.Request) (*http.Response, error) {
						return &http.Response{Body: io.NopCloser(strings.NewReader(`{"foo":"bar"`))}, nil
					},
				}
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
			client: func(options common.HttpClientOptions) common.HttpClient {
				return &enginetest.HttpClient{
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
				}
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
			client: func(options common.HttpClientOptions) common.HttpClient {
				return &enginetest.HttpClient{
					DoFunc: func(request *http.Request) (*http.Response, error) {
						if request.Method != http.MethodPost {
							return nil, fmt.Errorf("expected HTTP method POST, but is %v", request.Method)
						}
						return &http.Response{}, nil
					},
				}
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
			client: func(options common.HttpClientOptions) common.HttpClient {
				return &enginetest.HttpClient{
					DoFunc: func(request *http.Request) (*http.Response, error) {
						if request.Method != http.MethodPut {
							return nil, fmt.Errorf("expected HTTP method PUT, but is %v", request.Method)
						}
						return &http.Response{}, nil
					},
				}
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
			client: func(options common.HttpClientOptions) common.HttpClient {
				return &enginetest.HttpClient{
					DoFunc: func(request *http.Request) (*http.Response, error) {
						if request.Method != http.MethodHead {
							return nil, fmt.Errorf("expected HTTP method HEAD, but is %v", request.Method)
						}
						return &http.Response{}, nil
					},
				}
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
			client: func(options common.HttpClientOptions) common.HttpClient {
				return &enginetest.HttpClient{
					DoFunc: func(request *http.Request) (*http.Response, error) {
						if request.Method != http.MethodPatch {
							return nil, fmt.Errorf("expected HTTP method PATCH, but is %v", request.Method)
						}
						return &http.Response{}, nil
					},
				}
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
			client: func(options common.HttpClientOptions) common.HttpClient {
				return &enginetest.HttpClient{
					DoFunc: func(request *http.Request) (*http.Response, error) {
						if request.Method != http.MethodDelete {
							return nil, fmt.Errorf("expected HTTP method DELETE, but is %v", request.Method)
						}
						return &http.Response{}, nil
					},
				}
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
			client: func(options common.HttpClientOptions) common.HttpClient {
				return &enginetest.HttpClient{
					DoFunc: func(request *http.Request) (*http.Response, error) {
						if request.Method != http.MethodOptions {
							return nil, fmt.Errorf("expected HTTP method OPTIONS, but is %v", request.Method)
						}
						return &http.Response{}, nil
					},
				}
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
			name: "timeout",
			client: func(options common.HttpClientOptions) common.HttpClient {
				r.Equal(t, 500*time.Millisecond, options.Timeout)
				return &enginetest.HttpClient{
					DoFunc: func(request *http.Request) (*http.Response, error) {
						return nil, &timeoutErr{}
					},
				}
			},
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('mokapi/http')
					m.get('https://foo.bar', { timeout: '500ms' })
				`)
				r.EqualError(t, err, "request to GET https://foo.bar timed out at mokapi/js/http.(*Module).Get-fm (native)")
			},
		},
		{
			name: "fetch get request",
			client: func(options common.HttpClientOptions) common.HttpClient {
				return &enginetest.HttpClient{
					DoFunc: func(request *http.Request) (*http.Response, error) {
						if request.Method != http.MethodGet {
							return nil, fmt.Errorf("expected HTTP method GET, but is %v", request.Method)
						}
						return &http.Response{StatusCode: http.StatusOK}, nil
					},
				}
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
			client: func(options common.HttpClientOptions) common.HttpClient {
				return &enginetest.HttpClient{
					DoFunc: func(request *http.Request) (*http.Response, error) {
						if request.Method != http.MethodPost {
							return nil, fmt.Errorf("expected HTTP method POST, but is %v", request.Method)
						}
						return &http.Response{StatusCode: http.StatusOK}, nil
					},
				}
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
			client: func(options common.HttpClientOptions) common.HttpClient {
				return &enginetest.HttpClient{
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
				}
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
			client: func(options common.HttpClientOptions) common.HttpClient {
				return &enginetest.HttpClient{
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
				}
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
			client: func(options common.HttpClientOptions) common.HttpClient {
				return &enginetest.HttpClient{
					DoFunc: func(request *http.Request) (*http.Response, error) {
						return nil, fmt.Errorf("TEST ERROR")
					},
				}
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
		{
			name: "fetch method not string",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('mokapi/http')
					const p = m.fetch('https://foo.bar', { method: 12 })
					let result;
					p.then(v => result = v).catch(err => result = err)
				`)
				r.NoError(t, err)
				time.Sleep(200 * time.Millisecond)

				v, err := vm.RunString("result")
				r.NoError(t, err)
				r.Equal(t, "unexpected type for 'method': got Integer, expected String", v.Export())
			},
		},
		{
			name: "fetch maxRedirects not number",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				_, err := vm.RunString(`
					const m = require('mokapi/http')
					const p = m.fetch('https://foo.bar', { maxRedirects: 'foo' })
					let result;
					p.then(v => result = v).catch(err => result = err)
				`)
				r.NoError(t, err)
				time.Sleep(200 * time.Millisecond)

				v, err := vm.RunString("result")
				r.NoError(t, err)
				r.Equal(t, "unexpected type for 'maxRedirects': got String, expected Number", v.Export())
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			vm := goja.New()
			vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
			host := &enginetest.Host{HttpClientFunc: tc.client}
			js.EnableInternal(vm, host, &eventloop.EventLoop{}, &dynamic.Config{})
			req, err := require.NewRegistry()
			r.NoError(t, err)
			req.Enable(vm)
			req.RegisterNativeModule("mokapi/http", mod.Require)

			tc.test(t, vm, host)
		})
	}
}
