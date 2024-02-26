package api

import (
	"mokapi/config/static"
	"mokapi/runtime"
	"mokapi/runtime/monitor"
	"mokapi/try"
	"net/http"
	"testing"
)

func TestHandler_Schema_Validate(t *testing.T) {
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
					"http://foo.api/api/schema/validate",
					map[string]string{
						"Data-Content-Type": "application/json",
					},
					`{ "schema": {"type": "string"}, "data":"\"foo\"" }`,
					h,
					try.HasBody(""),
					try.HasStatusCode(200),
				)
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
					"http://foo.api/api/schema/validate",
					map[string]string{
						"Data-Content-Type": "application/json",
					},
					`{ "schema": {"type": "object", "properties": { "foo":{ "type": "string" } } }, "data":"{\"foo\": \"bar\" }" }`,
					h,
					try.HasBody(""),
					try.HasStatusCode(200),
				)
			},
		},
		{
			name: "object invalid",
			app: &runtime.App{
				Monitor: monitor.New(),
			},
			fn: func(t *testing.T, h http.Handler, app *runtime.App) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/schema/validate",
					map[string]string{
						"Data-Content-Type": "application/json",
					},
					`{ "schema": {"type": "object", "properties": { "foo":{ "type": "string" }, "bar":{ "type": "integer" } } }, "data":"{\"foo\": 12, \"bar\": \"text\" }" }`,
					h,
					try.BodyContains("unmarshal data failed\n"),
					try.BodyContains("parse 'bar' failed: parse 'text' failed, expected schema type=integer\n"),
					try.BodyContains("parse 'foo' failed: parse 12 failed, expected schema type=string\n"),
					try.HasStatusCode(400),
				)
			},
		},
		{
			name: "object xml",
			app: &runtime.App{
				Monitor: monitor.New(),
			},
			fn: func(t *testing.T, h http.Handler, app *runtime.App) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/schema/validate",
					map[string]string{
						"Data-Content-Type": "application/xml",
					},
					`{ "schema": {"type": "object", "properties": { "foo":{ "type": "string" } }, "xml": { "name": "root" } }, "data":"<root><foo>bar</foo></root>" }`,
					h,
					try.HasBody(""),
					try.HasStatusCode(200),
				)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			h := New(tc.app, static.Api{})
			tc.fn(t, h, tc.app)
		})
	}
}
