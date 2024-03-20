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
					`{"name": "", "schema": {"type": "string"}}`,
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`"XidZuoWq "`))
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
					`{"name": "", "schema": {"type": "object", "properties": {"foo": {"type": "string"}}}}`,
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`{"foo":"XidZuoWq "}`))
			},
		},
		{
			name: "string accept application/xml",
			app: &runtime.App{
				Monitor: monitor.New(),
			},
			fn: func(t *testing.T, h http.Handler, app *runtime.App) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/schema/example",
					map[string]string{"Accept": "application/xml"},
					`{"name": "", "schema": { "type": "string", "xml": { "name": "text" } }}`,
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/xml"),
					try.HasBody(`<text>XidZuoWq </text>`))
			},
		},
		{
			name: "string accept */*",
			app: &runtime.App{
				Monitor: monitor.New(),
			},
			fn: func(t *testing.T, h http.Handler, app *runtime.App) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/schema/example",
					map[string]string{"Accept": "*/*"},
					`{"name": "", "schema": { "type": "string", "xml": { "name": "text" } }}`,
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`"XidZuoWq "`))
			},
		},
		{
			name: "string accept application/*",
			app: &runtime.App{
				Monitor: monitor.New(),
			},
			fn: func(t *testing.T, h http.Handler, app *runtime.App) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/schema/example",
					map[string]string{"Accept": "application/*"},
					`{"name": "", "schema": { "type": "string", "xml": { "name": "text" } }}`,
					h,
					try.HasStatusCode(400),
					try.HasHeader("Content-Type", "text/plain; charset=utf-8"),
					try.HasBody("Content type application/* not supported. Only json or xml are supported\n"))
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
