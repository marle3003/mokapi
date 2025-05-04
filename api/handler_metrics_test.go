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

func TestHandler_Metrics(t *testing.T) {
	testcases := []struct {
		name string
		app  *runtime.App
		fn   func(t *testing.T, h http.Handler, app *runtime.App)
	}{
		{
			name: "/api/metrics",
			app: &runtime.App{
				Monitor: monitor.New(),
			},
			fn: func(t *testing.T, h http.Handler, app *runtime.App) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/metrics",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(fmt.Sprintf(`[{"name":"app_start_timestamp","value":%v},{"name":"app_memory_usage_bytes","value":0},{"name":"app_job_run_total","value":0}]`, int64(app.Monitor.StartTime.Value()))))
			},
		},
		{
			name: "/api/metrics?names=app_start_timestamp,app_memory_usage_bytes",
			app: &runtime.App{
				Monitor: monitor.New(),
			},
			fn: func(t *testing.T, h http.Handler, app *runtime.App) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/metrics?names=app_start_timestamp,app_memory_usage_bytes",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(fmt.Sprintf(`[{"name":"app_start_timestamp","value":%v},{"name":"app_memory_usage_bytes","value":0}]`, int64(app.Monitor.StartTime.Value()))))
			},
		},
		{
			name: "/api/metrics/kafka",
			app: &runtime.App{
				Monitor: monitor.New(),
			},
			fn: func(t *testing.T, h http.Handler, app *runtime.App) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/metrics/kafka",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`[]`))
			},
		},
		{
			name: "/api/metrics/kafka with metric",
			app: &runtime.App{
				Monitor: monitor.New(),
			},
			fn: func(t *testing.T, h http.Handler, app *runtime.App) {
				app.Monitor.Kafka.Messages.WithLabel("foo", "bar").Add(1)
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/metrics/kafka",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`[{"name":"kafka_messages_total{service=\"foo\",topic=\"bar\"}","value":1}]`))
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := New(tc.app, static.Api{})
			tc.fn(t, h, tc.app)
		})
	}
}
