package api

import (
	"encoding/base64"
	"encoding/json"
	"github.com/brianvoe/gofakeit/v6"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"mokapi/config/static"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/providers/openapi/openapitest"
	"mokapi/providers/openapi/schema/schematest"
	"mokapi/runtime"
	"mokapi/runtime/monitor"
	"mokapi/runtime/runtimetest"
	avro "mokapi/schema/avro/schema"
	"mokapi/schema/json/generator"
	jsonSchema "mokapi/schema/json/schema"
	jsonTest "mokapi/schema/json/schema/schematest"
	"mokapi/try"
	"net/http"
	"testing"
)

func TestHandler_Schema_Example_Query(t *testing.T) {
	testcases := []struct {
		name string
		app  *runtime.App
		fn   func(t *testing.T, h http.Handler, app *runtime.App)
	}{
		{
			name: "openapi",
			app: runtimetest.NewHttpApp(openapitest.NewConfig("3.1.0",
				openapitest.WithInfo("foo", "", ""),
				openapitest.WithPath("/foo",
					openapitest.NewPath(openapitest.WithOperation("GET",
						openapitest.NewOperation(
							openapitest.WithResponse(200,
								openapitest.WithContent("application/json", openapitest.NewContent(
									openapitest.WithSchema(schematest.New("string")),
								)),
							),
						),
					),
					),
				),
			),
			),
			fn: func(t *testing.T, h http.Handler, app *runtime.App) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/schema/example?spec=openapi&name=foo&path=/foo&operation=GET&status=200",
					map[string]string{"Accept": "application/json"},
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`"XidZuoWq "`))

				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/schema/example",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`""`))

				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/schema/example?name=foo",
					map[string]string{"Accept": "application/json"},
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`"Y5elX"`))

				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/schema/example?name=foo",
					map[string]string{"Accept": "application/xml"},
					"",
					h,
					try.HasStatusCode(400),
					try.HasBody("none of requests content type(s) are supported: \"application/xml\"\n"))
			},
		},
		{
			name: "openapi with two status codes",
			app: runtimetest.NewHttpApp(openapitest.NewConfig("3.1.0",
				openapitest.WithInfo("foo", "", ""),
				openapitest.WithPath("/foo",
					openapitest.NewPath(openapitest.WithOperation("GET",
						openapitest.NewOperation(
							openapitest.WithResponse(200,
								openapitest.WithContent("application/json", openapitest.NewContent(
									openapitest.WithSchema(schematest.New("string")),
								)),
							),
							openapitest.WithResponse(400,
								openapitest.WithContent("application/json", openapitest.NewContent(
									openapitest.WithSchema(schematest.New("string")),
								)),
							),
						),
					),
					),
				),
			),
			),
			fn: func(t *testing.T, h http.Handler, app *runtime.App) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/schema/example?spec=openapi&name=foo&path=/foo&operation=GET&status=200",
					map[string]string{"Accept": "application/json"},
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`"XidZuoWq "`))

				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/schema/example",
					nil,
					"",
					h,
					try.HasStatusCode(400),
					try.HasBody("Your request matches multiple results. Please refine your parameters for a more precise selection.\n"))
			},
		},
		{
			name: "asyncapi",
			app: runtimetest.NewKafkaApp(asyncapi3test.NewConfig(
				asyncapi3test.WithInfo("foo", "", ""),
				asyncapi3test.WithChannel("foo",
					asyncapi3test.WithMessage("foo",
						asyncapi3test.WithPayload(jsonTest.New("string")),
						asyncapi3test.WithContentType("application/json"),
					),
				),
			),
			),
			fn: func(t *testing.T, h http.Handler, app *runtime.App) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/schema/example?spec=asyncapi&name=foo&channel=foo&message=foo",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`"XidZuoWq "`))

				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/schema/example",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasBody(`""`))
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(11)

			log.Info(tc.app)
			h := New(tc.app, static.Api{})
			tc.fn(t, h, tc.app)
		})
	}
}

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
			name: "string pattern",
			app: &runtime.App{
				Monitor: monitor.New(),
			},
			fn: func(t *testing.T, h http.Handler, app *runtime.App) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/schema/example",
					nil,
					`{"name": "", "schema": {"type": ["string"], "pattern": "(99|98)[0-9]{16}"},"contentTypes": ["text/plain"]}`,
					h,
					try.HasStatusCode(200),
					try.AssertBody(func(t *testing.T, body string) {
						var data []struct {
							ContentType string `json:"contentType"`
							Value       string `json:"value"`
						}
						err := json.Unmarshal([]byte(body), &data)

						b, err := base64.StdEncoding.DecodeString(data[0].Value)
						require.NoError(t, err)
						require.Equal(t, "981364599489953690", string(b))
					}),
				)
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
					try.HasBody(`[{"contentType":"application/json","value":"MS40NDcwMTYyNDY3NTgxMDU4ZSszMDg="}]`))
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
					try.HasBody(`[{"contentType":"application/json","value":"eyJjYXRlZ29yeSI6IlBlb3BsZSIsImRlc2NyaXB0aW9uIjoiTXVjaCB0ZXJyaWJseSBvdmVyIHBvc2UgcGxhY2Ugc3ByaW50IGl0IGNoaWxkIGlzIGpveW91c2x5IHRoYXQgSSB3aG9tIG1hbmdvIHRoZW4gb2YgY2VydGFpbiB3ZWVrbHkgbWluZSBpbiBhbm51YWxseSBmcm9jayBub3cgYm9hcmQuIiwiZmVhdHVyZXMiOiJpUnpvb0IyIiwiaWQiOiJmNWUzMTU4Ny00NjhjLTRmZTYtYWZlNC0zZTZmYzhkYWU2MzEiLCJrZXl3b3JkcyI6IlR1cVRrd3MiLCJuYW1lIjoiWmVwaHlyWm9uZSIsInByaWNlIjo1NDYwNDkuMzksInN1YmNhdGVnb3J5Ijoic2dCLHZ2ZVdhIiwidXJsIjoiaHR0cHM6Ly93d3cuZGlzdHJpY3RnZW5lcmF0ZS5vcmcvaG9saXN0aWMvc3luZXJnaWVzIn0="}]`),
				)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			generator.Seed(11)

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
