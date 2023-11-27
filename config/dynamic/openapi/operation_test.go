package openapi_test

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/openapi"
	"mokapi/config/dynamic/openapi/openapitest"
	"mokapi/config/dynamic/openapi/parameter"
	"mokapi/config/dynamic/openapi/ref"
	"net/http"
	"net/url"
	"strings"
	"testing"
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
		test func(t *testing.T)
	}{
		{
			name: "get",
			test: func(t *testing.T) {
				reader := &testReader{readFunc: func(cfg *common.Config) error {
					return nil
				}}
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo", openapitest.NewPath(
						openapitest.WithOperation(http.MethodGet, &openapi.Operation{}),
					)),
				)
				err := config.Parse(common.NewConfig(&url.URL{}, common.WithData(config)), reader)
				require.NoError(t, err)
				path := config.Paths["/foo"].Value
				require.Equal(t, path.Get.Path, path)
			},
		},
		{
			name: "get error",
			test: func(t *testing.T) {
				reader := &testReader{readFunc: func(cfg *common.Config) error {
					return fmt.Errorf("TEST ERROR")
				}}
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo", openapitest.NewPath(
						openapitest.WithOperation(http.MethodGet,
							openapitest.NewOperation(
								openapitest.WithOperationParamRef(&parameter.Ref{Reference: ref.Reference{Ref: "foo.yml"}}))),
					)),
				)
				err := config.Parse(common.NewConfig(&url.URL{}, common.WithData(config)), reader)
				require.EqualError(t, err, "parse path '/foo' failed: parse operation 'GET' failed: parse parameter index '0' failed: resolve reference 'foo.yml' failed: TEST ERROR")
			},
		},
		{
			name: "post",
			test: func(t *testing.T) {
				reader := &testReader{readFunc: func(cfg *common.Config) error {
					return nil
				}}
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo", openapitest.NewPath(
						openapitest.WithOperation(http.MethodPost, &openapi.Operation{}),
					)),
				)
				err := config.Parse(common.NewConfig(&url.URL{}, common.WithData(config)), reader)
				require.NoError(t, err)
				path := config.Paths["/foo"].Value
				require.Equal(t, path.Post.Path, path)
			},
		},
		{
			name: "post error",
			test: func(t *testing.T) {
				reader := &testReader{readFunc: func(cfg *common.Config) error {
					return fmt.Errorf("TEST ERROR")
				}}
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo", openapitest.NewPath(
						openapitest.WithOperation(http.MethodPost,
							openapitest.NewOperation(
								openapitest.WithOperationParamRef(&parameter.Ref{Reference: ref.Reference{Ref: "foo.yml"}}))),
					)),
				)
				err := config.Parse(common.NewConfig(&url.URL{}, common.WithData(config)), reader)
				require.EqualError(t, err, "parse path '/foo' failed: parse operation 'POST' failed: parse parameter index '0' failed: resolve reference 'foo.yml' failed: TEST ERROR")
			},
		},
		{
			name: "put",
			test: func(t *testing.T) {
				reader := &testReader{readFunc: func(cfg *common.Config) error {
					return nil
				}}
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo", openapitest.NewPath(
						openapitest.WithOperation(http.MethodPut, &openapi.Operation{}),
					)),
				)
				err := config.Parse(common.NewConfig(&url.URL{}, common.WithData(config)), reader)
				require.NoError(t, err)
				path := config.Paths["/foo"].Value
				require.Equal(t, path.Put.Path, path)
			},
		},
		{
			name: "put error",
			test: func(t *testing.T) {
				reader := &testReader{readFunc: func(cfg *common.Config) error {
					return fmt.Errorf("TEST ERROR")
				}}
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo", openapitest.NewPath(
						openapitest.WithOperation(http.MethodPut,
							openapitest.NewOperation(
								openapitest.WithOperationParamRef(&parameter.Ref{Reference: ref.Reference{Ref: "foo.yml"}}))),
					)),
				)
				err := config.Parse(common.NewConfig(&url.URL{}, common.WithData(config)), reader)
				require.EqualError(t, err, "parse path '/foo' failed: parse operation 'PUT' failed: parse parameter index '0' failed: resolve reference 'foo.yml' failed: TEST ERROR")
			},
		},
		{
			name: "patch",
			test: func(t *testing.T) {
				reader := &testReader{readFunc: func(cfg *common.Config) error {
					return nil
				}}
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo", openapitest.NewPath(
						openapitest.WithOperation(http.MethodPatch, &openapi.Operation{}),
					)),
				)
				err := config.Parse(common.NewConfig(&url.URL{}, common.WithData(config)), reader)
				require.NoError(t, err)
				path := config.Paths["/foo"].Value
				require.Equal(t, path.Patch.Path, path)
			},
		},
		{
			name: "patch error",
			test: func(t *testing.T) {
				reader := &testReader{readFunc: func(cfg *common.Config) error {
					return fmt.Errorf("TEST ERROR")
				}}
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo", openapitest.NewPath(
						openapitest.WithOperation(http.MethodPatch,
							openapitest.NewOperation(
								openapitest.WithOperationParamRef(&parameter.Ref{Reference: ref.Reference{Ref: "foo.yml"}}))),
					)),
				)
				err := config.Parse(common.NewConfig(&url.URL{}, common.WithData(config)), reader)
				require.EqualError(t, err, "parse path '/foo' failed: parse operation 'PATCH' failed: parse parameter index '0' failed: resolve reference 'foo.yml' failed: TEST ERROR")
			},
		},
		{
			name: "delete",
			test: func(t *testing.T) {
				reader := &testReader{readFunc: func(cfg *common.Config) error {
					return nil
				}}
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo", openapitest.NewPath(
						openapitest.WithOperation(http.MethodDelete, &openapi.Operation{}),
					)),
				)
				err := config.Parse(common.NewConfig(&url.URL{}, common.WithData(config)), reader)
				require.NoError(t, err)
				path := config.Paths["/foo"].Value
				require.Equal(t, path.Delete.Path, path)
			},
		},
		{
			name: "delete error",
			test: func(t *testing.T) {
				reader := &testReader{readFunc: func(cfg *common.Config) error {
					return fmt.Errorf("TEST ERROR")
				}}
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo", openapitest.NewPath(
						openapitest.WithOperation(http.MethodDelete,
							openapitest.NewOperation(
								openapitest.WithOperationParamRef(&parameter.Ref{Reference: ref.Reference{Ref: "foo.yml"}}))),
					)),
				)
				err := config.Parse(common.NewConfig(&url.URL{}, common.WithData(config)), reader)
				require.EqualError(t, err, "parse path '/foo' failed: parse operation 'DELETE' failed: parse parameter index '0' failed: resolve reference 'foo.yml' failed: TEST ERROR")
			},
		},
		{
			name: "head",
			test: func(t *testing.T) {
				reader := &testReader{readFunc: func(cfg *common.Config) error {
					return nil
				}}
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo", openapitest.NewPath(
						openapitest.WithOperation(http.MethodHead, &openapi.Operation{}),
					)),
				)
				err := config.Parse(common.NewConfig(&url.URL{}, common.WithData(config)), reader)
				require.NoError(t, err)
				path := config.Paths["/foo"].Value
				require.Equal(t, path.Head.Path, path)
			},
		},
		{
			name: "head error",
			test: func(t *testing.T) {
				reader := &testReader{readFunc: func(cfg *common.Config) error {
					return fmt.Errorf("TEST ERROR")
				}}
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo", openapitest.NewPath(
						openapitest.WithOperation(http.MethodHead,
							openapitest.NewOperation(
								openapitest.WithOperationParamRef(&parameter.Ref{Reference: ref.Reference{Ref: "foo.yml"}}))),
					)),
				)
				err := config.Parse(common.NewConfig(&url.URL{}, common.WithData(config)), reader)
				require.EqualError(t, err, "parse path '/foo' failed: parse operation 'HEAD' failed: parse parameter index '0' failed: resolve reference 'foo.yml' failed: TEST ERROR")
			},
		},
		{
			name: "options",
			test: func(t *testing.T) {
				reader := &testReader{readFunc: func(cfg *common.Config) error {
					return nil
				}}
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo", openapitest.NewPath(
						openapitest.WithOperation(http.MethodOptions, &openapi.Operation{}),
					)),
				)
				err := config.Parse(common.NewConfig(&url.URL{}, common.WithData(config)), reader)
				require.NoError(t, err)
				path := config.Paths["/foo"].Value
				require.Equal(t, path.Options.Path, path)
			},
		},
		{
			name: "options error",
			test: func(t *testing.T) {
				reader := &testReader{readFunc: func(cfg *common.Config) error {
					return fmt.Errorf("TEST ERROR")
				}}
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo", openapitest.NewPath(
						openapitest.WithOperation(http.MethodOptions,
							openapitest.NewOperation(
								openapitest.WithOperationParamRef(&parameter.Ref{Reference: ref.Reference{Ref: "foo.yml"}}))),
					)),
				)
				err := config.Parse(common.NewConfig(&url.URL{}, common.WithData(config)), reader)
				require.EqualError(t, err, "parse path '/foo' failed: parse operation 'OPTIONS' failed: parse parameter index '0' failed: resolve reference 'foo.yml' failed: TEST ERROR")
			},
		},
		{
			name: "trace",
			test: func(t *testing.T) {
				reader := &testReader{readFunc: func(cfg *common.Config) error {
					return nil
				}}
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo", openapitest.NewPath(
						openapitest.WithOperation(http.MethodTrace, &openapi.Operation{}),
					)),
				)
				err := config.Parse(common.NewConfig(&url.URL{}, common.WithData(config)), reader)
				require.NoError(t, err)
				path := config.Paths["/foo"].Value
				require.Equal(t, path.Trace.Path, path)
			},
		},
		{
			name: "trace error",
			test: func(t *testing.T) {
				reader := &testReader{readFunc: func(cfg *common.Config) error {
					return fmt.Errorf("TEST ERROR")
				}}
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo", openapitest.NewPath(
						openapitest.WithOperation(http.MethodTrace,
							openapitest.NewOperation(
								openapitest.WithOperationParamRef(&parameter.Ref{Reference: ref.Reference{Ref: "foo.yml"}}))),
					)),
				)
				err := config.Parse(common.NewConfig(&url.URL{}, common.WithData(config)), reader)
				require.EqualError(t, err, "parse path '/foo' failed: parse operation 'TRACE' failed: parse parameter index '0' failed: resolve reference 'foo.yml' failed: TEST ERROR")
			},
		},
		{
			name: "request body",
			test: func(t *testing.T) {
				reader := &testReader{readFunc: func(cfg *common.Config) error {
					return nil
				}}
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo", openapitest.NewPath(
						openapitest.WithOperation(http.MethodTrace, &openapi.Operation{
							RequestBody: &openapi.RequestBodyRef{Value: &openapi.RequestBody{Description: "foo"}},
						}),
					)),
				)
				err := config.Parse(common.NewConfig(&url.URL{}, common.WithData(config)), reader)
				require.NoError(t, err)
				path := config.Paths["/foo"].Value
				require.Equal(t, "foo", path.Trace.RequestBody.Value.Description)
			},
		},
		{
			name: "request body reference",
			test: func(t *testing.T) {
				reader := &testReader{readFunc: func(cfg *common.Config) error {
					cfg.Data = openapitest.NewConfig("3.0",
						openapitest.WithComponentRequestBody("foo", &openapi.RequestBody{Description: "foo"}),
					)
					return nil
				}}
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo", openapitest.NewPath(
						openapitest.WithOperation(http.MethodTrace, &openapi.Operation{
							RequestBody: &openapi.RequestBodyRef{Reference: ref.Reference{Ref: "foo.yml#/components/requestBodies/foo"}},
						}),
					)),
				)
				err := config.Parse(common.NewConfig(&url.URL{}, common.WithData(config)), reader)
				require.NoError(t, err)
				path := config.Paths["/foo"].Value
				require.Equal(t, "foo", path.Trace.RequestBody.Value.Description)
			},
		},
		{
			name: "request body reference error",
			test: func(t *testing.T) {
				reader := &testReader{readFunc: func(cfg *common.Config) error {
					return fmt.Errorf("TEST ERROR")
				}}
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo", openapitest.NewPath(
						openapitest.WithOperation(http.MethodTrace, &openapi.Operation{
							RequestBody: &openapi.RequestBodyRef{Reference: ref.Reference{Ref: "foo.yml#/components/requestBodies/foo"}},
						}),
					)),
				)
				err := config.Parse(common.NewConfig(&url.URL{}, common.WithData(config)), reader)
				require.EqualError(t, err, "parse path '/foo' failed: parse operation 'TRACE' failed: parse request body failed: resolve reference 'foo.yml#/components/requestBodies/foo' failed: TEST ERROR")
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
					"/foo", openapitest.NewPath())),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(m, openapitest.NewOperation())))),
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
					"/foo", openapitest.NewPath(openapitest.WithOperation(m, openapitest.NewOperation())))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(m, openapitest.NewOperation(
						openapitest.WithOperationInfo("foo", "bar", "id", true),
					))))),
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
				"/foo", openapitest.NewPath(openapitest.WithOperation("get", openapitest.NewOperation())))),
			openapitest.NewConfig("1.0", openapitest.WithPath(
				"/foo", openapitest.NewPath())),
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
