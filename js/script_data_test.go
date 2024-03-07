package js

import (
	r "github.com/stretchr/testify/require"
	"mokapi/config/static"
	"mokapi/engine/common"
	"testing"
)

func TestScript_Data(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T, host *testHost)
	}{
		{
			"resource array",
			func(t *testing.T, host *testHost) {
				host.on = func(event string, do func(args ...interface{}) (bool, error), tags map[string]string) {
					r.Equal(t, "http", event)
					request := &common.EventRequest{}
					request.Url.Path = "/foo/bar"
					response := &common.EventResponse{}
					b, err := do(request, response)
					r.NoError(t, err)
					r.True(t, b)
					r.Equal(t, []interface{}{int64(1), int64(2), int64(3), int64(4)}, response.Data)
				}
				s, err := New(newScript("",
					`
export const mokapi = {http: {"bar": [1, 2, 3, 4]}}`),
					host, static.JsConfig{})
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
			},
		},
		{
			"resource absolute precedence ",
			func(t *testing.T, host *testHost) {
				host.on = func(event string, do func(args ...interface{}) (bool, error), tags map[string]string) {
					r.Equal(t, "http", event)
					request := &common.EventRequest{}
					request.Url.Path = "/foo/bar"
					response := &common.EventResponse{}
					b, err := do(request, response)
					r.NoError(t, err)
					r.True(t, b)
					r.Equal(t, []interface{}{int64(1), int64(2), int64(3), int64(4)}, response.Data)
				}
				s, err := New(newScript("",
					`
export const mokapi = {"http": {"bar": [5,6], "foo": {"bar": [1, 2, 3, 4]}}}`),
					host, static.JsConfig{})
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
			},
		},
		{
			"using default function",
			func(t *testing.T, host *testHost) {
				host.on = func(event string, do func(args ...interface{}) (bool, error), tags map[string]string) {
					r.Equal(t, "http", event)
					request := &common.EventRequest{}
					request.Url.Path = "/foo/bar"
					response := &common.EventResponse{}
					b, err := do(request, response)
					r.NoError(t, err)
					r.True(t, b)
					r.Equal(t, []interface{}{int64(1), int64(2), int64(3), int64(4)}, response.Data)
				}
				s, err := New(newScript("",
					`
export default () => {
	return {
		"http": {
			"bar": [5,6],
			"foo": {
				"bar": [1, 2, 3, 4]
			}
		}
	}
}`),
					host, static.JsConfig{})
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			host := &testHost{}

			tc.f(t, host)
		})
	}
}
