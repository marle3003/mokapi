package api

import (
	"encoding/json"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/config/static"
	"mokapi/runtime"
	"mokapi/runtime/monitor"
	avro "mokapi/schema/avro/schema"
	"mokapi/schema/json/ref"
	jsonSchema "mokapi/schema/json/schema"
	"mokapi/try"
	"net/http"
	"testing"
)

func TestHandler_Schema_Example(t *testing.T) {
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
					`{"name": "", "schema": {"type": ["string"]}}`,
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`"XidZuoWq "`))
			},
		},
		{
			name: "parameter with type: string",
			app: &runtime.App{
				Monitor: monitor.New(),
			},
			fn: func(t *testing.T, h http.Handler, app *runtime.App) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/schema/example",
					map[string]string{"Accept": "text/plain"},
					`{"name": "", "schema": {"type": ["string"]}}`,
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "text/plain"),
					try.HasBody("XidZuoWq "))
			},
		},
		{
			name: "string or number",
			app: &runtime.App{
				Monitor: monitor.New(),
			},
			fn: func(t *testing.T, h http.Handler, app *runtime.App) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/schema/example",
					nil,
					`{"name": "", "schema": {"type": ["string","number"]}}`,
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody("1.644484108270445e+307"))
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
					`{"name": "", "schema": {"type": ["object"], "properties": {"foo": {"type": ["string"]}}}}`,
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
					`{"name": "", "schema": { "type": ["string"], "xml": { "name": "text" } }, "format": "application/vnd.oai.openapi;version=3.0.0"}`,
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
					`{"name": "", "schema": { "type": ["string"], "xml": { "name": "text" } }}`,
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
					try.HasBody("Content type application/* with schema format application/schema+json;version=2020-12 is not supported\n"))
			},
		},
		{
			name: "OpenAPI schema: string",
			app: &runtime.App{
				Monitor: monitor.New(),
			},
			fn: func(t *testing.T, h http.Handler, app *runtime.App) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/schema/example",
					nil,
					`{"name": "", "schema": {"type": ["string"]}, "format": "application/vnd.oai.openapi+json;version=3.0.0"}`,
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`"XidZuoWq "`))
			},
		},
		{
			name: "avro schema: string",
			app: &runtime.App{
				Monitor: monitor.New(),
			},
			fn: func(t *testing.T, h http.Handler, app *runtime.App) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/schema/example",
					nil,
					`{"name": "", "schema": {"type": ["string"]}, "format": "application/vnd.apache.avro;version=1.9.0"}`,
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`"XidZuoWq "`))
			},
		},
		{
			name: "schema without format",
			app:  &runtime.App{},
			fn: func(t *testing.T, h http.Handler, app *runtime.App) {
				try.Handler(t, http.MethodPost, "http://foo.api/api/schema/example", nil, `
{
  "schema": {
    "properties": {
      "category": {
        "type": "string"
      },
      "description": {
        "type": "string"
      },
      "features": {
        "type": "string"
      },
      "id": {
        "type": "string"
      },
      "keywords": {
        "type": "string"
      },
      "name": {
        "type": "string"
      },
      "price": {
        "type": "number"
      },
      "subcategory": {
        "type": "string"
      },
      "url": {
        "type": "string"
      }
    },
    "type": "object"
  }
}`,
					h,
					try.HasStatusCode(200),
					try.BodyContainsData(map[string]interface{}{"id": "eed4888d-99c1-4e10-85d6-8fce0adeb762"}),
				)
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

func TestSchemaInfo_UnmarshalJSON(t *testing.T) {
	testcases := []struct {
		name string
		data string
		test func(t *testing.T, s *schemaInfo, err error)
	}{
		{
			name: "empty",
			data: `{}`,
			test: func(t *testing.T, s *schemaInfo, err error) {
				require.NoError(t, err)
				require.Equal(t, &schemaInfo{}, s)
			},
		},
		{
			name: "format",
			data: `{"format": "foo"}`,
			test: func(t *testing.T, s *schemaInfo, err error) {
				require.NoError(t, err)
				require.Equal(t, &schemaInfo{Format: "foo"}, s)
			},
		},
		{
			name: "json schema",
			data: `{"schema": { "type": "object" }}`,
			test: func(t *testing.T, s *schemaInfo, err error) {
				require.NoError(t, err)
				require.NotNil(t, s.Schema)
				require.IsType(t, &jsonSchema.Ref{}, s.Schema)
				require.Equal(t, "object", s.Schema.(*jsonSchema.Ref).Value.Type[0])
			},
		},
		{
			name: "json schema with format",
			data: `{"schema": { "type": "object" }, "format": "foo"}`,
			test: func(t *testing.T, s *schemaInfo, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", s.Format)
				require.NotNil(t, s.Schema)
				require.IsType(t, &jsonSchema.Ref{}, s.Schema)
				require.Equal(t, "object", s.Schema.(*jsonSchema.Ref).Value.Type[0])
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var s *schemaInfo
			err := json.Unmarshal([]byte(tc.data), &s)
			tc.test(t, s, err)
		})
	}
}

func TestSchemaInfo_MarshalJSON(t *testing.T) {
	testcases := []struct {
		name string
		s    *schemaInfo
		test func(t *testing.T, s string, err error)
	}{
		{
			name: "empty",
			s:    &schemaInfo{},
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, "{}", s)
			},
		},
		{
			name: "format",
			s:    &schemaInfo{Format: "foo"},
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"format":"foo"}`, s)
			},
		},
		{
			name: "json schema only ref",
			s:    &schemaInfo{Schema: &jsonSchema.Ref{Reference: ref.Reference{Ref: "foo/bar"}}},
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"schema":{"ref":"foo/bar"}}`, s)
			},
		},
		{
			name: "json schema ref and value",
			s: &schemaInfo{Schema: &jsonSchema.Ref{
				Reference: ref.Reference{Ref: "foo/bar"},
				Value:     &jsonSchema.Schema{Type: jsonSchema.Types{"string"}},
			}},
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"schema":{"ref":"foo/bar","type":"string"}}`, s)
			},
		},
		{
			name: "avro",
			s:    &schemaInfo{Schema: &avro.Schema{Type: []interface{}{"string"}}},
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"schema":{"type":"string"}}`, s)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			b, err := json.Marshal(tc.s)
			tc.test(t, string(b), err)
		})
	}
}
