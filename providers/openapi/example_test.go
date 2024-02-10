package openapi_test

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/json/ref"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	"net/http"
	"net/url"
	"testing"
)

func TestExamples_UnmarshalJSON(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "one example",
			test: func(t *testing.T) {
				examples := openapi.Examples{}
				err := json.Unmarshal([]byte(`
{ "foo": { "summary": "foo summary", "value": { "foo": "bar" }, "description": "foo description", "externalValue": "https://foo.bar" } }`), &examples)
				require.NoError(t, err)
				require.Len(t, examples, 1)
				require.Contains(t, examples, "foo")
				foo := examples["foo"]
				require.Equal(t, "foo summary", foo.Value.Summary)
				require.Equal(t, "foo description", foo.Value.Description)
				require.Equal(t, map[string]interface{}{"foo": "bar"}, foo.Value.Value)
				require.Equal(t, "https://foo.bar", foo.Value.ExternalValue)
			},
		},
		{
			name: "two example",
			test: func(t *testing.T) {
				examples := openapi.Examples{}
				err := json.Unmarshal([]byte(`{
"foo": { "summary": "foo summary", "value": { "foo": "bar" }, "description": "foo description" },
"bar": { "summary": "bar summary", "value": { "bar": "baz" }, "description": "bar description" }
}`), &examples)
				require.NoError(t, err)
				require.Len(t, examples, 2)
				require.Contains(t, examples, "foo")
				require.Contains(t, examples, "bar")
				foo := examples["foo"]
				require.Equal(t, "foo summary", foo.Value.Summary)
				require.Equal(t, "foo description", foo.Value.Description)
				require.Equal(t, map[string]interface{}{"foo": "bar"}, foo.Value.Value)
				bar := examples["bar"]
				require.Equal(t, "bar summary", bar.Value.Summary)
				require.Equal(t, "bar description", bar.Value.Description)
				require.Equal(t, map[string]interface{}{"bar": "baz"}, bar.Value.Value)
			},
		},
		{
			name: "$ref",
			test: func(t *testing.T) {
				examples := openapi.Examples{}
				err := json.Unmarshal([]byte(`{ "foo": { "$ref": "foo.yml" } }`), &examples)
				require.NoError(t, err)
				require.Equal(t, "foo.yml", examples["foo"].Ref)
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

func TestExamples_UnmarshalYAML(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "one example",
			test: func(t *testing.T) {
				examples := openapi.Examples{}
				err := yaml.Unmarshal([]byte(`
foo: { summary: foo summary, value: { foo: bar }, description: foo description, externalValue: https://foo.bar }`), &examples)
				require.NoError(t, err)
				require.Len(t, examples, 1)
				require.Contains(t, examples, "foo")
				foo := examples["foo"]
				require.Equal(t, "foo summary", foo.Value.Summary)
				require.Equal(t, "foo description", foo.Value.Description)
				require.Equal(t, map[string]interface{}{"foo": "bar"}, foo.Value.Value)
				require.Equal(t, "https://foo.bar", foo.Value.ExternalValue)
			},
		},
		{
			name: "two example",
			test: func(t *testing.T) {
				examples := openapi.Examples{}
				err := yaml.Unmarshal([]byte(`
foo: { summary: foo summary, value: { foo: bar }, description: foo description }
bar: { summary: bar summary, value: { bar: baz }, description: bar description }
`), &examples)
				require.NoError(t, err)
				require.Len(t, examples, 2)
				require.Contains(t, examples, "foo")
				require.Contains(t, examples, "bar")
				foo := examples["foo"]
				require.Equal(t, "foo summary", foo.Value.Summary)
				require.Equal(t, "foo description", foo.Value.Description)
				require.Equal(t, map[string]interface{}{"foo": "bar"}, foo.Value.Value)
				bar := examples["bar"]
				require.Equal(t, "bar summary", bar.Value.Summary)
				require.Equal(t, "bar description", bar.Value.Description)
				require.Equal(t, map[string]interface{}{"bar": "baz"}, bar.Value.Value)
			},
		},
		{
			name: "$ref",
			test: func(t *testing.T) {
				examples := openapi.Examples{}
				err := yaml.Unmarshal([]byte(`foo: { $ref: foo.yml }`), &examples)
				require.NoError(t, err)
				require.Equal(t, "foo.yml", examples["foo"].Ref)
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

func TestExample_Parse(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "Example is nil",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, nil
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo", openapitest.NewPath(
						openapitest.WithOperation(http.MethodGet, openapitest.NewOperation(
							openapitest.WithResponse(http.StatusOK,
								openapitest.WithContent("application/json",
									&openapi.MediaType{Examples: map[string]*openapi.ExampleRef{"foo": nil}}),
							))),
					)),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.NoError(t, err)
			},
		},
		{
			name: "Example",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, nil
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo", openapitest.NewPath(
						openapitest.WithOperation(http.MethodGet, openapitest.NewOperation(
							openapitest.WithResponse(http.StatusOK,
								openapitest.WithContent("application/json",
									&openapi.MediaType{Examples: map[string]*openapi.ExampleRef{"foo": {}}}),
							))),
					)),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.NoError(t, err)
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
								openapitest.WithContent("application/json", &openapi.MediaType{
									Examples: map[string]*openapi.ExampleRef{"foo": {Reference: ref.Reference{Ref: "foo.yml"}}},
								}),
							))),
					)),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.EqualError(t, err, "parse path '/foo' failed: parse operation 'GET' failed: parse response '200' failed: parse content 'application/json' failed: parse example 'foo' failed: resolve reference 'foo.yml' failed: TEST ERROR")
			},
		},
		{
			name: "resolve external value",
			test: func(t *testing.T) {
				calledReader := false
				reader := dynamictest.ReaderFunc(func(u *url.URL, _ any) (*dynamic.Config, error) {
					require.Equal(t, "https://foo.bar", u.String())
					cfg := &dynamic.Config{Info: dynamic.ConfigInfo{Url: u}, Data: "foobar"}
					calledReader = true
					return cfg, nil
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithPath("/foo", openapitest.NewPath(
						openapitest.WithOperation(http.MethodGet, openapitest.NewOperation(
							openapitest.WithResponse(http.StatusOK,
								openapitest.WithContent("application/json", &openapi.MediaType{
									Examples: map[string]*openapi.ExampleRef{"foo": {Value: &openapi.Example{ExternalValue: "https://foo.bar"}}},
								}),
							))),
					)),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.NoError(t, err)
				require.True(t, calledReader, "reader not called")
				content := config.Paths["/foo"].Value.Get.Responses.GetResponse(http.StatusOK).Content["application/json"]
				require.Equal(t, "foobar", content.Examples["foo"].Value.Value)
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

func TestConfig_Patch_Example(t *testing.T) {
	testcases := []struct {
		name    string
		configs []*openapi.Config
		test    func(t *testing.T, result *openapi.Config)
	}{
		{
			name: "patch Examples",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/plain", &openapi.MediaType{
								Examples: map[string]*openapi.ExampleRef{"foo": {Value: &openapi.Example{}}},
							})),
						),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/plain", &openapi.MediaType{
								Examples: map[string]*openapi.ExampleRef{"foo": {Value: &openapi.Example{
									Summary:       "foo summary",
									Description:   "foo description",
									Value:         "foo",
									ExternalValue: "https://foo.bar",
								}}},
							})),
						),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				res := result.Paths["/foo"].Value.Post.Responses.GetResponse(200)
				content := res.Content["text/plain"]
				ex := content.Examples["foo"]
				require.Equal(t, "foo summary", ex.Value.Summary)
				require.Equal(t, "foo description", ex.Value.Description)
				require.Equal(t, "foo", ex.Value.Value)
				require.Equal(t, "https://foo.bar", ex.Value.ExternalValue)
			},
		},
		{
			name: "source Examples is nil",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/plain", &openapi.MediaType{
								Examples: map[string]*openapi.ExampleRef{"foo": nil},
							})),
						),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/plain", &openapi.MediaType{
								Examples: map[string]*openapi.ExampleRef{"foo": {Value: &openapi.Example{
									Summary:       "foo summary",
									Description:   "foo description",
									Value:         "foo",
									ExternalValue: "https://foo.bar",
								}}},
							})),
						),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				res := result.Paths["/foo"].Value.Post.Responses.GetResponse(200)
				content := res.Content["text/plain"]
				ex := content.Examples["foo"]
				require.Equal(t, "foo summary", ex.Value.Summary)
			},
		},
		{
			name: "source Examples value is nil",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/plain", &openapi.MediaType{
								Examples: map[string]*openapi.ExampleRef{"foo": {}},
							})),
						),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/plain", &openapi.MediaType{
								Examples: map[string]*openapi.ExampleRef{"foo": {Value: &openapi.Example{
									Summary: "foo summary",
								}}},
							})),
						),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				res := result.Paths["/foo"].Value.Post.Responses.GetResponse(200)
				content := res.Content["text/plain"]
				ex := content.Examples["foo"]
				require.Equal(t, "foo summary", ex.Value.Summary)
			},
		},
		{
			name: "patch Examples is nil",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/plain", &openapi.MediaType{
								Examples: map[string]*openapi.ExampleRef{"foo": {Value: &openapi.Example{
									Summary: "foo summary",
								}}},
							})),
						),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/plain", &openapi.MediaType{
								Examples: map[string]*openapi.ExampleRef{"foo": nil},
							})),
						),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				res := result.Paths["/foo"].Value.Post.Responses.GetResponse(200)
				content := res.Content["text/plain"]
				ex := content.Examples["foo"]
				require.Equal(t, "foo summary", ex.Value.Summary)
			},
		},
		{
			name: "patch Examples value is nil",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/plain", &openapi.MediaType{
								Examples: map[string]*openapi.ExampleRef{"foo": {Value: &openapi.Example{
									Summary: "foo summary",
								}}},
							})),
						),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/plain", &openapi.MediaType{
								Examples: map[string]*openapi.ExampleRef{"foo": {}},
							})),
						),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				res := result.Paths["/foo"].Value.Post.Responses.GetResponse(200)
				content := res.Content["text/plain"]
				ex := content.Examples["foo"]
				require.Equal(t, "foo summary", ex.Value.Summary)
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
