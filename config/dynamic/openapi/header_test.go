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
	"mokapi/config/dynamic/openapi/schema"
	"net/http"
	"net/url"
	"testing"
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
				require.Equal(t, parameter.Header, header.Type)
			},
		},
		{
			name: "overwrite type",
			test: func(t *testing.T) {
				header := &openapi.Header{}
				err := json.Unmarshal([]byte(`{ "in": "cookie" }`), &header)
				require.NoError(t, err)
				require.Equal(t, parameter.Header, header.Type)
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
				require.Equal(t, parameter.Header, header.Type)
			},
		},
		{
			name: "overwrite type",
			test: func(t *testing.T) {
				header := &openapi.Header{}
				err := yaml.Unmarshal([]byte(`in: cookie }`), &header)
				require.NoError(t, err)
				require.Equal(t, parameter.Header, header.Type)
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
		test func(t *testing.T)
	}{
		{
			name: "reference is nil",
			test: func(t *testing.T) {
				reader := &testReader{readFunc: func(cfg *common.Config) error {
					return nil
				}}
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo", openapitest.NewPath(
						openapitest.WithOperation(http.MethodGet, openapitest.NewOperation(
							openapitest.WithResponse(http.StatusOK,
								openapitest.WithResponseHeaderRef("foo", nil),
							))),
					)),
				)
				err := config.Parse(common.NewConfig(common.ConfigInfo{Url: &url.URL{}}, common.WithData(config)), reader)
				require.NoError(t, err)
			},
		},
		{
			name: "header",
			test: func(t *testing.T) {
				reader := &testReader{readFunc: func(cfg *common.Config) error {
					return nil
				}}
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo", openapitest.NewPath(
						openapitest.WithOperation(http.MethodGet, openapitest.NewOperation(
							openapitest.WithResponse(http.StatusOK,
								openapitest.WithResponseHeader("foo", "foo description", &schema.Schema{Type: "string"}),
							))),
					)),
				)
				err := config.Parse(common.NewConfig(common.ConfigInfo{Url: &url.URL{}}, common.WithData(config)), reader)
				require.NoError(t, err)
				h := config.Paths["/foo"].Value.Get.Responses.GetResponse(http.StatusOK).Headers["foo"]
				require.Equal(t, "foo description", h.Value.Description)
			},
		},
		{
			name: "error by resolving example ref",
			test: func(t *testing.T) {
				reader := &testReader{readFunc: func(cfg *common.Config) error {
					return fmt.Errorf("TEST ERROR")
				}}
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo", openapitest.NewPath(
						openapitest.WithOperation(http.MethodGet, openapitest.NewOperation(
							openapitest.WithResponse(http.StatusOK,
								openapitest.WithResponseHeaderRef("foo", &openapi.HeaderRef{Reference: ref.Reference{Ref: "foo"}}),
							))),
					)),
				)
				err := config.Parse(common.NewConfig(common.ConfigInfo{Url: &url.URL{}}, common.WithData(config)), reader)
				require.EqualError(t, err, "parse path '/foo' failed: parse operation 'GET' failed: parse response '200' failed: parse header 'foo' failed: resolve reference 'foo' failed: TEST ERROR")
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
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponseRef(200, &openapi.ResponseRef{Value: &openapi.Response{
								Headers: map[string]*openapi.HeaderRef{},
							}}),
						),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponseRef(200, &openapi.ResponseRef{Value: &openapi.Response{
								Headers: map[string]*openapi.HeaderRef{
									"foo": {Value: &openapi.Header{Parameter: parameter.Parameter{Description: "foo"}}},
								},
							}}),
						),
					),
					))),
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
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponseRef(200, &openapi.ResponseRef{Value: &openapi.Response{
								Headers: map[string]*openapi.HeaderRef{
									"foo": {Value: &openapi.Header{Parameter: parameter.Parameter{Description: "foo"}}},
								},
							}}),
						),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponseRef(200, &openapi.ResponseRef{Value: &openapi.Response{
								Headers: map[string]*openapi.HeaderRef{
									"foo": {Value: &openapi.Header{Parameter: parameter.Parameter{Description: "bar"}}},
								},
							}}),
						),
					),
					))),
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
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponseRef(200, &openapi.ResponseRef{Value: &openapi.Response{
								Headers: map[string]*openapi.HeaderRef{
									"foo": {Value: nil},
								},
							}}),
						),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponseRef(200, &openapi.ResponseRef{Value: &openapi.Response{
								Headers: map[string]*openapi.HeaderRef{
									"foo": {Value: &openapi.Header{Parameter: parameter.Parameter{Description: "foo"}}},
								},
							}}),
						),
					),
					))),
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
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponseRef(200, &openapi.ResponseRef{Value: &openapi.Response{
								Headers: map[string]*openapi.HeaderRef{
									"foo": {Value: nil},
								},
							}}),
						),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponseRef(200, &openapi.ResponseRef{Value: &openapi.Response{
								Headers: map[string]*openapi.HeaderRef{
									"foo": {Value: &openapi.Header{Parameter: parameter.Parameter{Description: "foo"}}},
								},
							}}),
						),
					),
					))),
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
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponseRef(200, &openapi.ResponseRef{Value: &openapi.Response{
								Headers: map[string]*openapi.HeaderRef{
									"foo": {Value: &openapi.Header{Parameter: parameter.Parameter{Description: "foo"}}},
								},
							}}),
						),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponseRef(200, &openapi.ResponseRef{Value: &openapi.Response{
								Headers: map[string]*openapi.HeaderRef{
									"foo": nil,
								},
							}}),
						),
					),
					))),
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
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponseRef(200, &openapi.ResponseRef{Value: &openapi.Response{
								Headers: map[string]*openapi.HeaderRef{
									"foo": {Value: &openapi.Header{Parameter: parameter.Parameter{Description: "foo"}}},
								},
							}}),
						),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponseRef(200, &openapi.ResponseRef{Value: &openapi.Response{
								Headers: map[string]*openapi.HeaderRef{
									"foo": {},
								},
							}}),
						),
					),
					))),
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
