package js

import (
	"fmt"
	r "github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestScript_Http_Get(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T, host *testHost)
	}{
		{
			"simple",
			func(t *testing.T, host *testHost) {
				s, err := New("",
					`import http from 'mokapi/http'
						 export default function() {
						 	return http.get('http://foo.bar')
						 }`,
					host)
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
				r.Equal(t, "GET", host.httpClient.req.Method)
				r.Equal(t, "http://foo.bar", host.httpClient.req.URL.String())
			},
		},
		{
			"header",
			func(t *testing.T, host *testHost) {
				s, err := New("",
					`import http from 'mokapi/http'
						 export default function() {
						  	return http.get('http://foo.bar', {headers: {foo: "bar"}})
						 }`,
					host)
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
				r.Equal(t, "bar", host.httpClient.req.Header.Get("foo"))
			},
		},
		{
			"header with array",
			func(t *testing.T, host *testHost) {
				s, err := New("",
					`import http from 'mokapi/http'
						 export default function() {
						  	return http.get('http://foo.bar', {headers: {foo: ["hello", "world"]}})
						 }`,
					host)
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
				r.Equal(t, []string{"hello", "world"}, host.httpClient.req.Header.Values("foo"))
			},
		},
		{
			"header set to null",
			func(t *testing.T, host *testHost) {
				s, err := New("",
					`import http from 'mokapi/http'
						 export default function() {
						  	return http.get('http://foo.bar', {headers: null})
						 }`,
					host)
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
				r.Len(t, host.httpClient.req.Header, 0)
			},
		},
		{
			"invalid url",
			func(t *testing.T, host *testHost) {
				s, err := New("",
					`import http from 'mokapi/http'
						 export default function() {
						  	return http.get('://')
						 }`,
					host)
				r.NoError(t, err)
				err = s.Run()
				r.Error(t, err)
			},
		},
		{
			"response body",
			func(t *testing.T, host *testHost) {
				host.httpClient.doFunc = func(request *http.Request) (*http.Response, error) {
					return &http.Response{Body: io.NopCloser(strings.NewReader("hello world"))}, nil
				}
				s, err := New("",
					`import http from 'mokapi/http'
						 export default function() {
						  	return http.get('http://foo.bar')
						 }`,
					host)
				r.NoError(t, err)
				v, err := s.RunDefault()
				r.NoError(t, err)
				result := v.Export().(response)
				r.Equal(t, "hello world", result.Body)
			},
		},
		{
			"response body json",
			func(t *testing.T, host *testHost) {
				host.httpClient.doFunc = func(request *http.Request) (*http.Response, error) {
					return &http.Response{Body: io.NopCloser(strings.NewReader(`{"foo": "bar"}`))}, nil
				}
				s, err := New("",
					`import http from 'mokapi/http'
						 export default function() {
						  	return http.get('http://foo.bar').json()
						 }`,
					host)
				r.NoError(t, err)
				v, err := s.RunDefault()
				r.NoError(t, err)
				result := v.Export()
				r.Equal(t, map[string]interface{}{"foo": "bar"}, result)
			},
		},
		{
			"response header",
			func(t *testing.T, host *testHost) {
				host.httpClient.doFunc = func(request *http.Request) (*http.Response, error) {
					return &http.Response{Header: map[string][]string{"Allow": {"OPTIONS", "GET", "HEAD", "POST"}}}, nil
				}
				s, err := New("",
					`import http from 'mokapi/http'
						 export default function() {
						  	const res = http.options('https://foo.bar')
							return res.headers['Allow']
						 }`,
					host)
				r.NoError(t, err)
				v, err := s.RunDefault()
				r.NoError(t, err)
				result := v.Export()
				r.Equal(t, []string{"OPTIONS", "GET", "HEAD", "POST"}, result)
			},
		},
		{
			"client error",
			func(t *testing.T, host *testHost) {
				host.httpClient.doFunc = func(request *http.Request) (*http.Response, error) {
					return nil, fmt.Errorf("test error")
				}
				s, err := New("",
					`import http from 'mokapi/http'
						 export default function() {
						  	return http.get('http://foo.bar')
						 }`,
					host)
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

			host := &testHost{
				httpClient: &testClient{},
			}

			tc.f(t, host)
		})
	}
}

func TestScript_Http_Post(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T, host *testHost)
	}{
		{
			"simple",
			func(t *testing.T, host *testHost) {
				s, err := New("",
					`import http from 'mokapi/http'
						 export default function() {
  							return http.post('http://foo.bar')
						 }`,
					host)
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
				r.Equal(t, "POST", host.httpClient.req.Method)
				r.Equal(t, "http://foo.bar", host.httpClient.req.URL.String())
			},
		},
		{
			"json with body as string",
			func(t *testing.T, host *testHost) {
				s, err := New("",
					`import http from 'mokapi/http'
						 export default function() {
						  	return http.post("http://localhost/foo", "body", {headers: {'Content-Type': "application/json"}})
						 }`,
					host)
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
				b, err := ioutil.ReadAll(host.httpClient.req.Body)
				r.Equal(t, "body", string(b))
				r.Equal(t, "application/json", host.httpClient.req.Header.Get("Content-Type"))
			},
		},
		{
			"json with body as object",
			func(t *testing.T, host *testHost) {
				s, err := New("",
					`import http from 'mokapi/http'
						 export default function() {
						  	return http.post("http://localhost/foo", {"foo":"bar"}, {headers: {'Content-Type': "application/json"}})
						 }`,
					host)
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
				b, err := io.ReadAll(host.httpClient.req.Body)
				r.Equal(t, `{"foo":"bar"}`, string(b))
				r.Equal(t, "application/json", host.httpClient.req.Header.Get("Content-Type"))
			},
		},
		{
			"json with body as object without Content-Type",
			func(t *testing.T, host *testHost) {
				s, err := New("",
					`import http from 'mokapi/http'
						 export default function() {
						  	return http.post("http://localhost/foo", {"foo":"bar"})
						 }`,
					host)
				r.NoError(t, err)
				err = s.Run()
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

			tc.f(t, host)
		})
	}
}

func TestScript_Http(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T, host *testHost)
	}{
		{
			"GET",
			func(t *testing.T, host *testHost) {
				r.Equal(t, "GET", host.httpClient.req.Method)
			},
		},
		{
			"POST",
			func(t *testing.T, host *testHost) {
				r.Equal(t, "POST", host.httpClient.req.Method)
			},
		},
		{
			"PUT",
			func(t *testing.T, host *testHost) {
				r.Equal(t, "PUT", host.httpClient.req.Method)
			},
		},
		{
			"HEAD",
			func(t *testing.T, host *testHost) {
				r.Equal(t, "HEAD", host.httpClient.req.Method)
			},
		},
		{
			"PATCH",
			func(t *testing.T, host *testHost) {
				r.Equal(t, "PATCH", host.httpClient.req.Method)
			},
		},
		{
			"DEL",
			func(t *testing.T, host *testHost) {
				r.Equal(t, "DELETE", host.httpClient.req.Method)
			},
		},
		{
			"DELETE",
			func(t *testing.T, host *testHost) {
				r.Equal(t, "DELETE", host.httpClient.req.Method)
			},
		},
		{
			"OPTIONS",
			func(t *testing.T, host *testHost) {
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

			s, err := New("",
				fmt.Sprintf(`import http from 'mokapi/http'
						 export default function() {
						 	return http.%v('http://foo.bar')
						 }`, strings.ToLower(tc.name)),
				host)
			r.NoError(t, err)
			err = s.Run()
			r.NoError(t, err)

			tc.f(t, host)
		})
	}
}
