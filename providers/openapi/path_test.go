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

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestPath_UnmarshalJSON(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "paths",
			test: func(t *testing.T) {
				p := openapi.PathItems{}
				err := json.Unmarshal([]byte(`{ "/foo": {} }`), &p)
				require.NoError(t, err)
				require.Contains(t, p, "/foo")
			},
		},
		{
			name: "path value",
			test: func(t *testing.T) {
				p := &openapi.PathRef{}
				err := json.Unmarshal([]byte(`{ "summary": "foo", "description": "bar" }`), &p)
				require.NoError(t, err)
				require.Equal(t, "", p.Ref)
				require.Equal(t, "foo", p.Value.Summary)
				require.Equal(t, "bar", p.Value.Description)
			},
		},
		{
			name: "path ref",
			test: func(t *testing.T) {
				p := &openapi.PathRef{}
				err := json.Unmarshal([]byte(`{ "$ref": "#/foo/bar", "summary": "The specific item summary", "description": "The specific item description" }`), &p)
				require.NoError(t, err)
				require.Equal(t, "#/foo/bar", p.Ref)
				require.Equal(t, "The specific item summary", p.Summary)
				require.Equal(t, "The specific item description", p.Description)
				require.Nil(t, p.Value)
			},
		},
		{
			name: "path get",
			test: func(t *testing.T) {
				p := &openapi.Path{}
				err := json.Unmarshal([]byte(`{ "get": {} }`), &p)
				require.NoError(t, err)
				require.NotNil(t, p.Get)
			},
		},
		{
			name: "path post",
			test: func(t *testing.T) {
				p := &openapi.Path{}
				err := json.Unmarshal([]byte(`{ "post": {} }`), &p)
				require.NoError(t, err)
				require.NotNil(t, p.Post)
			},
		},
		{
			name: "path put",
			test: func(t *testing.T) {
				p := &openapi.Path{}
				err := json.Unmarshal([]byte(`{ "put": {} }`), &p)
				require.NoError(t, err)
				require.NotNil(t, p.Put)
			},
		},
		{
			name: "path patch",
			test: func(t *testing.T) {
				p := &openapi.Path{}
				err := json.Unmarshal([]byte(`{ "patch": {} }`), &p)
				require.NoError(t, err)
				require.NotNil(t, p.Patch)
			},
		},
		{
			name: "path delete",
			test: func(t *testing.T) {
				p := &openapi.Path{}
				err := json.Unmarshal([]byte(`{ "delete": {} }`), &p)
				require.NoError(t, err)
				require.NotNil(t, p.Delete)
			},
		},
		{
			name: "path head",
			test: func(t *testing.T) {
				p := &openapi.Path{}
				err := json.Unmarshal([]byte(`{ "head": {} }`), &p)
				require.NoError(t, err)
				require.NotNil(t, p.Head)
			},
		},
		{
			name: "path options",
			test: func(t *testing.T) {
				p := &openapi.Path{}
				err := json.Unmarshal([]byte(`{ "options": {} }`), &p)
				require.NoError(t, err)
				require.NotNil(t, p.Options)
			},
		},
		{
			name: "path trace",
			test: func(t *testing.T) {
				p := &openapi.Path{}
				err := json.Unmarshal([]byte(`{ "trace": {} }`), &p)
				require.NoError(t, err)
				require.NotNil(t, p.Trace)
			},
		},
		{
			name: "path parameters",
			test: func(t *testing.T) {
				p := &openapi.Path{}
				err := json.Unmarshal([]byte(`{ "parameters": [] }`), &p)
				require.NoError(t, err)
				require.NotNil(t, p.Parameters)
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

func TestPath_UnmarshalYAML(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "paths value",
			test: func(t *testing.T) {
				p := openapi.PathItems{}
				err := yaml.Unmarshal([]byte(`/foo: {}`), &p)
				require.NoError(t, err)
				require.Contains(t, p, "/foo")
			},
		},
		{
			name: "path value",
			test: func(t *testing.T) {
				p := &openapi.PathRef{}
				err := yaml.Unmarshal([]byte(`{ summary: foo, description: bar }`), &p)
				require.NoError(t, err)
				require.Equal(t, "", p.Ref)
				require.Equal(t, "foo", p.Value.Summary)
				require.Equal(t, "bar", p.Value.Description)
			},
		},
		{
			name: "path ref",
			test: func(t *testing.T) {
				p := &openapi.PathRef{}
				err := yaml.Unmarshal([]byte(`$ref: '#/foo/bar'`), &p)
				require.NoError(t, err)
				require.Equal(t, "#/foo/bar", p.Ref)
				require.Nil(t, p.Value)
			},
		},
		{
			name: "path get",
			test: func(t *testing.T) {
				p := &openapi.Path{}
				err := yaml.Unmarshal([]byte(`get: {}`), &p)
				require.NoError(t, err)
				require.NotNil(t, p.Get)
			},
		},
		{
			name: "path post",
			test: func(t *testing.T) {
				p := &openapi.Path{}
				err := yaml.Unmarshal([]byte(`post: {}`), &p)
				require.NoError(t, err)
				require.NotNil(t, p.Post)
			},
		},
		{
			name: "path put",
			test: func(t *testing.T) {
				p := &openapi.Path{}
				err := yaml.Unmarshal([]byte(`put: {}`), &p)
				require.NoError(t, err)
				require.NotNil(t, p.Put)
			},
		},
		{
			name: "path patch",
			test: func(t *testing.T) {
				p := &openapi.Path{}
				err := yaml.Unmarshal([]byte(`patch: {}`), &p)
				require.NoError(t, err)
				require.NotNil(t, p.Patch)
			},
		},
		{
			name: "path delete",
			test: func(t *testing.T) {
				p := &openapi.Path{}
				err := yaml.Unmarshal([]byte(`delete: {}`), &p)
				require.NoError(t, err)
				require.NotNil(t, p.Delete)
			},
		},
		{
			name: "path head",
			test: func(t *testing.T) {
				p := &openapi.Path{}
				err := yaml.Unmarshal([]byte(`head: {}`), &p)
				require.NoError(t, err)
				require.NotNil(t, p.Head)
			},
		},
		{
			name: "path options",
			test: func(t *testing.T) {
				p := &openapi.Path{}
				err := yaml.Unmarshal([]byte(`options: {}`), &p)
				require.NoError(t, err)
				require.NotNil(t, p.Options)
			},
		},
		{
			name: "path trace",
			test: func(t *testing.T) {
				p := &openapi.Path{}
				err := yaml.Unmarshal([]byte(`trace: {}`), &p)
				require.NoError(t, err)
				require.NotNil(t, p.Trace)
			},
		},
		{
			name: "path custom method",
			test: func(t *testing.T) {
				p := &openapi.Path{}
				err := yaml.Unmarshal([]byte(`additionalOperations: { LINK: {} }`), &p)
				require.NoError(t, err)
				require.NotNil(t, p.AdditionalOperations["LINK"])
			},
		},
		{
			name: "path parameters",
			test: func(t *testing.T) {
				p := &openapi.Path{}
				err := yaml.Unmarshal([]byte(`parameters: []`), &p)
				require.NoError(t, err)
				require.NotNil(t, p.Parameters)
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

func TestPath_Operations(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "get",
			test: func(t *testing.T) {
				p := &openapi.Path{Get: &openapi.Operation{}}
				ops := p.Operations()
				require.Len(t, ops, 1)
				require.Contains(t, ops, http.MethodGet)
			},
		},
		{
			name: "post",
			test: func(t *testing.T) {
				p := &openapi.Path{Post: &openapi.Operation{}}
				ops := p.Operations()
				require.Len(t, ops, 1)
				require.Contains(t, ops, http.MethodPost)
			},
		},
		{
			name: "put",
			test: func(t *testing.T) {
				p := &openapi.Path{Put: &openapi.Operation{}}
				ops := p.Operations()
				require.Len(t, ops, 1)
				require.Contains(t, ops, http.MethodPut)
			},
		},
		{
			name: "patch",
			test: func(t *testing.T) {
				p := &openapi.Path{Patch: &openapi.Operation{}}
				ops := p.Operations()
				require.Len(t, ops, 1)
				require.Contains(t, ops, http.MethodPatch)
			},
		},
		{
			name: "delete",
			test: func(t *testing.T) {
				p := &openapi.Path{Delete: &openapi.Operation{}}
				ops := p.Operations()
				require.Len(t, ops, 1)
				require.Contains(t, ops, http.MethodDelete)
			},
		},
		{
			name: "head",
			test: func(t *testing.T) {
				p := &openapi.Path{Head: &openapi.Operation{}}
				ops := p.Operations()
				require.Len(t, ops, 1)
				require.Contains(t, ops, http.MethodHead)
			},
		},
		{
			name: "options",
			test: func(t *testing.T) {
				p := &openapi.Path{Options: &openapi.Operation{}}
				ops := p.Operations()
				require.Len(t, ops, 1)
				require.Contains(t, ops, http.MethodOptions)
			},
		},
		{
			name: "trace",
			test: func(t *testing.T) {
				p := &openapi.Path{Trace: &openapi.Operation{}}
				ops := p.Operations()
				require.Len(t, ops, 1)
				require.Contains(t, ops, http.MethodTrace)
			},
		},
		{
			name: "put & trace",
			test: func(t *testing.T) {
				p := &openapi.Path{Put: &openapi.Operation{}, Trace: &openapi.Operation{}}
				ops := p.Operations()
				require.Len(t, ops, 2)
				require.Contains(t, ops, http.MethodPut)
				require.Contains(t, ops, http.MethodTrace)
			},
		},
		{
			name: "query",
			test: func(t *testing.T) {
				p := &openapi.Path{Query: &openapi.Operation{}}
				ops := p.Operations()
				require.Len(t, ops, 1)
				require.Contains(t, ops, "QUERY")
			},
		},
		{
			name: "custom LINK",
			test: func(t *testing.T) {
				p := &openapi.Path{AdditionalOperations: map[string]*openapi.Operation{"LINK": {}}}
				ops := p.Operations()
				require.Len(t, ops, 1)
				require.Contains(t, ops, "LINK")
			},
		},
		{
			name: "get operation custom LINK",
			test: func(t *testing.T) {
				p := &openapi.Path{AdditionalOperations: map[string]*openapi.Operation{"LINK": {}}}
				op := p.Operation("link")
				require.NotNil(t, op)
			},
		},
		{
			name: "get operation but not defined",
			test: func(t *testing.T) {
				p := &openapi.Path{AdditionalOperations: map[string]*openapi.Operation{"LINK": {}}}
				op := p.Operation("not")
				require.Nil(t, op)
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

func TestPath_Parse(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "reader returns error",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, fmt.Errorf("TEST ERROR")
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithPathRef("foo",
						&openapi.PathRef{Reference: dynamic.Reference{Ref: "foo.yml#/paths/foo"}}))
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.EqualError(t, err, "parse path 'foo' failed: resolve reference 'foo.yml#/paths/foo' failed: TEST ERROR")
			},
		},
		{
			name: "path is nil",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, nil
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithPathRef("foo", nil))
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.NoError(t, err)
			},
		},
		{
			name: "reference to a file",
			test: func(t *testing.T) {
				target := &openapi.Path{}
				reader := dynamictest.ReaderFunc(func(u *url.URL, _ any) (*dynamic.Config, error) {
					require.Equal(t, "/foo.yml", u.String())
					cfg := &dynamic.Config{Info: dynamic.ConfigInfo{Url: u},
						Data: openapitest.NewConfig("3.0",
							openapitest.WithPath("/foo", target)),
					}
					return cfg, nil
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithPathRef("/foo",
						&openapi.PathRef{Reference: dynamic.Reference{Ref: "foo.yml#/paths/foo"}}))
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.NoError(t, err)
				require.Equal(t, target, config.Paths["/foo"].Value)
			},
		},
		{
			name: "file reference but path is nil",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(u *url.URL, _ any) (*dynamic.Config, error) {
					require.Equal(t, "/foo.yml", u.String())
					cfg := &dynamic.Config{Info: dynamic.ConfigInfo{Url: u},
						Data: openapitest.NewConfig("3.0",
							openapitest.WithPath("/foo", nil)),
					}
					return cfg, nil
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithPathRef("/foo",
						&openapi.PathRef{Reference: dynamic.Reference{Ref: "foo.yml#/paths/foo"}}))
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.NoError(t, err)
				require.Nil(t, config.Paths["/foo"].Value)
			},
		},
		{
			name: "file reference but local reference not found",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(u *url.URL, _ any) (*dynamic.Config, error) {
					require.Equal(t, "/foo.yml", u.String())
					config := openapitest.NewConfig("3.0")
					config.Paths = openapi.PathItems{}
					cfg := &dynamic.Config{Info: dynamic.ConfigInfo{Url: u},
						Data: config,
					}
					return cfg, nil
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithPathRef("/foo",
						&openapi.PathRef{Reference: dynamic.Reference{Ref: "foo.yml#/paths/foo"}}))
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.EqualError(t, err, "parse path '/foo' failed: resolve reference 'foo.yml#/paths/foo' failed: value is null")
			},
		},
		{
			name: "parameters with error",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, fmt.Errorf("TEST ERROR")
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo", openapitest.NewPath(
						openapitest.WithPathParamRef(&openapi.ParameterRef{Reference: dynamic.Reference{Ref: "foo.yml"}}),
					)),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.EqualError(t, err, "parse path '/foo' failed: resolve reference 'foo.yml' failed: TEST ERROR")
			},
		},
		{
			name: "path as reference",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, fmt.Errorf("TEST ERROR")
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithPathRef("foo",
						&openapi.PathRef{Reference: dynamic.Reference{Ref: "#/components/pathItems/foo"}},
					),
					openapitest.WithComponentPathItem("foo", openapitest.NewPath()),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.NoError(t, err)
				require.NotNil(t, config.Paths["foo"].Value)
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

func TestConfig_Patch_Path(t *testing.T) {
	testcases := []struct {
		name    string
		configs []*openapi.Config
		test    func(t *testing.T, result *openapi.Config)
	}{
		{
			name: "patch both path are nil",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", nil,
				)),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", nil,
				)),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Len(t, result.Paths, 1)
				require.Contains(t, result.Paths, "/foo")
				require.Nil(t, result.Paths["/foo"].Value)
			},
		},
		{
			name: "left path ref is nil",
			configs: []*openapi.Config{
				{Paths: map[string]*openapi.PathRef{"/foo": nil}},
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(
						openapitest.WithOperation(
							"post", openapitest.NewOperation(),
						),
					),
				)),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Len(t, result.Paths, 1)
				require.Contains(t, result.Paths, "/foo")
				require.NotNil(t, result.Paths["/foo"].Value.Post)
			},
		},
		{
			name: "patch without path",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(),
					),
					))),
				openapitest.NewConfig("1.0"),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Len(t, result.Paths, 1)
				require.Contains(t, result.Paths, "/foo")
				require.NotNil(t, result.Paths["/foo"].Value.Post)
			},
		},
		{
			name: "patch path",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0"),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Len(t, result.Paths, 1)
				require.Contains(t, result.Paths, "/foo")
				require.NotNil(t, result.Paths["/foo"].Value.Post)
			},
		},
		{
			name: "patch path original not resolved",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPathRef(
					"/foo", &openapi.PathRef{})),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Len(t, result.Paths, 1)
				require.Contains(t, result.Paths, "/foo")
				require.NotNil(t, result.Paths["/foo"].Value.Post)
			},
		},
		{
			name: "patch summary and description",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath())),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithPathInfo("foo", "bar")))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Len(t, result.Paths, 1)
				require.Contains(t, result.Paths, "/foo")
				require.Equal(t, "foo", result.Paths["/foo"].Value.Summary)
				require.Equal(t, "bar", result.Paths["/foo"].Value.Description)
			},
		},
		{
			name: "add parameters",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath())),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithPathParam("foo")))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				e := result.Paths["/foo"].Value
				require.Len(t, e.Parameters, 1)
			},
		},
		{
			name: "patch parameters",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithPathParam("foo")))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithPathParam("foo", openapitest.WithParamSchema(schematest.New("number")))))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				e := result.Paths["/foo"].Value
				require.Len(t, e.Parameters, 1)
				require.Equal(t, "number", e.Parameters[0].Value.Schema.Type.String())
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
