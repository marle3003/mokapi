package js_test

import (
	r "github.com/stretchr/testify/require"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/jstest"
	"testing"
)

func TestScript_Console(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, host *enginetest.Host)
	}{
		{
			name: "log",
			test: func(t *testing.T, host *enginetest.Host) {
				var log interface{}
				host.InfoFunc = func(args ...interface{}) {
					log = args[0]
				}
				s, err := jstest.New(
					jstest.WithSource(
						`export default function() {
						 	console.log('foo')
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
				r.Equal(t, "foo", log)
			},
		},
		{
			name: "log",
			test: func(t *testing.T, host *enginetest.Host) {
				var log interface{}
				host.InfoFunc = func(args ...interface{}) {
					log = args[0]
				}
				s, err := jstest.New(
					jstest.WithSource(
						`export default function() {
						 	console.log({ foo: 123, bar: 'mokapi' })
						 }`),
					js.WithHost(host))
				r.NoError(t, err)
				err = s.Run()
				r.NoError(t, err)
				r.Equal(t, `{"foo":123,"bar":"mokapi"}`, log)
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
