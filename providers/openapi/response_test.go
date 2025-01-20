package openapi_test

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/media"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	"mokapi/providers/openapi/schema/schematest"
	"net/http"
	"net/url"
	"strconv"
	"testing"
)

func TestResponse_UnmarshalJSON(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "200",
			test: func(t *testing.T) {
				res := openapi.Responses{}
				err := json.Unmarshal([]byte(`{ "200": { "description": "foo" } }`), &res)
				require.NoError(t, err)
				require.Equal(t, 1, res.Len())
				r, _ := res.Get("200")
				require.Equal(t, "foo", r.Value.Description)
			},
		},
		{
			name: "default status code",
			test: func(t *testing.T) {
				res := openapi.Responses{}
				err := json.Unmarshal([]byte(`{ "default": { "description": "foo" } }`), &res)
				require.NoError(t, err)
				require.Equal(t, 1, res.Len())
				r, _ := res.Get("default")
				require.Equal(t, "foo", r.Value.Description)
			},
		},
		{
			name: "invalid status code",
			test: func(t *testing.T) {
				res := openapi.Responses{}
				err := json.Unmarshal([]byte(`{ "foo": { "description": "foo" } }`), &res)
				// For JSON files, line and column position for error is handled in dynamic package
				require.EqualError(t, err, "structural error at foo: invalid http status code 'foo': only valid HTTP status codes, default or range (1XX, 2XX,...) are allowed")
				require.Equal(t, 0, res.Len())
			},
		},
		{
			name: "unexpected array",
			test: func(t *testing.T) {
				res := openapi.Responses{}
				err := json.Unmarshal([]byte(`[]`), &res)
				require.EqualError(t, err, "expected openapi.Responses map, got [")
				require.Equal(t, 0, res.Len())
			},
		},
		{
			name: "response unexpected array",
			test: func(t *testing.T) {
				res := openapi.Responses{}
				err := json.Unmarshal([]byte(`{ "200": [{ "description": "foo" }] }`), &res)
				require.EqualError(t, err, "structural error at 200: expected object but received an array")
				require.Equal(t, 0, res.Len())
			},
		},
		{
			name: "check order of response",
			test: func(t *testing.T) {
				res := openapi.Responses{}
				err := json.Unmarshal([]byte(`{ "200": { "description": "200" }, "204": { "description": "204" }, "202": { "description": "202" }, "301": { "description": "301" }, "404": { "description": "404" } }`), &res)
				require.NoError(t, err)
				require.Equal(t, 5, res.Len())
				require.Equal(t, []string{"200", "204", "202", "301", "404"}, res.Keys())
			},
		},
		{
			name: "range 2XX",
			test: func(t *testing.T) {
				responses := openapi.Responses{}
				err := json.Unmarshal([]byte(`{ "2XX": { "description": "200" }, "204": { "description": "204" }, "301": { "description": "301" } }`), &responses)
				require.NoError(t, err)
				res := responses.GetResponse(202)
				require.NotNil(t, res)
				require.Equal(t, "200", res.Description)
			},
		},
		{
			name: "explicit takes precedence over the range definition",
			test: func(t *testing.T) {
				responses := openapi.Responses{}
				err := json.Unmarshal([]byte(`{ "2XX": { "description": "200" }, "204": { "description": "204" }, "301": { "description": "301" } }`), &responses)
				require.NoError(t, err)
				res := responses.GetResponse(204)
				require.NotNil(t, res)
				require.Equal(t, "204", res.Description)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.test(t)
		})
	}
}

func TestResponse_UnmarshalYAML(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "200",
			test: func(t *testing.T) {
				res := openapi.Responses{}
				err := yaml.Unmarshal([]byte(`'200': { description: foo }`), &res)
				require.NoError(t, err)
				require.Equal(t, 1, res.Len())
				r, _ := res.Get("200")
				require.Equal(t, "foo", r.Value.Description)
			},
		},
		{
			name: "default status code",
			test: func(t *testing.T) {
				res := openapi.Responses{}
				err := yaml.Unmarshal([]byte(`default: { description: foo }`), &res)
				require.NoError(t, err)
				require.Equal(t, 1, res.Len())
				r, _ := res.Get("default")
				require.Equal(t, "foo", r.Value.Description)
			},
		},
		{
			name: "invalid status code",
			test: func(t *testing.T) {
				res := openapi.Responses{}
				err := yaml.Unmarshal([]byte(`foo: { description: foo }`), &res)
				require.EqualError(t, err, "invalid http status code 'foo': only valid HTTP status codes, default or range (1XX, 2XX,...) are allowed at line 1, column 1")
				require.Equal(t, 0, res.Len())
			},
		},
		{
			name: "array instead of map",
			test: func(t *testing.T) {
				res := openapi.Responses{}
				err := yaml.Unmarshal([]byte(`- 200: [{ description: foo }]`), &res)
				require.EqualError(t, err, "expected openapi.Responses map, got !!seq")
				require.Equal(t, 0, res.Len())
			},
		},
		{
			name: "response unexpected array",
			test: func(t *testing.T) {
				res := openapi.Responses{}
				err := yaml.Unmarshal([]byte(`'200': [{ description: foo }]`), &res)
				require.EqualError(t, err, "yaml: unmarshal errors:\n  line 1: cannot unmarshal !!seq into openapi.Response")
				require.Equal(t, 0, res.Len())
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.test(t)
		})
	}
}

func TestResponses_GetResponse(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "200",
			test: func(t *testing.T) {
				r := &openapi.Responses{}
				r.Set(strconv.Itoa(http.StatusOK), &openapi.ResponseRef{Value: &openapi.Response{}})
				require.NotNil(t, r.GetResponse(http.StatusOK))
			},
		},
		{
			name: "200 & 201",
			test: func(t *testing.T) {
				r := &openapi.Responses{}
				r.Set(strconv.Itoa(http.StatusOK), &openapi.ResponseRef{Value: &openapi.Response{Description: "200"}})
				r.Set(strconv.Itoa(http.StatusCreated), &openapi.ResponseRef{Value: &openapi.Response{Description: "201"}})
				require.NotNil(t, r.GetResponse(http.StatusCreated))
				require.Equal(t, "201", r.GetResponse(http.StatusCreated).Description)
			},
		},
		{
			name: "default",
			test: func(t *testing.T) {
				r := &openapi.Responses{}
				r.Set("default", &openapi.ResponseRef{Value: &openapi.Response{}})
				require.NotNil(t, r.GetResponse(http.StatusOK))
			},
		},
		{
			name: "reference of ResponseRef not resolved returns nil",
			test: func(t *testing.T) {
				r := &openapi.Responses{}
				r.Set("200", &openapi.ResponseRef{})
				require.Nil(t, r.GetResponse(http.StatusOK))
			},
		},
		{
			name: "No error if ResponseRef is nil",
			test: func(t *testing.T) {
				r := &openapi.Responses{}
				r.Set("200", nil)
				require.Nil(t, r.GetResponse(http.StatusOK))
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.test(t)
		})
	}
}

func TestResponse_GetContent(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "application/json",
			test: func(t *testing.T) {
				r := &openapi.Response{Content: map[string]*openapi.MediaType{}}
				r.Content["application/json"] = &openapi.MediaType{ContentType: media.ParseContentType("application/json")}
				require.NotNil(t, r.GetContent(media.ParseContentType("application/json")))
			},
		},
		{
			name: "ACCEPT application/json with best match not at first place",
			test: func(t *testing.T) {
				r := &openapi.Response{Content: map[string]*openapi.MediaType{}}
				r.Content["application/*"] = &openapi.MediaType{ContentType: media.ParseContentType("application/*")}
				r.Content["application/json"] = &openapi.MediaType{ContentType: media.ParseContentType("application/json")}
				require.NotNil(t, r.GetContent(media.ParseContentType("application/json")))
				require.Equal(t, "application/json", r.GetContent(media.ParseContentType("application/json")).ContentType.Key())
			},
		},
		{
			name: "text/plain & application/json",
			test: func(t *testing.T) {
				r := &openapi.Response{Content: map[string]*openapi.MediaType{}}
				r.Content["application/json"] = &openapi.MediaType{ContentType: media.ParseContentType("application/json")}
				r.Content["text/plain"] = &openapi.MediaType{ContentType: media.ParseContentType("text/plain")}
				require.NotNil(t, r.GetContent(media.ParseContentType("text/plain")))
				require.Equal(t, "text/plain", r.GetContent(media.ParseContentType("text/plain")).ContentType.Key())
			},
		},
		{
			name: "no match",
			test: func(t *testing.T) {
				r := &openapi.Response{Content: map[string]*openapi.MediaType{}}
				r.Content["application/json"] = &openapi.MediaType{ContentType: media.ParseContentType("application/json")}
				r.Content["text/plain"] = &openapi.MediaType{ContentType: media.ParseContentType("text/plain")}
				require.Nil(t, r.GetContent(media.ParseContentType("text/html")))
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.test(t)
		})
	}
}

func TestResponse_Parse(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "no refs",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, nil
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo", openapitest.NewPath(
						openapitest.WithOperation(http.MethodGet, openapitest.NewOperation(
							openapitest.WithResponse(http.StatusOK),
						)),
					)),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.NoError(t, err)
			},
		},
		{
			name: "responses is nil",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, nil
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo", openapitest.NewPath(
						openapitest.WithOperation(http.MethodGet, &openapi.Operation{}),
					)),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.NoError(t, err)
			},
		},
		{
			name: "ResponseRef is nil",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, nil
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo", openapitest.NewPath(
						openapitest.WithOperation(http.MethodGet, openapitest.NewOperation(
							openapitest.WithResponseRef(http.StatusOK, nil),
						)),
					)),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.NoError(t, err)
			},
		},
		{
			name: "error by resolving response ref",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, fmt.Errorf("TEST ERROR")
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo", openapitest.NewPath(
						openapitest.WithOperation(http.MethodGet, openapitest.NewOperation(
							openapitest.WithResponseRef(http.StatusOK, &openapi.ResponseRef{Reference: dynamic.Reference{Ref: "foo.yml"}}),
						)),
					)),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.EqualError(t, err, "parse path '/foo' failed: parse operation 'GET' failed: parse response '200' failed: resolve reference 'foo.yml' failed: TEST ERROR")
			},
		},
		{
			name: "error by resolving header ref",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, fmt.Errorf("TEST ERROR")
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo", openapitest.NewPath(
						openapitest.WithOperation(http.MethodGet, openapitest.NewOperation(
							openapitest.WithResponse(http.StatusOK, openapitest.WithResponseHeaderRef("foo", &openapi.HeaderRef{Reference: dynamic.Reference{Ref: "foo.yml"}})),
						)),
					)),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.EqualError(t, err, "parse path '/foo' failed: parse operation 'GET' failed: parse response '200' failed: parse header 'foo' failed: resolve reference 'foo.yml' failed: TEST ERROR")
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.test(t)
		})
	}
}

func TestConfig_Patch_Response(t *testing.T) {
	testcases := []struct {
		name    string
		configs []*openapi.Config
		test    func(t *testing.T, result *openapi.Config)
	}{
		{
			name: "add response",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", &openapi.Operation{},
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithResponseDescription("foo"))),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				res := result.Paths["/foo"].Value.Post.Responses.GetResponse(200)
				require.Equal(t, "foo", res.Description)
			},
		},
		{
			name: "patch response",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(204, openapitest.WithResponseDescription("bar"))),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithResponseDescription("foo"))),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				res := result.Paths["/foo"].Value.Post.Responses.GetResponse(204)
				require.Equal(t, "bar", res.Description)
				res = result.Paths["/foo"].Value.Post.Responses.GetResponse(200)
				require.Equal(t, "foo", res.Description)
			},
		},
		{
			name: "patch is nil",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(204, openapitest.WithResponseDescription("bar"))),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", &openapi.Operation{}),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				res := result.Paths["/foo"].Value.Post.Responses.GetResponse(204)
				require.Equal(t, "bar", res.Description)
			},
		},
		{
			name: "patch response is nil",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(204, openapitest.WithResponseDescription("bar"))),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponseRef(202, &openapi.ResponseRef{}),
						),
					)))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				res := result.Paths["/foo"].Value.Post.Responses.GetResponse(202)
				require.Nil(t, res)
			},
		},
		{
			name: "patch description",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200)),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithResponseDescription("foo"))),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				res := result.Paths["/foo"].Value.Post.Responses.GetResponse(200)
				require.Equal(t, "foo", res.Description)
			},
		},
		{
			name: "patch add content type",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/plain", openapitest.NewContent()))),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("application/json", openapitest.NewContent()))),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				res := result.Paths["/foo"].Value.Post.Responses.GetResponse(200)
				require.Contains(t, res.Content, "text/plain")
				require.Contains(t, res.Content, "application/json")
			},
		},
		{
			name: "add content type schema",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/plain", openapitest.NewContent()))),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/plain",
								openapitest.NewContent(
									openapitest.WithSchema(schematest.New("number")))))),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				res := result.Paths["/foo"].Value.Post.Responses.GetResponse(200)
				require.Len(t, res.Content, 1)
				require.Equal(t, "number", res.Content["text/plain"].Schema.Value.Type.String())
			},
		},
		{
			name: "patch content type schema",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/plain",
								openapitest.NewContent(
									openapitest.WithSchema(schematest.New("number")))),
							),
						))))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/plain",
								openapitest.NewContent(
									openapitest.WithSchema(schematest.New("number", schematest.WithFormat("double")))),
							),
							)))))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				res := result.Paths["/foo"].Value.Post.Responses.GetResponse(200)
				require.Len(t, res.Content, 1)
				require.Equal(t, "number", res.Content["text/plain"].Schema.Value.Type.String())
				require.Equal(t, "double", res.Content["text/plain"].Schema.Value.Format)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			c := tc.configs[0]
			for _, p := range tc.configs[1:] {
				c.Patch(p)
			}
			tc.test(t, c)
		})
	}
}
