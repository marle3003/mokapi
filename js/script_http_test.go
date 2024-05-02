package js

import (
	"fmt"
	r "github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"mokapi/config/static"
	"net/http"
	"strings"
	"testing"
)

func TestScript_Http_Get(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, host *testHost)
	}{
		{
			name: "simple",
			test: func(t *testing.T, host *testHost) {
				s, err := New(newScript("",
					`import http from 'mokapi/http'
						 export default function() {
						 	return http.get('http://foo.bar')
						 }`),
					host, static.JsConfig{})
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, "GET", host.httpClient.req.Method)
				r.Equal(t, "http://foo.bar", host.httpClient.req.URL.String())
			},
		},
		{
			name: "header",
			test: func(t *testing.T, host *testHost) {
				s, err := New(newScript("",
					`import http from 'mokapi/http'
						 export default function() {
						  	return http.get('http://foo.bar', {headers: {foo: "bar"}})
						 }`),
					host, static.JsConfig{})
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, "bar", host.httpClient.req.Header.Get("foo"))
			},
		},
		{
			name: "header with array",
			test: func(t *testing.T, host *testHost) {
				s, err := New(newScript("",
					`import http from 'mokapi/http'
						 export default function() {
						  	return http.get('http://foo.bar', {headers: {foo: ["hello", "world"]}})
						 }`),
					host, static.JsConfig{})
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, []string{"hello", "world"}, host.httpClient.req.Header.Values("foo"))
			},
		},
		{
			name: "header set to null",
			test: func(t *testing.T, host *testHost) {
				s, err := New(newScript("",
					`import http from 'mokapi/http'
						 export default function() {
						  	return http.get('http://foo.bar', {headers: null})
						 }`),
					host, static.JsConfig{})
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.NoError(t, err)
				r.Len(t, host.httpClient.req.Header, 0)
			},
		},
		{
			name: "invalid url",
			test: func(t *testing.T, host *testHost) {
				s, err := New(newScript("",
					`import http from 'mokapi/http'
						 export default function() {
						  	return http.get('://')
						 }`),
					host, static.JsConfig{})
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.Error(t, err)
			},
		},
		{
			name: "response body",
			test: func(t *testing.T, host *testHost) {
				host.httpClient.doFunc = func(request *http.Request) (*http.Response, error) {
					return &http.Response{Body: io.NopCloser(strings.NewReader("hello world"))}, nil
				}
				s, err := New(newScript("",
					`import http from 'mokapi/http'
						 export default function() {
						  	return http.get('http://foo.bar')
						 }`),
					host, static.JsConfig{})
				r.NoError(t, err)
				v, err := s.RunDefault()
				r.NoError(t, err)
				result := v.Export().(response)
				r.Equal(t, "hello world", result.Body)
			},
		},
		{
			name: "response body json",
			test: func(t *testing.T, host *testHost) {
				host.httpClient.doFunc = func(request *http.Request) (*http.Response, error) {
					return &http.Response{Body: io.NopCloser(strings.NewReader(`{"foo": "bar"}`))}, nil
				}
				s, err := New(newScript("",
					`import http from 'mokapi/http'
						 export default function() {
						  	return http.get('http://foo.bar').json()
						 }`),
					host, static.JsConfig{})
				r.NoError(t, err)
				v, err := s.RunDefault()
				r.NoError(t, err)
				result := v.Export()
				r.Equal(t, map[string]interface{}{"foo": "bar"}, result)
			},
		},
		{
			name: "response header",
			test: func(t *testing.T, host *testHost) {
				host.httpClient.doFunc = func(request *http.Request) (*http.Response, error) {
					return &http.Response{Header: map[string][]string{"Allow": {"OPTIONS", "GET", "HEAD", "POST"}}}, nil
				}
				s, err := New(newScript("",
					`import http from 'mokapi/http'
						 export default function() {
						  	const res = http.options('https://foo.bar')
							return res.headers['Allow']
						 }`),
					host, static.JsConfig{})
				r.NoError(t, err)
				v, err := s.RunDefault()
				r.NoError(t, err)
				result := v.Export()
				r.Equal(t, []string{"OPTIONS", "GET", "HEAD", "POST"}, result)
			},
		},
		{
			name: "client error",
			test: func(t *testing.T, host *testHost) {
				host.httpClient.doFunc = func(request *http.Request) (*http.Response, error) {
					return nil, fmt.Errorf("test error")
				}
				s, err := New(newScript("",
					`import http from 'mokapi/http'
						 export default function() {
						  	return http.get('http://foo.bar')
						 }`),
					host, static.JsConfig{})
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.Error(t, err)
			},
		},
		{
			name: "using deprecated module",
			test: func(t *testing.T, host *testHost) {
				s, err := New(newScript("",
					`import http from 'http'
						 export default function() {
						 	return http.get('http://foo.bar')
						 }`),
					host, static.JsConfig{})
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, "GET", host.httpClient.req.Method)
				r.Equal(t, "http://foo.bar", host.httpClient.req.URL.String())
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			host := &testHost{
				httpClient: &testClient{},
			}

			tc.test(t, host)
		})
	}
}

func TestScript_Http_Post(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, host *testHost)
	}{
		{
			name: "simple",
			test: func(t *testing.T, host *testHost) {
				s, err := New(newScript("",
					`import http from 'mokapi/http'
						 export default function() {
  							return http.post('http://foo.bar')
						 }`),
					host, static.JsConfig{})
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.NoError(t, err)
				r.Equal(t, "POST", host.httpClient.req.Method)
				r.Equal(t, "http://foo.bar", host.httpClient.req.URL.String())
			},
		},
		{
			name: "json with body as string",
			test: func(t *testing.T, host *testHost) {
				s, err := New(newScript("",
					`import http from 'mokapi/http'
						 export default function() {
						  	return http.post("http://localhost/foo", "body", {headers: {'Content-Type': "application/json"}})
						 }`),
					host, static.JsConfig{})
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.NoError(t, err)
				b, err := ioutil.ReadAll(host.httpClient.req.Body)
				r.Equal(t, "body", string(b))
				r.Equal(t, "application/json", host.httpClient.req.Header.Get("Content-Type"))
			},
		},
		{
			name: "json with body as object",
			test: func(t *testing.T, host *testHost) {
				s, err := New(newScript("",
					`import http from 'mokapi/http'
						 export default function() {
						  	return http.post("http://localhost/foo", {"foo":"bar"}, {headers: {'Content-Type': "application/json"}})
						 }`),
					host, static.JsConfig{})
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.NoError(t, err)
				b, err := io.ReadAll(host.httpClient.req.Body)
				r.Equal(t, `{"foo":"bar"}`, string(b))
				r.Equal(t, "application/json", host.httpClient.req.Header.Get("Content-Type"))
			},
		},
		{
			name: "json with body as object without Content-Type",
			test: func(t *testing.T, host *testHost) {
				s, err := New(newScript("",
					`import http from 'mokapi/http'
						 export default function() {
						  	return http.post("http://localhost/foo", {"foo":"bar"})
						 }`),
					host, static.JsConfig{})
				r.NoError(t, err)
				_, err = s.RunDefault()
				r.NoError(t, err)
				b, err := io.ReadAll(host.httpClient.req.Body)
				r.Equal(t, `{"foo":"bar"}`, string(b))
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			host := &testHost{
				httpClient: &testClient{},
			}

			tc.test(t, host)
		})
	}
}

func TestScript_Http(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, host *testHost)
	}{
		{
			name: "GET",
			test: func(t *testing.T, host *testHost) {
				r.Equal(t, "GET", host.httpClient.req.Method)
			},
		},
		{
			name: "POST",
			test: func(t *testing.T, host *testHost) {
				r.Equal(t, "POST", host.httpClient.req.Method)
			},
		},
		{
			name: "PUT",
			test: func(t *testing.T, host *testHost) {
				r.Equal(t, "PUT", host.httpClient.req.Method)
			},
		},
		{
			name: "HEAD",
			test: func(t *testing.T, host *testHost) {
				r.Equal(t, "HEAD", host.httpClient.req.Method)
			},
		},
		{
			name: "PATCH",
			test: func(t *testing.T, host *testHost) {
				r.Equal(t, "PATCH", host.httpClient.req.Method)
			},
		},
		{
			name: "DEL",
			test: func(t *testing.T, host *testHost) {
				r.Equal(t, "DELETE", host.httpClient.req.Method)
			},
		},
		{
			name: "DELETE",
			test: func(t *testing.T, host *testHost) {
				r.Equal(t, "DELETE", host.httpClient.req.Method)
			},
		},
		{
			name: "OPTIONS",
			test: func(t *testing.T, host *testHost) {
				r.Equal(t, "OPTIONS", host.httpClient.req.Method)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			host := &testHost{
				httpClient: &testClient{},
			}

			s, err := New(newScript("",
				fmt.Sprintf(`import http from 'mokapi/http'
						 export default function() {
						 	return http.%v('http://foo.bar')
						 }`, strings.ToLower(tc.name))),
				host, static.JsConfig{})
			r.NoError(t, err)
			_, err = s.RunDefault()
			r.NoError(t, err)

			tc.test(t, host)
		})
	}
}
