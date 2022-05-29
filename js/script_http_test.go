package js

import (
	r "github.com/stretchr/testify/require"
	"io/ioutil"
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
					`
import http from 'http'
export default function() {
  var s = http.get('http://foo.bar')
return s
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
					`
import http from 'http'
export default function() {
  var s = http.get('http://foo.bar', {headers: {foo: "bar"}})
return s
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
					`
import http from 'http'
export default function() {
  var s = http.get('http://foo.bar', {headers: {foo: ["hello", "world"]}})
return s
}`,
					host)
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
				r.Equal(t, "hello,world", host.httpClient.req.Header.Get("foo"))
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
					`
import http from 'http'
export default function() {
  var s = http.post('http://foo.bar')
return s
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
			"content type",
			func(t *testing.T, host *testHost) {
				s, err := New("",
					`
import http from 'http'
export default function() {
  var s = http.post("http://localhost/foo", "body", {headers: {'Content-Type': "application/json"}})
return s
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
