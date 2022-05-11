package api

import (
	"fmt"
	"mokapi/config/static"
	"mokapi/runtime"
	"mokapi/runtime/monitor"
	"mokapi/try"
	"net/http"
	"testing"
)

func TestHandler_Dashboard(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			name: "/api/dashboard empty",
			f: func(t *testing.T) {
				app := &runtime.App{
					Version: "1.0",
					Monitor: monitor.New(),
				}
				h := New(app, static.Api{})

				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/dashboard",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(fmt.Sprintf(`{"metrics":[{"name":"app_start_timestamp","value":%v},{"name":"app_memory_usage_bytes","value":0}]}`,
						int64(app.Monitor.StartTime.Value()))))
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.f(t)
		})
	}
}
