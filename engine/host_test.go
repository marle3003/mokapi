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
		test func(t *testing.T, host *scriptHost)
	}{
		{
			name: "every but one time",
			test: func(t *testing.T, host *scriptHost) {
				opt := common.JobOptions{Times: 1}
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
			name: "every but one time and not immediately",
			test: func(t *testing.T, host *scriptHost) {
				opt := common.JobOptions{Times: 1, SkipImmediateFirstRun: true}
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

				select {
				case <-ch:
					counter++
				case <-time.After(150 * time.Millisecond):
					break
				}

				require.Equal(t, 1, counter)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			engine := New(&dynamictest.Reader{}, runtime.New(), static.JsConfig{}, false)
			engine.Start()
			defer engine.Close()

			tc.test(t, newScriptHost(newScript("test.js", ""), engine))
		})
	}
}
