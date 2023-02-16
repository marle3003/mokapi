package api

import (
	"github.com/brianvoe/gofakeit/v6"
	"mokapi/config/static"
	"mokapi/runtime"
	"mokapi/runtime/monitor"
	"mokapi/try"
	"net/http"
	"testing"
)

func TestHandler_Schema(t *testing.T) {
	testcases := []struct {
		name string
		app  *runtime.App
		fn   func(t *testing.T, h http.Handler, app *runtime.App)
	}{
		{
			name: "string",
			app: &runtime.App{
				Monitor: monitor.New(),
			},
			fn: func(t *testing.T, h http.Handler, app *runtime.App) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/schema/example",
					nil,
					`{"type": "string"}`,
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`"gbRMaRxHkiJBPta"`))
			},
		},
		{
			name: "object",
			app: &runtime.App{
				Monitor: monitor.New(),
			},
			fn: func(t *testing.T, h http.Handler, app *runtime.App) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/schema/example",
					nil,
					`{"type": "object", "properties": {"foo": {"type": "string"}}}`,
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`{"foo":"gbRMaRxHkiJBPta"}`))
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(11)

			h := New(tc.app, static.Api{})
			tc.fn(t, h, tc.app)
		})
	}
}
