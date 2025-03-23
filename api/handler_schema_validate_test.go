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
					nil,
					`{ "schema": {"type": ["string"]}, "data":"\"foo\"", "contentType": "application/json" }`,
					h,
					try.HasBody(""),
					try.HasStatusCode(204),
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
					nil,
					`{ "schema": {"type": ["object"], "properties": { "foo":{ "type": ["string"] } } }, "data":"{\"foo\": \"bar\" }", "contentType": "application/json" }`,
					h,
					try.HasBody(""),
					try.HasStatusCode(204),
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
					nil,
					`{ "schema": {"type": ["object"], "properties": { "foo":{ "type": ["string"] }, "bar":{ "type": ["integer"] } } }, "data":"{\"foo\": 12, \"bar\": \"text\" }", "contentType": "application/json" }`,
					h,
					try.BodyContains("- #/foo/type: invalid type, expected string but got number"),
					try.BodyContains("- #/bar/type: invalid type, expected integer but got string"),
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
					nil,
					`{ "schema": {"type": ["object"], "properties": { "foo":{ "type": ["string"] } }, "xml": { "name": "root" } }, "data":"<root><foo>bar</foo></root>", "format": "application/vnd.oai.openapi;version=3.0.0", "contentType": "application/xml" }`,
					h,
					try.HasBody(""),
					try.HasStatusCode(204),
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
					nil,
					`{ "schema": {"type": ["object"], "properties": { "foo":{ "type": ["string"] } }, "additionalProperties": false }, "data":"{ \"foo\":\"bar\", \"foo2\": \"val\" }", "contentType": "application/json" }`,
					h,
					try.HasStatusCode(400),
					try.HasBody("error count 1:\n- #/additionalProperties: property 'foo2' not defined and the schema does not allow additional properties\n"),
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
					nil,
					`{ "schema": {"type": ["object"], "properties": { "foo":{ "type": ["string"] } }, "additionalProperties": false }, "data":"{ \"foo\":\"bar\", \"foo2\": \"val\", \"foo3\": \"val\" }", "contentType": "application/json" }`,
					h,
					try.HasStatusCode(400),
					try.BodyContains("- #/additionalProperties: property 'foo2' not defined and the schema does not allow additional properties"),
					try.BodyContains("property 'foo3' not defined and the schema does not allow additional properties"),
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
					nil,
					`{ "schema": {"type": ["object"], "properties": { "foo":{ "type": ["string"] } }, "additionalProperties": false }, "data":"{\"foo2\": \"val\" }", "contentType": "application/json" }`,
					h,
					try.HasStatusCode(400),
					try.HasBody("error count 1:\n- #/additionalProperties: property 'foo2' not defined and the schema does not allow additional properties\n"),
				)
			},
		},
		{
			name: "JSON schema contains reference and the value",
			app: &runtime.App{
				Monitor: monitor.New(),
			},
			fn: func(t *testing.T, h http.Handler, app *runtime.App) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/schema/validate",
					nil,
					`{ "schema": { "$ref": "#/components/schemas/Foo", "type": ["object"], "properties": { "foo":{ "type": ["string"] } } }, "data":"{\"foo\": 123 }", "contentType": "application/json" }`,
					h,
					try.HasBody("error count 1:\n- #/foo/type: invalid type, expected string but got number\n"),
					try.HasStatusCode(400),
				)
			},
		},
		{
			name: "OpenAPI schema contains reference and the value",
			app: &runtime.App{
				Monitor: monitor.New(),
			},
			fn: func(t *testing.T, h http.Handler, app *runtime.App) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/schema/validate",
					nil,
					`{ "format": "application/vnd.oai.openapi+json;version=3.0.0", "schema": { "$ref": "#/components/schemas/Foo", "type": ["object"], "properties": { "foo":{ "type": ["string"] } } }, "data":"{\"foo\": 123 }", "contentType": "application/json" }`,
					h,
					try.HasBody("error count 1:\n- #/foo/type: invalid type, expected string but got number\n"),
					try.HasStatusCode(400),
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
