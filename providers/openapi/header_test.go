package openapi_test

import (
	"encoding/json"
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	"mokapi/providers/openapi/schema/schematest"
	"net/http"
	"net/url"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestHeader_UnmarshalJSON(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "type is header",
			test: func(t *testing.T) {
				header := &openapi.Header{}
				err := json.Unmarshal([]byte(`{  }`), &header)
				require.NoError(t, err)
				require.Equal(t, openapi.ParameterHeader, header.Type)
			},
		},
		{
			name: "overwrite type",
			test: func(t *testing.T) {
				header := &openapi.Header{}
				err := json.Unmarshal([]byte(`{ "in": "cookie" }`), &header)
				require.NoError(t, err)
				require.Equal(t, openapi.ParameterHeader, header.Type)
			},
		},
		{
			name: "description",
			test: func(t *testing.T) {
				header := &openapi.Header{}
				err := json.Unmarshal([]byte(`{ "description": "foo" }`), &header)
				require.NoError(t, err)
				require.Equal(t, "foo", header.Description)
			},
		},
		{
			name: "required",
			test: func(t *testing.T) {
				header := &openapi.Header{}
				err := json.Unmarshal([]byte(`{ "required": true  }`), &header)
				require.NoError(t, err)
				require.True(t, header.Required)
			},
		},
		{
			name: "deprecated",
			test: func(t *testing.T) {
				header := &openapi.Header{}
				err := json.Unmarshal([]byte(`{ "deprecated": true  }`), &header)
				require.NoError(t, err)
				require.True(t, header.Deprecated)
			},
		},
		{
			name: "style",
			test: func(t *testing.T) {
				header := &openapi.Header{}
				err := json.Unmarshal([]byte(`{ "style": "simple"  }`), &header)
				require.NoError(t, err)
				require.Equal(t, "simple", header.Style)
			},
		},
		{
			name: "explode",
			test: func(t *testing.T) {
				header := &openapi.Header{}
				err := json.Unmarshal([]byte(`{ "explode": false  }`), &header)
				require.NoError(t, err)
				require.False(t, header.IsExplode())
			},
		},
		{
			name: "schema",
			test: func(t *testing.T) {
				header := &openapi.Header{}
				err := json.Unmarshal([]byte(`{ "schema": {}  }`), &header)
				require.NoError(t, err)
				require.NotNil(t, header.Schema)
			},
		},
		{
			name: "reference",
			test: func(t *testing.T) {
				ref := &openapi.HeaderRef{}
				err := json.Unmarshal([]byte(`{ "$ref": "foo.yml"  }`), &ref)
				require.NoError(t, err)
				require.Equal(t, "foo.yml", ref.Ref)
			},
		},
		{
			name: "value",
			test: func(t *testing.T) {
				ref := &openapi.HeaderRef{}
				err := json.Unmarshal([]byte(`{ "description": "foo"  }`), &ref)
				require.NoError(t, err)
				require.Equal(t, "foo", ref.Value.Description)
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

func TestHeader_UnmarshalYAML(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "type is header",
			test: func(t *testing.T) {
				header := &openapi.Header{}
				err := yaml.Unmarshal([]byte(`{}`), &header)
				require.NoError(t, err)
				require.Equal(t, openapi.ParameterHeader, header.Type)
			},
		},
		{
			name: "overwrite type",
			test: func(t *testing.T) {
				header := &openapi.Header{}
				err := yaml.Unmarshal([]byte(`in: cookie }`), &header)
				require.NoError(t, err)
				require.Equal(t, openapi.ParameterHeader, header.Type)
			},
		},
		{
			name: "description",
			test: func(t *testing.T) {
				header := &openapi.Header{}
				err := yaml.Unmarshal([]byte(`description: foo`), &header)
				require.NoError(t, err)
				require.Equal(t, "foo", header.Description)
			},
		},
		{
			name: "required",
			test: func(t *testing.T) {
				header := &openapi.Header{}
				err := yaml.Unmarshal([]byte(`required: true`), &header)
				require.NoError(t, err)
				require.True(t, header.Required)
			},
		},
		{
			name: "deprecated",
			test: func(t *testing.T) {
				header := &openapi.Header{}
				err := yaml.Unmarshal([]byte(`deprecated: true`), &header)
				require.NoError(t, err)
				require.True(t, header.Deprecated)
			},
		},
		{
			name: "style",
			test: func(t *testing.T) {
				header := &openapi.Header{}
				err := yaml.Unmarshal([]byte(`style: simple`), &header)
				require.NoError(t, err)
				require.Equal(t, "simple", header.Style)
			},
		},
		{
			name: "explode",
			test: func(t *testing.T) {
				header := &openapi.Header{}
				err := yaml.Unmarshal([]byte(`explode: false`), &header)
				require.NoError(t, err)
				require.False(t, header.IsExplode())
			},
		},
		{
			name: "schema",
			test: func(t *testing.T) {
				header := &openapi.Header{}
				err := yaml.Unmarshal([]byte(`schema: {}`), &header)
				require.NoError(t, err)
				require.NotNil(t, header.Schema)
			},
		},
		{
			name: "reference",
			test: func(t *testing.T) {
				ref := &openapi.HeaderRef{}
				err := yaml.Unmarshal([]byte(`$ref: foo.yml`), &ref)
				require.NoError(t, err)
				require.Equal(t, "foo.yml", ref.Ref)
			},
		},
		{
			name: "value",
			test: func(t *testing.T) {
				ref := &openapi.HeaderRef{}
				err := yaml.Unmarshal([]byte(`description: foo`), &ref)
				require.NoError(t, err)
				require.Equal(t, "foo", ref.Value.Description)
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

func TestHeader_Parse(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, log *test.Hook)
	}{
		{
			name: "reference is nil",
			test: func(t *testing.T, _ *test.Hook) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, nil
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo",
						openapitest.WithOperation(http.MethodGet,
							openapitest.WithResponse(http.StatusOK,
								openapitest.UseResponseHeaderRef("foo", nil),
							)),
					),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.NoError(t, err)
			},
		},
		{
			name: "header",
			test: func(t *testing.T, _ *test.Hook) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, nil
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo",
						openapitest.WithOperation(http.MethodGet,
							openapitest.WithResponse(http.StatusOK,
								openapitest.WithResponseHeader("foo", "foo description", schematest.New("string")),
							)),
					),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.NoError(t, err)
				h := config.Paths["/foo"].Value.Get.Responses.GetResponse(http.StatusOK).Headers["foo"]
				require.Equal(t, "foo description", h.Value.Description)
			},
		},
		{
			name: "error by resolving example ref",
			test: func(t *testing.T, log *test.Hook) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, fmt.Errorf("TEST ERROR")
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithInfo("HTTP API", "", ""),
					openapitest.WithPath("/foo",
						openapitest.WithOperation(http.MethodGet,
							openapitest.WithResponse(http.StatusOK,
								openapitest.WithResponseHeaderRef("foo", "foo"),
							)),
					),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.Equal(t, logrus.Fields{"method": "GET", "api": "HTTP API", "namespace": "http", "path": "/foo"}, log.LastEntry().Data)
				require.Equal(t, "parse response '200' failed: parse header 'foo' failed: resolve reference '/foo' failed: TEST ERROR", log.LastEntry().Message)
				require.Equal(t, openapi.StatusInvalid, config.Paths["/foo"].Value.Operation(http.MethodGet).Status)
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			hook := test.NewGlobal()
			tc.test(t, hook)
		})
	}
}

func TestConfig_Patch_Header(t *testing.T) {
	testcases := []struct {
		name    string
		configs []*openapi.Config
		test    func(t *testing.T, result *openapi.Config)
	}{
		{
			name: "add header",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.WithOperation(
						"post",
						openapitest.UseResponse(200, &openapi.Response{
							Headers: map[string]*openapi.HeaderRef{},
						}),
					),
				),
				),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.WithOperation(
						"post",
						openapitest.UseResponse(200, &openapi.Response{
							Headers: map[string]*openapi.HeaderRef{
								"foo": {Value: &openapi.Header{Parameter: openapi.Parameter{Description: "foo"}}},
							},
						}),
					),
				),
				),
			},
			test: func(t *testing.T, result *openapi.Config) {
				res := result.Paths["/foo"].Value.Post.Responses.GetResponse(200)
				require.Equal(t, "foo", res.Headers["foo"].Value.Description)
			},
		},
		{
			name: "patch header",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.WithOperation(
						"post",
						openapitest.UseResponse(200, &openapi.Response{
							Headers: map[string]*openapi.HeaderRef{
								"foo": {Value: &openapi.Header{Parameter: openapi.Parameter{Description: "foo"}}},
							},
						}),
					),
				),
				),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.WithOperation(
						"post",
						openapitest.UseResponse(200, &openapi.Response{
							Headers: map[string]*openapi.HeaderRef{
								"foo": {Value: &openapi.Header{Parameter: openapi.Parameter{Description: "bar"}}},
							},
						}),
					),
				),
				),
			},
			test: func(t *testing.T, result *openapi.Config) {
				res := result.Paths["/foo"].Value.Post.Responses.GetResponse(200)
				require.Equal(t, "bar", res.Headers["foo"].Value.Description)
			},
		},
		{
			name: "source header is nil",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.WithOperation(
						"post",
						openapitest.UseResponse(200, &openapi.Response{
							Headers: map[string]*openapi.HeaderRef{
								"foo": {Value: nil},
							},
						}),
					),
				),
				),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.WithOperation(
						"post",
						openapitest.UseResponse(200, &openapi.Response{
							Headers: map[string]*openapi.HeaderRef{
								"foo": {Value: &openapi.Header{Parameter: openapi.Parameter{Description: "foo"}}},
							},
						}),
					),
				),
				),
			},
			test: func(t *testing.T, result *openapi.Config) {
				res := result.Paths["/foo"].Value.Post.Responses.GetResponse(200)
				require.Equal(t, "foo", res.Headers["foo"].Value.Description)
			},
		},
		{
			name: "source header value is nil",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.WithOperation(
						"post",
						openapitest.UseResponse(200, &openapi.Response{
							Headers: map[string]*openapi.HeaderRef{
								"foo": {Value: nil},
							},
						}),
					),
				),
				),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.WithOperation(
						"post",
						openapitest.UseResponse(200, &openapi.Response{
							Headers: map[string]*openapi.HeaderRef{
								"foo": {Value: &openapi.Header{Parameter: openapi.Parameter{Description: "foo"}}},
							},
						}),
					),
				),
				),
			},
			test: func(t *testing.T, result *openapi.Config) {
				res := result.Paths["/foo"].Value.Post.Responses.GetResponse(200)
				require.Equal(t, "foo", res.Headers["foo"].Value.Description)
			},
		},
		{
			name: "patch headers is nil",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.WithOperation(
						"post",
						openapitest.UseResponse(200, &openapi.Response{
							Headers: map[string]*openapi.HeaderRef{
								"foo": {Value: &openapi.Header{Parameter: openapi.Parameter{Description: "foo"}}},
							},
						}),
					),
				),
				),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.WithOperation(
						"post",
						openapitest.UseResponse(200, &openapi.Response{
							Headers: map[string]*openapi.HeaderRef{
								"foo": nil,
							},
						}),
					),
				),
				),
			},
			test: func(t *testing.T, result *openapi.Config) {
				res := result.Paths["/foo"].Value.Post.Responses.GetResponse(200)
				require.Equal(t, "foo", res.Headers["foo"].Value.Description)
			},
		},
		{
			name: "patch header value is nil",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.WithOperation(
						"post",
						openapitest.UseResponse(200, &openapi.Response{
							Headers: map[string]*openapi.HeaderRef{
								"foo": {Value: &openapi.Header{Parameter: openapi.Parameter{Description: "foo"}}},
							},
						}),
					),
				),
				),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.WithOperation(
						"post",
						openapitest.UseResponse(200, &openapi.Response{
							Headers: map[string]*openapi.HeaderRef{
								"foo": {},
							},
						}),
					),
				),
				),
			},
			test: func(t *testing.T, result *openapi.Config) {
				res := result.Paths["/foo"].Value.Post.Responses.GetResponse(200)
				require.Equal(t, "foo", res.Headers["foo"].Value.Description)
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
