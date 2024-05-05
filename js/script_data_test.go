package js_test

import (
	r "github.com/stretchr/testify/require"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/jstest"
	"testing"
)

func TestScript_Data(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, host *enginetest.Host)
	}{
		{
			name: "resource array",
			test: func(t *testing.T, host *enginetest.Host) {
				host.OnFunc = func(event string, do func(args ...interface{}) (bool, error), tags map[string]string) {
					r.Equal(t, "http", event)
					request := &common.EventRequest{}
					request.Url.Path = "/foo/bar"
					response := &common.EventResponse{}
					b, err := do(request, response)
					r.NoError(t, err)
					r.True(t, b)
					r.Equal(t, []interface{}{int64(1), int64(2), int64(3), int64(4)}, response.Data)
				}
				s, err := jstest.New(jstest.WithSource(`
export const mokapi = {http: {"bar": [1, 2, 3, 4]}}`),
					js.WithHost(host))
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
			},
		},
		{
			name: "resource absolute precedence ",
			test: func(t *testing.T, host *enginetest.Host) {
				host.OnFunc = func(event string, do func(args ...interface{}) (bool, error), tags map[string]string) {
					r.Equal(t, "http", event)
					request := &common.EventRequest{}
					request.Url.Path = "/foo/bar"
					response := &common.EventResponse{}
					b, err := do(request, response)
					r.NoError(t, err)
					r.True(t, b)
					r.Equal(t, []interface{}{int64(1), int64(2), int64(3), int64(4)}, response.Data)
				}
				s, err := jstest.New(jstest.WithSource(`
export const mokapi = {"http": {"bar": [5,6], "foo": {"bar": [1, 2, 3, 4]}}}`),
					js.WithHost(host))
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
			},
		},
		{
			name: "using default function",
			test: func(t *testing.T, host *enginetest.Host) {
				host.OnFunc = func(event string, do func(args ...interface{}) (bool, error), tags map[string]string) {
					r.Equal(t, "http", event)
					request := &common.EventRequest{}
					request.Url.Path = "/foo/bar"
					response := &common.EventResponse{}
					b, err := do(request, response)
					r.NoError(t, err)
					r.True(t, b)
					r.Equal(t, []interface{}{int64(1), int64(2), int64(3), int64(4)}, response.Data)
				}
				s, err := jstest.New(jstest.WithSource(`
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
					js.WithHost(host))
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
			tc.test(t, &enginetest.Host{})
		})
	}
}
