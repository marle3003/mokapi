package api

import (
	"fmt"
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
			name: "/api/dashboard",
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
					try.HasBody(fmt.Sprintf(`{"startTime":"%v","http":{"RequestCounter":{},"RequestErrorCounter":{}},"kafka":{"Messages":{}}}`,
						app.Monitor.StartTime.Format(time.RFC3339Nano))))
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
