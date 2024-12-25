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
					`{ "schema": {"type": ["string"]}, "data":"\"foo\"" }`,
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
					`{ "schema": {"type": ["object"], "properties": { "foo":{ "type": ["string"] } } }, "data":"{\"foo\": \"bar\" }" }`,
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
					`{ "schema": {"type": ["object"], "properties": { "foo":{ "type": ["string"] }, "bar":{ "type": ["integer"] } } }, "data":"{\"foo\": 12, \"bar\": \"text\" }" }`,
					h,
					try.BodyContains("invalid type, expected string but got number\nschema path #/foo/type\n"),
					try.BodyContains("invalid type, expected integer but got string\nschema path #/bar/type\n"),
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
					`{ "schema": {"type": ["object"], "properties": { "foo":{ "type": ["string"] } }, "xml": { "name": "root" } }, "data":"<root><foo>bar</foo></root>" }`,
					h,
					try.HasBody(""),
					try.HasStatusCode(200),
				)
			},
		},
		{
			name: "object with additionalProperty=false contains one not defined properties",
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
					`{ "schema": {"type": ["object"], "properties": { "foo":{ "type": ["string"] } }, "additionalProperties": false }, "data":"{ \"foo\":\"bar\", \"foo2\": \"val\" }" }`,
					h,
					try.HasStatusCode(400),
					try.HasBody("found 1 error:\nproperty 'foo2' not defined and the schema does not allow additional properties\nschema path #/additionalProperties\n"),
				)
			},
		},
		{
			name: "object with additionalProperty=false contains two not defined properties",
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
					`{ "schema": {"type": ["object"], "properties": { "foo":{ "type": ["string"] } }, "additionalProperties": false }, "data":"{ \"foo\":\"bar\", \"foo2\": \"val\", \"foo3\": \"val\" }" }`,
					h,
					try.HasStatusCode(400),
					try.BodyContains("property 'foo2' not defined and the schema does not allow additional properties\n"),
					try.BodyContains("property 'foo3' not defined and the schema does not allow additional properties\nschema path #/additionalProperties\n"),
				)
			},
		},
		{
			name: "object with additionalProperty=false but match number of properties",
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
					`{ "schema": {"type": ["object"], "properties": { "foo":{ "type": ["string"] } }, "additionalProperties": false }, "data":"{\"foo2\": \"val\" }" }`,
					h,
					try.HasStatusCode(400),
					try.HasBody("found 1 error:\nproperty 'foo2' not defined and the schema does not allow additional properties\nschema path #/additionalProperties\n"),
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
