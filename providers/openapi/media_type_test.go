package openapi_test

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	"mokapi/providers/openapi/schema/schematest"
	"net/http"
	"net/url"
	"testing"
)

func TestMediaType_UnmarshalJSON(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "simple",
			test: func(t *testing.T) {
				mt := &openapi.MediaType{}
				err := json.Unmarshal([]byte(`{ "schema": {"type": "string"}, "example": "foo", "examples": {"foo": {"summary": "foo example"}} }`), &mt)
				require.NoError(t, err)
				require.Equal(t, "string", mt.Schema.Type.String())
				require.Equal(t, "foo", mt.Example)
				require.Equal(t, "foo example", mt.Examples["foo"].Value.Summary)
				require.True(t, mt.ContentType.IsEmpty())
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

func TestMediaType_UnmarshalYAML(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "simple",
			test: func(t *testing.T) {
				mt := &openapi.MediaType{}
				err := yaml.Unmarshal([]byte(`
schema: 
  type: string
example: foo
examples: 
  foo: 
    summary: foo example`), &mt)
				require.NoError(t, err)
				require.Equal(t, "string", mt.Schema.Type.String())
				require.Equal(t, "foo", mt.Example)
				require.Equal(t, "foo example", mt.Examples["foo"].Value.Summary)
				require.True(t, mt.ContentType.IsEmpty())
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

func TestMediaType_Parse(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "no error",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, nil
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo", openapitest.NewPath(
						openapitest.WithOperation(http.MethodGet, openapitest.NewOperation(
							openapitest.WithResponse(http.StatusOK,
								openapitest.WithContent("application/json", &openapi.MediaType{}),
							))),
					)),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.NoError(t, err)
			},
		},
		{
			name: "MediaType is nil",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, nil
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo", openapitest.NewPath(
						openapitest.WithOperation(http.MethodGet, openapitest.NewOperation(
							openapitest.WithResponse(http.StatusOK,
								openapitest.WithContent("application/json", nil),
							))),
					)),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.NoError(t, err)
			},
		},
		{
			name: "error by resolving schema ref",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, fmt.Errorf("TEST ERROR")
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo", openapitest.NewPath(
						openapitest.WithOperation(http.MethodGet, openapitest.NewOperation(
							openapitest.WithResponse(http.StatusOK,
								openapitest.WithContent("application/json",
									openapitest.NewContent(
										openapitest.WithSchemaRef("foo.yml"),
									),
								),
							),
						),
						),
					)),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.EqualError(t, err, "parse path '/foo' failed: parse operation 'GET' failed: parse response '200' failed: parse content 'application/json' failed: parse schema failed: resolve reference 'foo.yml' failed: TEST ERROR")
			},
		},
		{
			name: "error by resolving example ref",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, fmt.Errorf("TEST ERROR")
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo", openapitest.NewPath(
						openapitest.WithOperation(http.MethodGet, openapitest.NewOperation(
							openapitest.WithResponse(http.StatusOK,
								openapitest.WithContent("application/json",
									&openapi.MediaType{Examples: map[string]*openapi.ExampleRef{"foo": {Reference: dynamic.Reference{Ref: "foo.yml"}}}},
								),
							),
						),
						),
					)),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.EqualError(t, err, "parse path '/foo' failed: parse operation 'GET' failed: parse response '200' failed: parse content 'application/json' failed: parse example 'foo' failed: resolve reference 'foo.yml' failed: TEST ERROR")
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

func TestConfig_Patch_MediaType(t *testing.T) {
	testcases := []struct {
		name    string
		configs []*openapi.Config
		test    func(t *testing.T, result *openapi.Config)
	}{
		{
			name: "patch is nil",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/plain", &openapi.MediaType{})),
						),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/html", nil)),
						),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				res := result.Paths["/foo"].Value.Post.Responses.GetResponse(200)
				require.Len(t, res.Content, 2)
				require.NotNil(t, res.Content["text/plain"])
				require.Nil(t, res.Content["text/html"])
			},
		},
		{
			name: "add schema",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/html", &openapi.MediaType{})),
						),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/html", &openapi.MediaType{Schema: schematest.New("string")})),
						),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				res := result.Paths["/foo"].Value.Post.Responses.GetResponse(200)
				mt := res.Content["text/html"]
				require.Equal(t, "string", mt.Schema.Type.String())
			},
		},
		{
			name: "patch schema",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/html", &openapi.MediaType{Schema: schematest.New("string")})),
						),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/html", &openapi.MediaType{Schema: schematest.New("object")})),
						),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				res := result.Paths["/foo"].Value.Post.Responses.GetResponse(200)
				mt := res.Content["text/html"]
				require.Equal(t, "[string, object]", mt.Schema.Type.String())
			},
		},
		{
			name: "add example",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/html", &openapi.MediaType{})),
						),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/html", &openapi.MediaType{Example: "foo"})),
						),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				res := result.Paths["/foo"].Value.Post.Responses.GetResponse(200)
				mt := res.Content["text/html"]
				require.Equal(t, "foo", mt.Example)
			},
		},
		{
			name: "patch example",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/html", &openapi.MediaType{Example: "foo"})),
						),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/html", &openapi.MediaType{Example: "bar"})),
						),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				res := result.Paths["/foo"].Value.Post.Responses.GetResponse(200)
				mt := res.Content["text/html"]
				require.Equal(t, "bar", mt.Example)
			},
		},
		{
			name: "patch example is nil",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/html", &openapi.MediaType{Example: "foo"})),
						),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/html", &openapi.MediaType{Example: nil})),
						),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				res := result.Paths["/foo"].Value.Post.Responses.GetResponse(200)
				mt := res.Content["text/html"]
				require.Equal(t, "foo", mt.Example)
			},
		},
		{
			name: "add examples",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/html", &openapi.MediaType{})),
						),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/html", &openapi.MediaType{
								Examples: map[string]*openapi.ExampleRef{"foo": {Value: &openapi.Example{Value: "foo"}}},
							})),
						),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				res := result.Paths["/foo"].Value.Post.Responses.GetResponse(200)
				mt := res.Content["text/html"]
				require.Equal(t, "foo", mt.Examples["foo"].Value.Value)
			},
		},
		{
			name: "patch examples foo",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/html", &openapi.MediaType{
								Examples: map[string]*openapi.ExampleRef{"foo": {Value: &openapi.Example{Value: "foo"}}},
							})),
						),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/html", &openapi.MediaType{
								Examples: map[string]*openapi.ExampleRef{"foo": {Value: &openapi.Example{Value: "bar"}}},
							})),
						),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				res := result.Paths["/foo"].Value.Post.Responses.GetResponse(200)
				mt := res.Content["text/html"]
				require.Equal(t, "bar", mt.Examples["foo"].Value.Value)
			},
		},
		{
			name: "patch example is nil",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/html", &openapi.MediaType{
								Examples: map[string]*openapi.ExampleRef{"foo": {Value: &openapi.Example{Value: "foo"}}},
							})),
						),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/html", &openapi.MediaType{Examples: nil})),
						),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				res := result.Paths["/foo"].Value.Post.Responses.GetResponse(200)
				mt := res.Content["text/html"]
				require.Equal(t, "foo", mt.Examples["foo"].Value.Value)
			},
		},
		{
			name: "patch examples overwrites example",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/html", &openapi.MediaType{
								Example: "foo",
							})),
						),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/html", &openapi.MediaType{
								Examples: map[string]*openapi.ExampleRef{"foo": {Value: &openapi.Example{Value: "bar"}}},
							})),
						),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				res := result.Paths["/foo"].Value.Post.Responses.GetResponse(200)
				mt := res.Content["text/html"]
				require.Nil(t, mt.Example)
				require.Equal(t, "bar", mt.Examples["foo"].Value.Value)
			},
		},
		{
			name: "patch example overwrites examples",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/html", &openapi.MediaType{
								Examples: map[string]*openapi.ExampleRef{"foo": {Value: &openapi.Example{Value: "foo"}}},
							})),
						),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/html", &openapi.MediaType{
								Example: "bar",
							})),
						),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				res := result.Paths["/foo"].Value.Post.Responses.GetResponse(200)
				mt := res.Content["text/html"]
				require.Len(t, mt.Examples, 0)
				require.Equal(t, "bar", mt.Example)
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
