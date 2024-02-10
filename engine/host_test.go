package engine

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/engine/common"
	"mokapi/runtime"
	"testing"
	"time"
)

func TestHost_Every(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T, host *scriptHost)
	}{
		{
			"every but one time",
			func(t *testing.T, host *scriptHost) {
				opt := common.JobOptions{Times: 1, RunFirstTimeImmediately: true}
				var err error
				ch := make(chan bool, 1)
				_, err = host.Every("100ms", func() {
					ch <- true
				}, opt)
				require.NoError(t, err)

				var counter int
				select {
				case <-ch:
					counter++
				case <-time.After(50 * time.Millisecond):
					break
				}

				require.Equal(t, 1, counter)
			},
		},
		{
			"every but one time and not immediately",
			func(t *testing.T, host *scriptHost) {
				opt := common.JobOptions{Times: 1, RunFirstTimeImmediately: false}
				var err error
				ch := make(chan bool, 1)
				_, err = host.Every("100ms", func() {
					ch <- true
				}, opt)
				require.NoError(t, err)

				var counter int
				select {
				case <-ch:
					counter++
				case <-time.After(50 * time.Millisecond):
					break
				}

				require.Equal(t, 0, counter)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			engine := New(&dynamictest.Reader{}, runtime.New(), static.JsConfig{})
			engine.Start()
			defer engine.Close()

			tc.f(t, newScriptHost(newScript("test.js", ""), engine))
		})
	}
}
