package js_test

import (
	"fmt"
	"github.com/sirupsen/logrus/hooks/test"
	r "github.com/stretchr/testify/require"
	"io"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/script"
	"mokapi/config/static"
	"mokapi/engine"
	"mokapi/engine/enginetest"
	"mokapi/js"
	module "mokapi/js/http"
	"mokapi/js/jstest"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestScript_Http_Get(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, host *enginetest.Host)
	}{
		{
			name: "simple",
			test: func(t *testing.T, host *enginetest.Host) {
				s, err := jstest.New(jstest.WithSource(
					`import http from 'mokapi/http'
						 export default function() {
						 	return http.get('https://foo.bar')
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
				r.Equal(t, "GET", host.HttpClientTest.LastRequest.Method)
				r.Equal(t, "https://foo.bar", host.HttpClientTest.LastRequest.URL.String())
			},
		},
		{
			name: "header",
			test: func(t *testing.T, host *enginetest.Host) {
				s, err := jstest.New(jstest.WithSource(
					`import http from 'mokapi/http'
						 export default function() {
						  	return http.get('https://foo.bar', {headers: {foo: "bar"}})
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
				r.Equal(t, []string{"bar"}, host.HttpClientTest.LastRequest.Header["foo"])
			},
		},
		{
			name: "header with array",
			test: func(t *testing.T, host *enginetest.Host) {
				s, err := jstest.New(jstest.WithSource(
					`import http from 'mokapi/http'
						 export default function() {
						  	return http.get('https://foo.bar', {headers: {foo: ["hello", "world"]}})
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
				r.Equal(t, []string{"hello", "world"}, host.HttpClientTest.LastRequest.Header["foo"])
			},
		},
		{
			name: "header set to null",
			test: func(t *testing.T, host *enginetest.Host) {
				s, err := jstest.New(jstest.WithSource(
					`import http from 'mokapi/http'
						 export default function() {
						  	return http.get('https://foo.bar', {headers: null})
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
				r.Len(t, host.HttpClientTest.LastRequest.Header, 0)
			},
		},
		{
			name: "invalid url",
			test: func(t *testing.T, host *enginetest.Host) {
				s, err := jstest.New(jstest.WithSource(
					`import http from 'mokapi/http'
						 export default function() {
						  	return http.get('://')
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				err = s.Run()
				r.Error(t, err)
			},
		},
		{
			name: "response body",
			test: func(t *testing.T, host *enginetest.Host) {
				host.HttpClientTest.DoFunc = func(request *http.Request) (*http.Response, error) {
					return &http.Response{Body: io.NopCloser(strings.NewReader("hello world"))}, nil
				}
				s, err := jstest.New(jstest.WithSource(
					`import http from 'mokapi/http'
						 export default function() {
						  	return http.get('https://foo.bar')
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				v, err := s.RunDefault()
				r.NoError(t, err)
				result := v.Export().(module.Response)
				r.Equal(t, "hello world", result.Body)
			},
		},
		{
			name: "response body json",
			test: func(t *testing.T, host *enginetest.Host) {
				host.HttpClientTest.DoFunc = func(request *http.Request) (*http.Response, error) {
					return &http.Response{Body: io.NopCloser(strings.NewReader(`{"foo": "bar"}`))}, nil
				}
				s, err := jstest.New(jstest.WithSource(
					`import http from 'mokapi/http'
						 export default function() {
						  	return http.get('https://foo.bar').json()
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				v, err := s.RunDefault()
				r.NoError(t, err)
				result := v.Export()
				r.Equal(t, map[string]interface{}{"foo": "bar"}, result)
			},
		},
		{
			name: "response header",
			test: func(t *testing.T, host *enginetest.Host) {
				host.HttpClientTest.DoFunc = func(request *http.Request) (*http.Response, error) {
					return &http.Response{Header: map[string][]string{"Allow": {"OPTIONS", "GET", "HEAD", "POST"}}}, nil
				}
				s, err := jstest.New(jstest.WithSource(
					`import http from 'mokapi/http'
						 export default function() {
						  	const res = http.options('https://foo.bar')
							return res.headers['Allow']
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				v, err := s.RunDefault()
				r.NoError(t, err)
				result := v.Export()
				r.Equal(t, []string{"OPTIONS", "GET", "HEAD", "POST"}, result)
			},
		},
		{
			name: "client error",
			test: func(t *testing.T, host *enginetest.Host) {
				host.HttpClientTest.DoFunc = func(request *http.Request) (*http.Response, error) {
					return nil, fmt.Errorf("test error")
				}
				s, err := jstest.New(jstest.WithSource(
					`import http from 'mokapi/http'
						 export default function() {
						  	return http.get('https://foo.bar')
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				err = s.Run()
				r.Error(t, err)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			host := &enginetest.Host{
				HttpClientTest: &enginetest.HttpClient{},
			}

			tc.test(t, host)
		})
	}
}

func TestScript_Http_Post(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, host *enginetest.Host)
	}{
		{
			name: "simple",
			test: func(t *testing.T, host *enginetest.Host) {
				s, err := jstest.New(jstest.WithSource(
					`import http from 'mokapi/http'
						 export default function() {
  							return http.post('https://foo.bar')
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
				r.Equal(t, "POST", host.HttpClientTest.LastRequest.Method)
				r.Equal(t, "https://foo.bar", host.HttpClientTest.LastRequest.URL.String())
			},
		},
		{
			name: "json with body as string",
			test: func(t *testing.T, host *enginetest.Host) {
				s, err := jstest.New(jstest.WithSource(
					`import http from 'mokapi/http'
						 export default function() {
						  	return http.post("http://localhost/foo", "body", {headers: {'Content-Type': "application/json"}})
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
				b, err := io.ReadAll(host.HttpClientTest.LastRequest.Body)
				r.Equal(t, "body", string(b))
				r.Equal(t, "application/json", host.HttpClientTest.LastRequest.Header.Get("Content-Type"))
			},
		},
		{
			name: "json with body as object",
			test: func(t *testing.T, host *enginetest.Host) {
				s, err := jstest.New(jstest.WithSource(
					`import http from 'mokapi/http'
						 export default function() {
						  	return http.post("http://localhost/foo", {"foo":"bar"}, {headers: {'Content-Type': "application/json"}})
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
				b, err := io.ReadAll(host.HttpClientTest.LastRequest.Body)
				r.Equal(t, `{"foo":"bar"}`, string(b))
				r.Equal(t, "application/json", host.HttpClientTest.LastRequest.Header.Get("Content-Type"))
			},
		},
		{
			name: "json with body as object without Content-Type",
			test: func(t *testing.T, host *enginetest.Host) {
				s, err := jstest.New(jstest.WithSource(
					`import http from 'mokapi/http'
						 export default function() {
						  	return http.post("http://localhost/foo", {"foo":"bar"})
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
				b, err := io.ReadAll(host.HttpClientTest.LastRequest.Body)
				r.Equal(t, `{"foo":"bar"}`, string(b))
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			host := &enginetest.Host{
				HttpClientTest: &enginetest.HttpClient{},
			}

			tc.test(t, host)
		})
	}
}

func TestScript_Http(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, host *enginetest.Host)
	}{
		{
			name: "GET",
			test: func(t *testing.T, host *enginetest.Host) {
				r.Equal(t, "GET", host.HttpClientTest.LastRequest.Method)
			},
		},
		{
			name: "POST",
			test: func(t *testing.T, host *enginetest.Host) {
				r.Equal(t, "POST", host.HttpClientTest.LastRequest.Method)
			},
		},
		{
			name: "PUT",
			test: func(t *testing.T, host *enginetest.Host) {
				r.Equal(t, "PUT", host.HttpClientTest.LastRequest.Method)
			},
		},
		{
			name: "HEAD",
			test: func(t *testing.T, host *enginetest.Host) {
				r.Equal(t, "HEAD", host.HttpClientTest.LastRequest.Method)
			},
		},
		{
			name: "PATCH",
			test: func(t *testing.T, host *enginetest.Host) {
				r.Equal(t, "PATCH", host.HttpClientTest.LastRequest.Method)
			},
		},
		{
			name: "DEL",
			test: func(t *testing.T, host *enginetest.Host) {
				r.Equal(t, "DELETE", host.HttpClientTest.LastRequest.Method)
			},
		},
		{
			name: "DELETE",
			test: func(t *testing.T, host *enginetest.Host) {
				r.Equal(t, "DELETE", host.HttpClientTest.LastRequest.Method)
			},
		},
		{
			name: "OPTIONS",
			test: func(t *testing.T, host *enginetest.Host) {
				r.Equal(t, "OPTIONS", host.HttpClientTest.LastRequest.Method)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			host := &enginetest.Host{
				HttpClientTest: &enginetest.HttpClient{},
			}

			s, err := jstest.New(jstest.WithSource(
				fmt.Sprintf(`import http from 'mokapi/http'
						 export default function() {
						 	return http.%v('https://foo.bar')
						 }`, strings.ToLower(tc.name))),
				js.WithHost(host))
			r.NoError(t, err)
			err = s.Run()
			r.NoError(t, err)

			tc.test(t, host)
		})
	}
}

func TestFetch(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, host *enginetest.Host)
	}{
		{
			name: "simple",
			test: func(t *testing.T, host *enginetest.Host) {
				host.HttpClientTest.DoFunc = func(request *http.Request) (*http.Response, error) {
					r.Equal(t, "GET", request.Method)
					r.Equal(t, "https://foo.bar", request.URL.String())

					return &http.Response{StatusCode: http.StatusOK}, nil
				}

				s, err := jstest.New(jstest.WithSource(
					`import http from 'mokapi/http'
						 export default async function() {
						 	const res = await http.fetch('https://foo.bar')
							return { status: res.statusCode }
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				v, err := s.RunDefault()
				r.NoError(t, err)
				res := v.Export().(map[string]interface{})
				r.Equal(t, int64(http.StatusOK), res["status"])
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			host := &enginetest.Host{
				HttpClientTest: &enginetest.HttpClient{},
			}
			tc.test(t, host)
		})
	}
}

func TestMaxRedirects(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		n := 0
		if len(request.URL.Path) > 1 {
			var err error
			n, err = strconv.Atoi(request.URL.Path[1:])
			if err != nil {
				panic(err)
			}
		}

		writer.Header().Set("Location", fmt.Sprintf("%s/%d", server.URL, n+1))
		writer.WriteHeader(302)
	}))

	testcases := []struct {
		name string
		code string
	}{{
		name: "max redirects 0",
		code: fmt.Sprintf(`import http from 'mokapi/http'
export default function() {
	const res = http.get('%s', { maxRedirects: 0 });
	console.log(res.headers.Location[0]); 
}
`, server.URL),
	}}

	hook := test.NewGlobal()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			e := engine.NewEngine(engine.WithScriptLoader(engine.NewDefaultScriptLoader(&static.Config{})))
			err := e.AddScript(&dynamic.Config{
				Info: dynamic.ConfigInfo{
					Url: mustParse("test.ts"),
				},
				Raw:  []byte(tc.code),
				Data: &script.Script{Filename: "test.ts"},
			})
			r.NoError(t, err)
			r.Len(t, hook.Entries, 2)
			r.Equal(t, fmt.Sprintf("Stopped after 5 redirects, original URL was %s", server.URL), hook.Entries[0].Message)
			r.Equal(t, fmt.Sprintf("%s/6", server.URL), hook.Entries[1].Message)

			hook.Reset()

		})
	}
}
