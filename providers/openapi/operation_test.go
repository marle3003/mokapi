package openapi_test

import (
	"encoding/json"
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestOperation_UnmarshalJSON(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "tags",
			test: func(t *testing.T) {
				o := openapi.Operation{}
				err := json.Unmarshal([]byte(`{ "tags": [ "foo", "bar" ] }`), &o)
				require.NoError(t, err)
				require.Len(t, o.Tags, 2)
				require.Contains(t, o.Tags, "foo")
				require.Contains(t, o.Tags, "bar")
			},
		},
		{
			name: "summary",
			test: func(t *testing.T) {
				o := openapi.Operation{}
				err := json.Unmarshal([]byte(`{ "summary": "foo" }`), &o)
				require.NoError(t, err)
				require.Equal(t, "foo", o.Summary)
			},
		},
		{
			name: "description",
			test: func(t *testing.T) {
				o := openapi.Operation{}
				err := json.Unmarshal([]byte(`{ "description": "foo" }`), &o)
				require.NoError(t, err)
				require.Equal(t, "foo", o.Description)
			},
		},
		{
			name: "deprecated",
			test: func(t *testing.T) {
				o := openapi.Operation{}
				err := json.Unmarshal([]byte(`{ "deprecated": true }`), &o)
				require.NoError(t, err)
				require.True(t, o.Deprecated)
			},
		},
		{
			name: "operationId",
			test: func(t *testing.T) {
				o := openapi.Operation{}
				err := json.Unmarshal([]byte(`{ "operationId": "foo" }`), &o)
				require.NoError(t, err)
				require.Equal(t, "foo", o.OperationId)
			},
		},
		{
			name: "parameters",
			test: func(t *testing.T) {
				o := openapi.Operation{}
				err := json.Unmarshal([]byte(`{ "parameters": [] }`), &o)
				require.NoError(t, err)
				require.NotNil(t, o.Parameters)
			},
		},
		{
			name: "requestBody",
			test: func(t *testing.T) {
				o := openapi.Operation{}
				err := json.Unmarshal([]byte(`{ "requestBody": {} }`), &o)
				require.NoError(t, err)
				require.NotNil(t, o.RequestBody)
			},
		},
		{
			name: "responses",
			test: func(t *testing.T) {
				o := openapi.Operation{}
				err := json.Unmarshal([]byte(`{ "responses": {} }`), &o)
				require.NoError(t, err)
				require.NotNil(t, o.Responses)
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

func TestOperation_UnmarshalYAML(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "tags",
			test: func(t *testing.T) {
				o := openapi.Operation{}
				err := yaml.Unmarshal([]byte(`tags: [ "foo", "bar" ]`), &o)
				require.NoError(t, err)
				require.Len(t, o.Tags, 2)
				require.Contains(t, o.Tags, "foo")
				require.Contains(t, o.Tags, "bar")
			},
		},
		{
			name: "summary",
			test: func(t *testing.T) {
				o := openapi.Operation{}
				err := yaml.Unmarshal([]byte(`summary: "foo"`), &o)
				require.NoError(t, err)
				require.Equal(t, "foo", o.Summary)
			},
		},
		{
			name: "description",
			test: func(t *testing.T) {
				o := openapi.Operation{}
				err := yaml.Unmarshal([]byte(`description: "foo"`), &o)
				require.NoError(t, err)
				require.Equal(t, "foo", o.Description)
			},
		},
		{
			name: "deprecated",
			test: func(t *testing.T) {
				o := openapi.Operation{}
				err := yaml.Unmarshal([]byte(`deprecated: true`), &o)
				require.NoError(t, err)
				require.True(t, o.Deprecated)
			},
		},
		{
			name: "operationId",
			test: func(t *testing.T) {
				o := openapi.Operation{}
				err := yaml.Unmarshal([]byte(`operationId: "foo"`), &o)
				require.NoError(t, err)
				require.Equal(t, "foo", o.OperationId)
			},
		},
		{
			name: "parameters",
			test: func(t *testing.T) {
				o := openapi.Operation{}
				err := yaml.Unmarshal([]byte(`parameters: []`), &o)
				require.NoError(t, err)
				require.NotNil(t, o.Parameters)
			},
		},
		{
			name: "requestBody",
			test: func(t *testing.T) {
				o := openapi.Operation{}
				err := yaml.Unmarshal([]byte(`requestBody: {}`), &o)
				require.NoError(t, err)
				require.NotNil(t, o.RequestBody)
			},
		},
		{
			name: "responses",
			test: func(t *testing.T) {
				o := openapi.Operation{}
				err := yaml.Unmarshal([]byte(`responses: {}`), &o)
				require.NoError(t, err)
				require.NotNil(t, o.Responses)
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

func TestOperation_Parse(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, log *test.Hook)
	}{
		{
			name: "get",
			test: func(t *testing.T, _ *test.Hook) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, nil
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo",
						openapitest.WithOperation(http.MethodGet),
					),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.NoError(t, err)
				path := config.Paths["/foo"].Value
				require.Equal(t, path.Get.Path, path)
			},
		},
		{
			name: "get error",
			test: func(t *testing.T, log *test.Hook) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, fmt.Errorf("TEST ERROR")
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithInfo("HTTP API", "", ""),
					openapitest.WithPath("/foo",
						openapitest.WithOperation(http.MethodGet,
							openapitest.WithOperationParamRef("foo.yml"))),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.Equal(t, logrus.Fields{"method": "GET", "api": "HTTP API", "namespace": "http", "path": "/foo"}, log.LastEntry().Data)
				require.Equal(t, "parse parameter index '0' failed: resolve reference '/foo.yml' failed: TEST ERROR", log.LastEntry().Message)
				require.Equal(t, openapi.StatusInvalid, config.Paths["/foo"].Value.Operation(http.MethodGet).Status)
				require.NoError(t, err)
			},
		},
		{
			name: "post",
			test: func(t *testing.T, _ *test.Hook) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, nil
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo",
						openapitest.WithOperation(http.MethodPost),
					),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.NoError(t, err)
				path := config.Paths["/foo"].Value
				require.Equal(t, path.Post.Path, path)
			},
		},
		{
			name: "post error",
			test: func(t *testing.T, log *test.Hook) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, fmt.Errorf("TEST ERROR")
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithInfo("HTTP API", "", ""),
					openapitest.WithPath("/foo",
						openapitest.WithOperation(http.MethodPost,
							openapitest.WithOperationParamRef("foo.yml"))),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.Equal(t, logrus.Fields{"method": "POST", "api": "HTTP API", "namespace": "http", "path": "/foo"}, log.LastEntry().Data)
				require.Equal(t, "parse parameter index '0' failed: resolve reference '/foo.yml' failed: TEST ERROR", log.LastEntry().Message)
				require.Equal(t, openapi.StatusInvalid, config.Paths["/foo"].Value.Operation(http.MethodPost).Status)
				require.NoError(t, err)
			},
		},
		{
			name: "put",
			test: func(t *testing.T, _ *test.Hook) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, nil
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo",
						openapitest.WithOperation(http.MethodPut),
					),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.NoError(t, err)
				path := config.Paths["/foo"].Value
				require.Equal(t, path.Put.Path, path)
			},
		},
		{
			name: "put error",
			test: func(t *testing.T, log *test.Hook) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, fmt.Errorf("TEST ERROR")
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithInfo("HTTP API", "", ""),
					openapitest.WithPath("/foo",
						openapitest.WithOperation(http.MethodPut,
							openapitest.WithOperationParamRef("foo.yml"))),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.Equal(t, logrus.Fields{"method": "PUT", "api": "HTTP API", "namespace": "http", "path": "/foo"}, log.LastEntry().Data)
				require.Equal(t, "parse parameter index '0' failed: resolve reference '/foo.yml' failed: TEST ERROR", log.LastEntry().Message)
				require.Equal(t, openapi.StatusInvalid, config.Paths["/foo"].Value.Operation(http.MethodPut).Status)
				require.NoError(t, err)
			},
		},
		{
			name: "patch",
			test: func(t *testing.T, _ *test.Hook) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, nil
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo",
						openapitest.WithOperation(http.MethodPatch),
					),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.NoError(t, err)
				path := config.Paths["/foo"].Value
				require.Equal(t, path.Patch.Path, path)
			},
		},
		{
			name: "patch error",
			test: func(t *testing.T, log *test.Hook) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, fmt.Errorf("TEST ERROR")
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithInfo("HTTP API", "", ""),
					openapitest.WithPath("/foo",
						openapitest.WithOperation(http.MethodPatch,
							openapitest.WithOperationParamRef("foo.yml"))),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.Equal(t, logrus.Fields{"method": "PATCH", "api": "HTTP API", "namespace": "http", "path": "/foo"}, log.LastEntry().Data)
				require.Equal(t, "parse parameter index '0' failed: resolve reference '/foo.yml' failed: TEST ERROR", log.LastEntry().Message)
				require.Equal(t, openapi.StatusInvalid, config.Paths["/foo"].Value.Operation(http.MethodPatch).Status)
				require.NoError(t, err)
			},
		},
		{
			name: "delete",
			test: func(t *testing.T, _ *test.Hook) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, nil
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo",
						openapitest.WithOperation(http.MethodDelete),
					),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.NoError(t, err)
				path := config.Paths["/foo"].Value
				require.Equal(t, path.Delete.Path, path)
			},
		},
		{
			name: "delete error",
			test: func(t *testing.T, log *test.Hook) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, fmt.Errorf("TEST ERROR")
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithInfo("HTTP API", "", ""),
					openapitest.WithPath("/foo",
						openapitest.WithOperation(http.MethodDelete,
							openapitest.WithOperationParamRef("foo.yml"))),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.Equal(t, logrus.Fields{"method": "DELETE", "api": "HTTP API", "namespace": "http", "path": "/foo"}, log.LastEntry().Data)
				require.Equal(t, "parse parameter index '0' failed: resolve reference '/foo.yml' failed: TEST ERROR", log.LastEntry().Message)
				require.Equal(t, openapi.StatusInvalid, config.Paths["/foo"].Value.Operation(http.MethodDelete).Status)
				require.NoError(t, err)
			},
		},
		{
			name: "head",
			test: func(t *testing.T, _ *test.Hook) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, nil
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo",
						openapitest.WithOperation(http.MethodHead),
					),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.NoError(t, err)
				path := config.Paths["/foo"].Value
				require.Equal(t, path.Head.Path, path)
			},
		},
		{
			name: "head error",
			test: func(t *testing.T, log *test.Hook) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, fmt.Errorf("TEST ERROR")
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithInfo("HTTP API", "", ""),
					openapitest.WithPath("/foo",
						openapitest.WithOperation(http.MethodHead,
							openapitest.WithOperationParamRef("foo.yml"))),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.Equal(t, logrus.Fields{"method": "HEAD", "api": "HTTP API", "namespace": "http", "path": "/foo"}, log.LastEntry().Data)
				require.Equal(t, "parse parameter index '0' failed: resolve reference '/foo.yml' failed: TEST ERROR", log.LastEntry().Message)
				require.Equal(t, openapi.StatusInvalid, config.Paths["/foo"].Value.Operation(http.MethodHead).Status)
				require.NoError(t, err)
			},
		},
		{
			name: "options",
			test: func(t *testing.T, _ *test.Hook) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, nil
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo",
						openapitest.WithOperation(http.MethodOptions),
					),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.NoError(t, err)
				path := config.Paths["/foo"].Value
				require.Equal(t, path.Options.Path, path)
			},
		},
		{
			name: "options error",
			test: func(t *testing.T, log *test.Hook) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, fmt.Errorf("TEST ERROR")
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithInfo("HTTP API", "", ""),
					openapitest.WithPath("/foo",
						openapitest.WithOperation(http.MethodOptions,
							openapitest.WithOperationParamRef("foo.yml"))),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.Equal(t, logrus.Fields{"method": "OPTIONS", "api": "HTTP API", "namespace": "http", "path": "/foo"}, log.LastEntry().Data)
				require.Equal(t, "parse parameter index '0' failed: resolve reference '/foo.yml' failed: TEST ERROR", log.LastEntry().Message)
				require.Equal(t, openapi.StatusInvalid, config.Paths["/foo"].Value.Operation(http.MethodOptions).Status)
				require.NoError(t, err)
			},
		},
		{
			name: "trace",
			test: func(t *testing.T, _ *test.Hook) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, nil
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo",
						openapitest.WithOperation(http.MethodTrace),
					),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.NoError(t, err)
				path := config.Paths["/foo"].Value
				require.Equal(t, path.Trace.Path, path)
			},
		},
		{
			name: "trace error",
			test: func(t *testing.T, log *test.Hook) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, fmt.Errorf("TEST ERROR")
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithInfo("HTTP API", "", ""),
					openapitest.WithPath("/foo",
						openapitest.WithOperation(http.MethodTrace,
							openapitest.WithOperationParamRef("foo.yml"))),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.Equal(t, logrus.Fields{"method": "TRACE", "api": "HTTP API", "namespace": "http", "path": "/foo"}, log.LastEntry().Data)
				require.Equal(t, "parse parameter index '0' failed: resolve reference '/foo.yml' failed: TEST ERROR", log.LastEntry().Message)
				require.Equal(t, openapi.StatusInvalid, config.Paths["/foo"].Value.Operation(http.MethodTrace).Status)
				require.NoError(t, err)
			},
		},
		{
			name: "query",
			test: func(t *testing.T, _ *test.Hook) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, nil
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo",
						openapitest.WithOperation("QUERY"),
					),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.NoError(t, err)
				path := config.Paths["/foo"].Value
				require.Equal(t, path.Query.Path, path)
			},
		},
		{
			name: "query error",
			test: func(t *testing.T, log *test.Hook) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, fmt.Errorf("TEST ERROR")
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithInfo("HTTP API", "", ""),
					openapitest.WithPath("/foo",
						openapitest.WithOperation("QUERY",
							openapitest.WithOperationParamRef("foo.yml"))),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.Equal(t, logrus.Fields{"method": "QUERY", "api": "HTTP API", "namespace": "http", "path": "/foo"}, log.LastEntry().Data)
				require.Equal(t, "parse parameter index '0' failed: resolve reference '/foo.yml' failed: TEST ERROR", log.LastEntry().Message)
				require.Equal(t, openapi.StatusInvalid, config.Paths["/foo"].Value.Operation("QUERY").Status)
				require.NoError(t, err)
			},
		},
		{
			name: "custom LINK",
			test: func(t *testing.T, _ *test.Hook) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, nil
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo",
						openapitest.WithOperation("LINK"),
					),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.NoError(t, err)
				path := config.Paths["/foo"].Value
				require.Equal(t, path.AdditionalOperations["LINK"].Path, path)
			},
		},
		{
			name: "custom LINK error",
			test: func(t *testing.T, log *test.Hook) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, fmt.Errorf("TEST ERROR")
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithInfo("HTTP API", "", ""),
					openapitest.WithPath("/foo",
						openapitest.WithOperation("LINK",
							openapitest.WithOperationParamRef("foo.yml"))),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.Equal(t, logrus.Fields{"method": "LINK", "api": "HTTP API", "namespace": "http", "path": "/foo"}, log.LastEntry().Data)
				require.Equal(t, "parse parameter index '0' failed: resolve reference '/foo.yml' failed: TEST ERROR", log.LastEntry().Message)
				require.Equal(t, openapi.StatusInvalid, config.Paths["/foo"].Value.Operation("LINK").Status)
				require.NoError(t, err)
			},
		},
		{
			name: "request body",
			test: func(t *testing.T, _ *test.Hook) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, nil
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo",
						openapitest.UseOperation(http.MethodTrace, &openapi.Operation{
							RequestBody: &openapi.RequestBodyRef{Value: &openapi.RequestBody{Description: "foo"}},
						}),
					),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.NoError(t, err)
				path := config.Paths["/foo"].Value
				require.Equal(t, "foo", path.Trace.RequestBody.Value.Description)
			},
		},
		{
			name: "request body reference",
			test: func(t *testing.T, _ *test.Hook) {
				reader := dynamictest.ReaderFunc(func(u *url.URL, _ any) (*dynamic.Config, error) {
					cfg := &dynamic.Config{Info: dynamic.ConfigInfo{Url: u},
						Data: openapitest.NewConfig("3.0",
							openapitest.WithComponentRequestBody("foo", &openapi.RequestBody{Description: "foo"}),
						),
					}
					return cfg, nil
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo",
						openapitest.UseOperation(http.MethodTrace, &openapi.Operation{
							RequestBody: &openapi.RequestBodyRef{Reference: dynamic.Reference[*openapi.RequestBodyRef]{Ref: "foo.yml#/components/requestBodies/foo"}},
						}),
					),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.NoError(t, err)
				path := config.Paths["/foo"].Value
				require.Equal(t, "foo", path.Trace.RequestBody.Value.Description)
			},
		},
		{
			name: "request body reference error",
			test: func(t *testing.T, log *test.Hook) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, fmt.Errorf("TEST ERROR")
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithInfo("HTTP API", "", ""),
					openapitest.WithPath("/foo",
						openapitest.UseOperation(http.MethodTrace, &openapi.Operation{
							RequestBody: &openapi.RequestBodyRef{Reference: dynamic.Reference[*openapi.RequestBodyRef]{Ref: "foo.yml#/components/requestBodies/foo"}},
						}),
					),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.Equal(t, logrus.Fields{"method": "TRACE", "api": "HTTP API", "namespace": "http", "path": "/foo"}, log.LastEntry().Data)
				require.Equal(t, "parse request body failed: resolve reference '/foo.yml#/components/requestBodies/foo' failed: TEST ERROR", log.LastEntry().Message)
				require.Equal(t, openapi.StatusInvalid, config.Paths["/foo"].Value.Operation(http.MethodTrace).Status)
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

func TestConfig_Patch_Operation(t *testing.T) {
	methods := []string{"get", "post", "put", "patch", "delete", "head", "options", "trace"}
	type testType struct {
		name    string
		configs []*openapi.Config
		test    func(t *testing.T, result *openapi.Config)
	}
	var testcases []testType
	for _, m := range methods {
		m := m
		testcases = append(testcases, testType{
			name: "patch path with " + m,
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo")),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.WithOperation(m))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Len(t, result.Paths, 1)
				require.Contains(t, result.Paths, "/foo")
				require.NotNil(t, result.Paths["/foo"].Value.Operations()[strings.ToTitle(m)])
			},
		})
		testcases = append(testcases, testType{
			name: fmt.Sprintf("patch path %v info", m),
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.WithOperation(m))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.WithOperation(m,
						openapitest.WithOperationInfo("foo", "bar", "id", true),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Len(t, result.Paths, 1)
				require.Contains(t, result.Paths, "/foo")
				o := result.Paths["/foo"].Value.Operations()[strings.ToTitle(m)]
				require.Equal(t, "foo", o.Summary)
				require.Equal(t, "bar", o.Description)
				require.Equal(t, "id", o.OperationId)
				require.True(t, o.Deprecated)
			},
		})
	}

	testcases = append(testcases, testType{
		name: "patch is nil",
		configs: []*openapi.Config{
			openapitest.NewConfig("1.0", openapitest.WithPath(
				"/foo", openapitest.WithOperation("get"))),
			openapitest.NewConfig("1.0", openapitest.WithPath(
				"/foo")),
		},
		test: func(t *testing.T, result *openapi.Config) {
			require.Len(t, result.Paths, 1)
			require.Contains(t, result.Paths, "/foo")
			o := result.Paths["/foo"].Value.Get
			require.Equal(t, "", o.Summary)
		},
	})

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
