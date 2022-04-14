package api

import (
	"fmt"
	"mokapi/config/dynamic/openapi/openapitest"
	"mokapi/config/static"
	"mokapi/runtime"
	"mokapi/runtime/monitor"
	"mokapi/try"
	"net/http"
	"testing"
	"time"
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
					try.HasBody(fmt.Sprintf(`{"startTime":%v,"memoryUsage":0,"httpRequests":0,"httpErrorRequests":0,"kafkaMessages":0}`,
						int64(app.Monitor.StartTime.Value()))))
			},
		},
		{
			name: "/api/dashboard with services",
			f: func(t *testing.T) {
				app := &runtime.App{
					Monitor: monitor.New(),
					Http: map[string]*runtime.HttpInfo{
						"foo": {
							Config: openapitest.NewConfig("3.0.0", openapitest.WithInfo("foo", "", "")),
						},
					},
				}
				now := time.Now().Unix()
				app.Monitor.Http.LastRequest.WithLabel("foo").Set(float64(now))
				h := New(app, static.Api{})

				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/dashboard",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(
						fmt.Sprintf(
							`{"startTime":%v,"memoryUsage":0,"httpRequests":0,"httpErrorRequests":0,"kafkaMessages":0}`,
							now),
					),
				)
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
