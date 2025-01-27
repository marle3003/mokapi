package api

import (
	"encoding/json"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/config/static"
	"mokapi/runtime"
	"mokapi/runtime/monitor"
	avro "mokapi/schema/avro/schema"
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
					try.HasBody(`[{"contentType":"application/json","value":"IlhpZFp1b1dxICI="}]`))
			},
		},
		{
			name: "text/plain",
			app: &runtime.App{
				Monitor: monitor.New(),
			},
			fn: func(t *testing.T, h http.Handler, app *runtime.App) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/schema/example",
					nil,
					`{"name": "", "schema": {"type": ["string"]}, "contentTypes": ["application/json"]}`,
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`[{"contentType":"application/json","value":"IlhpZFp1b1dxICI="}]`))
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
					try.HasBody(`[{"contentType":"application/json","value":"MS42NDQ0ODQxMDgyNzA0NDVlKzMwNw=="}]`))
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
					try.HasBody(`[{"contentType":"application/json","value":"eyJmb28iOiJYaWRadW9XcSAifQ=="}]`))
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
					nil,
					`{"name": "", "schema": { "type": ["string"], "xml": { "name": "text" } }, "format": "application/vnd.oai.openapi;version=3.0.0", "contentTypes": ["application/xml"]}`,
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`[{"contentType":"application/xml","value":"PHRleHQ+WGlkWnVvV3EgPC90ZXh0Pg=="}]`))
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
					nil,
					`{"name": "", "schema": { "type": ["string"], "xml": { "name": "text" } }, "contentTypes": ["*/*"]}`,
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`[{"contentType":"application/json","value":"IlhpZFp1b1dxICI="}]`))
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
					nil,
					`{"name": "", "schema": { "type": "string", "xml": { "name": "text" } }, "contentTypes": ["application/*"]}`,
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`[{"contentType":"application/*","value":null,"error":"Content type application/* with schema format application/schema+json;version=2020-12 is not supported"}]`))
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
					try.HasBody(`[{"contentType":"application/json","value":"IlhpZFp1b1dxICI="}]`))
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
					`{"name": "", "schema": {"type": ["string"]}, "format": "application/vnd.apache.avro;version=1.9.0", "contentTypes": ["avro/binary","application/json"]}`,
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`[{"contentType":"avro/binary","value":"ElhpZFp1b1dxIA=="},{"contentType":"application/json","value":"IlhpZFp1b1dxICI="}]`))
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
					try.HasBody(`[{"contentType":"application/json","value":"eyJjYXRlZ29yeSI6IkxpdGVyYXR1cmUiLCJkZXNjcmlwdGlvbiI6Ik91cnNlbHZlcyBleGFsdGF0aW9uIHdob20gdGhpcyBtZSBmYXIgc21pbGUgd2hlcmUgd2FzIGJ5IGFybXkgcGFydHkgcmljaGVzIHRoZWlycyBpbnN0ZWFkLiIsImZlYXR1cmVzIjoiRDRlemxZZWhDSUEwTyIsImlkIjoiZWVkNDg4OGQtOTljMS00ZTEwLTg1ZDYtOGZjZTBhZGViNzYyIiwia2V5d29yZHMiOiJXZ1NmaVlzZmZsbnpiIiwibmFtZSI6Ilplbml0aExpZ2h0IiwicHJpY2UiOjIzMTcyOC44Niwic3ViY2F0ZWdvcnkiOiJ4Q0t1eVkiLCJ1cmwiOiJodHRwczovL3d3dy5jaGllZnZpc3VhbGl6ZS5pby9zeW5kaWNhdGUifQ=="}]`),
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
				require.IsType(t, &jsonSchema.Schema{}, s.Schema)
				require.Equal(t, "object", s.Schema.(*jsonSchema.Schema).Type[0])
			},
		},
		{
			name: "json schema with format",
			data: `{"schema": { "type": "object" }, "format": "foo"}`,
			test: func(t *testing.T, s *schemaInfo, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", s.Format)
				require.NotNil(t, s.Schema)
				require.IsType(t, &jsonSchema.Schema{}, s.Schema)
				require.Equal(t, "object", s.Schema.(*jsonSchema.Schema).Type[0])
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
			s:    &schemaInfo{Schema: &jsonSchema.Schema{Ref: "foo/bar"}},
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"schema":{"$ref":"foo/bar"}}`, s)
			},
		},
		{
			name: "json schema ref and value",
			s: &schemaInfo{Schema: &jsonSchema.Schema{
				Ref:  "foo/bar",
				Type: jsonSchema.Types{"string"},
			}},
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"schema":{"$ref":"foo/bar","type":"string"}}`, s)
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
