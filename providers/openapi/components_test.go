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
	"mokapi/providers/openapi/parameter"
	"mokapi/providers/openapi/schema"
	jsonSchema "mokapi/schema/json/schema"
	"net/url"
	"testing"
)

func TestComponents_UnmarshalJSON(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "schemas",
			test: func(t *testing.T) {
				c := openapi.Components{}
				err := json.Unmarshal([]byte(`{ "schemas": {"foo": {"type": "string"}} }`), &c)
				require.NoError(t, err)
				require.Equal(t, 1, c.Schemas.Len())
				require.Equal(t, "string", c.Schemas.Get("foo").Value.Type.String())
			},
		},
		{
			name: "responses",
			test: func(t *testing.T) {
				c := openapi.Components{}
				err := json.Unmarshal([]byte(`{ "responses": {"foo": {"description": "foo"}} }`), &c)
				require.NoError(t, err)
				require.Equal(t, 1, c.Responses.Len())
				r, _ := c.Responses.Get("foo")
				require.Equal(t, "foo", r.Value.Description)
			},
		},
		{
			name: "requestBodies",
			test: func(t *testing.T) {
				c := openapi.Components{}
				err := json.Unmarshal([]byte(`{ "requestBodies": {"foo": {"description": "foo"}} }`), &c)
				require.NoError(t, err)
				require.Len(t, c.RequestBodies, 1)
				require.Equal(t, "foo", c.RequestBodies["foo"].Value.Description)
			},
		},
		{
			name: "parameters",
			test: func(t *testing.T) {
				c := openapi.Components{}
				err := json.Unmarshal([]byte(`{ "parameters": {"foo": {"description": "foo"}} }`), &c)
				require.NoError(t, err)
				require.Len(t, c.Parameters, 1)
				require.Equal(t, "foo", c.Parameters["foo"].Value.Description)
			},
		},
		{
			name: "examples",
			test: func(t *testing.T) {
				c := openapi.Components{}
				err := json.Unmarshal([]byte(`{ "examples": {"foo": {"description": "foo"}} }`), &c)
				require.NoError(t, err)
				require.Len(t, c.Examples, 1)
				require.Equal(t, "foo", c.Examples["foo"].Value.Description)
			},
		},
		{
			name: "examples",
			test: func(t *testing.T) {
				c := openapi.Components{}
				err := json.Unmarshal([]byte(`{ "examples": {"foo": {"description": "foo"}} }`), &c)
				require.NoError(t, err)
				require.Len(t, c.Examples, 1)
				require.Equal(t, "foo", c.Examples["foo"].Value.Description)
			},
		},
		{
			name: "headers",
			test: func(t *testing.T) {
				c := openapi.Components{}
				err := json.Unmarshal([]byte(`{ "headers": {"foo": {"description": "foo"}} }`), &c)
				require.NoError(t, err)
				require.Len(t, c.Headers, 1)
				require.Equal(t, "foo", c.Headers["foo"].Value.Description)
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

func TestComponents_UnmarshalYAML(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "schemas",
			test: func(t *testing.T) {
				c := openapi.Components{}
				err := yaml.Unmarshal([]byte(`schemas: {foo: {type: string}}`), &c)
				require.NoError(t, err)
				require.Equal(t, 1, c.Schemas.Len())
				require.Equal(t, "string", c.Schemas.Get("foo").Value.Type.String())
			},
		},
		{
			name: "responses",
			test: func(t *testing.T) {
				c := openapi.Components{}
				err := yaml.Unmarshal([]byte(`responses: {foo: {description: foo}}`), &c)
				require.NoError(t, err)
				require.Equal(t, 1, c.Responses.Len())
				r, _ := c.Responses.Get("foo")
				require.Equal(t, "foo", r.Value.Description)
			},
		},
		{
			name: "requestBodies",
			test: func(t *testing.T) {
				c := openapi.Components{}
				err := yaml.Unmarshal([]byte(`requestBodies: {foo: {description: foo}}`), &c)
				require.NoError(t, err)
				require.Len(t, c.RequestBodies, 1)
				require.Equal(t, "foo", c.RequestBodies["foo"].Value.Description)
			},
		},
		{
			name: "parameters",
			test: func(t *testing.T) {
				c := openapi.Components{}
				err := yaml.Unmarshal([]byte(`parameters: {foo: {description: foo}}`), &c)
				require.NoError(t, err)
				require.Len(t, c.Parameters, 1)
				require.Equal(t, "foo", c.Parameters["foo"].Value.Description)
			},
		},
		{
			name: "examples",
			test: func(t *testing.T) {
				c := openapi.Components{}
				err := yaml.Unmarshal([]byte(`examples: {foo: {description: foo}}`), &c)
				require.NoError(t, err)
				require.Len(t, c.Examples, 1)
				require.Equal(t, "foo", c.Examples["foo"].Value.Description)
			},
		},
		{
			name: "examples",
			test: func(t *testing.T) {
				c := openapi.Components{}
				err := yaml.Unmarshal([]byte(`examples: {foo: {description: foo}}`), &c)
				require.NoError(t, err)
				require.Len(t, c.Examples, 1)
				require.Equal(t, "foo", c.Examples["foo"].Value.Description)
			},
		},
		{
			name: "headers",
			test: func(t *testing.T) {
				c := openapi.Components{}
				err := yaml.Unmarshal([]byte(`headers: {foo: {description: foo}}`), &c)
				require.NoError(t, err)
				require.Len(t, c.Headers, 1)
				require.Equal(t, "foo", c.Headers["foo"].Value.Description)
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

func TestComponents_Parse(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "schema ref",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(u *url.URL, _ any) (*dynamic.Config, error) {
					cfg := &dynamic.Config{
						Info: dynamic.ConfigInfo{Url: u},
						Data: openapitest.NewConfig("3.0",
							openapitest.WithComponentSchema("foo", &schema.Schema{Type: jsonSchema.Types{"string"}}),
						),
					}
					return cfg, nil
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithComponentSchemaRef("foo", &schema.Ref{Reference: dynamic.Reference{Ref: "foo.yml#/components/schemas/foo"}}),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.NoError(t, err)
				require.Equal(t, "string", config.Components.Schemas.Get("foo").Value.Type.String())
			},
		},
		{
			name: "schema ref error",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, fmt.Errorf("TESTING ERROR")
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithComponentSchemaRef("foo", &schema.Ref{Reference: dynamic.Reference{Ref: "foo.yml#/components/schemas/foo"}}),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.EqualError(t, err, "parse components failed: parse schema 'foo' failed: resolve reference 'foo.yml#/components/schemas/foo' failed: TESTING ERROR")
			},
		},
		{
			name: "response ref",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(u *url.URL, _ any) (*dynamic.Config, error) {
					cfg := &dynamic.Config{
						Info: dynamic.ConfigInfo{Url: u},
						Data: openapitest.NewConfig("3.0",
							openapitest.WithComponentResponse("foo", &openapi.Response{Description: "foo"}),
						),
					}
					return cfg, nil
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithComponentResponseRef("foo", &openapi.ResponseRef{Reference: dynamic.Reference{Ref: "foo.yml#/components/responses/foo"}}),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.NoError(t, err)
				r, _ := config.Components.Responses.Get("foo")
				require.Equal(t, "foo", r.Value.Description)
			},
		},
		{
			name: "response ref error",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, fmt.Errorf("TESTING ERROR")
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithComponentResponseRef("foo", &openapi.ResponseRef{Reference: dynamic.Reference{Ref: "foo.yml#/components/schemas/foo"}}),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.EqualError(t, err, "parse components failed: parse response 'foo' failed: resolve reference 'foo.yml#/components/schemas/foo' failed: TESTING ERROR")
			},
		},
		{
			name: "requestBody ref",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(u *url.URL, _ any) (*dynamic.Config, error) {
					cfg := &dynamic.Config{
						Info: dynamic.ConfigInfo{Url: u},
						Data: openapitest.NewConfig("3.0",
							openapitest.WithComponentRequestBody("foo", &openapi.RequestBody{Description: "foo"}),
						),
					}
					return cfg, nil
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithComponentRequestBodyRef("foo", &openapi.RequestBodyRef{Reference: dynamic.Reference{Ref: "foo.yml#/components/requestBodies/foo"}}),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.NoError(t, err)
				require.Equal(t, "foo", config.Components.RequestBodies["foo"].Value.Description)
			},
		},
		{
			name: "requestBody ref error",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, fmt.Errorf("TESTING ERROR")
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithComponentRequestBodyRef("foo", &openapi.RequestBodyRef{Reference: dynamic.Reference{Ref: "foo.yml#/components/requestBodies/foo"}}),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.EqualError(t, err, "parse components failed: parse request body 'foo' failed: resolve reference 'foo.yml#/components/requestBodies/foo' failed: TESTING ERROR")
			},
		},
		{
			name: "parameter ref",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(u *url.URL, _ any) (*dynamic.Config, error) {
					cfg := &dynamic.Config{
						Info: dynamic.ConfigInfo{Url: u},
						Data: openapitest.NewConfig("3.0",
							openapitest.WithComponentParameter("foo", &parameter.Parameter{Description: "foo"}),
						),
					}
					return cfg, nil
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithComponentParameterRef("foo", &parameter.Ref{Reference: dynamic.Reference{Ref: "foo.yml#/components/parameters/foo"}}),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.NoError(t, err)
				require.Equal(t, "foo", config.Components.Parameters["foo"].Value.Description)
			},
		},
		{
			name: "parameter ref error",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, fmt.Errorf("TESTING ERROR")
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithComponentParameterRef("foo", &parameter.Ref{Reference: dynamic.Reference{Ref: "foo.yml#/components/parameters/foo"}}),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.EqualError(t, err, "parse components failed: parse parameter 'foo' failed: resolve reference 'foo.yml#/components/parameters/foo' failed: TESTING ERROR")
			},
		},
		{
			name: "example ref",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(u *url.URL, _ any) (*dynamic.Config, error) {
					cfg := &dynamic.Config{
						Info: dynamic.ConfigInfo{Url: u},
						Data: openapitest.NewConfig("3.0",
							openapitest.WithComponentExample("foo", &openapi.Example{Description: "foo"}),
						),
					}
					return cfg, nil
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithComponentExampleRef("foo", &openapi.ExampleRef{Reference: dynamic.Reference{Ref: "foo.yml#/components/examples/foo"}}),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.NoError(t, err)
				require.Equal(t, "foo", config.Components.Examples["foo"].Value.Description)
			},
		},
		{
			name: "example ref error",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, fmt.Errorf("TESTING ERROR")
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithComponentExampleRef("foo", &openapi.ExampleRef{Reference: dynamic.Reference{Ref: "foo.yml#/components/parameters/foo"}}),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.EqualError(t, err, "parse components failed: parse example 'foo' failed: resolve reference 'foo.yml#/components/parameters/foo' failed: TESTING ERROR")
			},
		},
		{
			name: "header ref",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(u *url.URL, _ any) (*dynamic.Config, error) {
					cfg := &dynamic.Config{
						Info: dynamic.ConfigInfo{Url: u},
						Data: openapitest.NewConfig("3.0",
							openapitest.WithComponentHeader("foo", &openapi.Header{Parameter: parameter.Parameter{Description: "foo"}}),
						),
					}
					return cfg, nil
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithComponentHeaderRef("foo", &openapi.HeaderRef{Reference: dynamic.Reference{Ref: "foo.yml#/components/headers/foo"}}),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.NoError(t, err)
				require.Equal(t, "foo", config.Components.Headers["foo"].Value.Description)
			},
		},
		{
			name: "header ref error",
			test: func(t *testing.T) {
				reader := dynamictest.ReaderFunc(func(_ *url.URL, _ any) (*dynamic.Config, error) {
					return nil, fmt.Errorf("TESTING ERROR")
				})
				config := openapitest.NewConfig("3.0",
					openapitest.WithComponentHeaderRef("foo", &openapi.HeaderRef{Reference: dynamic.Reference{Ref: "foo.yml#/components/headers/foo"}}),
				)
				err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
				require.EqualError(t, err, "parse components failed: parse header 'foo' failed: resolve reference 'foo.yml#/components/headers/foo' failed: TESTING ERROR")
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

func TestConfig_Patch_Components(t *testing.T) {
	testcases := []struct {
		name    string
		configs []*openapi.Config
		test    func(t *testing.T, result *openapi.Config)
	}{
		{
			name: "add schema",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0"),
				openapitest.NewConfig("1.0", openapitest.WithComponentSchema("foo", &schema.Schema{Type: jsonSchema.Types{"string"}})),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Equal(t, "string", result.Components.Schemas.Get("foo").Value.Type.String())
			},
		},
		{
			name: "patch schema",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithComponentSchema("foo", &schema.Schema{Type: jsonSchema.Types{"string"}})),
				openapitest.NewConfig("1.0", openapitest.WithComponentSchema("foo", &schema.Schema{Format: "bar"})),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Equal(t, "string", result.Components.Schemas.Get("foo").Value.Type.String())
				require.Equal(t, "bar", result.Components.Schemas.Get("foo").Value.Format)
			},
		},
		{
			name: "add response",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0"),
				openapitest.NewConfig("1.0", openapitest.WithComponentResponse("foo", &openapi.Response{Description: "foo"})),
			},
			test: func(t *testing.T, result *openapi.Config) {
				r, _ := result.Components.Responses.Get("foo")
				require.Equal(t, "foo", r.Value.Description)
			},
		},
		{
			name: "patch response",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithComponentResponse("foo", &openapi.Response{Description: "foo"})),
				openapitest.NewConfig("1.0", openapitest.WithComponentResponse("foo", &openapi.Response{Description: "bar"})),
			},
			test: func(t *testing.T, result *openapi.Config) {
				r, _ := result.Components.Responses.Get("foo")
				require.Equal(t, "bar", r.Value.Description)
			},
		},
		{
			name: "add request body",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0"),
				openapitest.NewConfig("1.0", openapitest.WithComponentRequestBody("foo", &openapi.RequestBody{Description: "foo"})),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Equal(t, "foo", result.Components.RequestBodies["foo"].Value.Description)
			},
		},
		{
			name: "patch request body",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithComponentRequestBody("foo", &openapi.RequestBody{Description: "foo"})),
				openapitest.NewConfig("1.0", openapitest.WithComponentRequestBody("foo", &openapi.RequestBody{Description: "bar"})),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Equal(t, "bar", result.Components.RequestBodies["foo"].Value.Description)
			},
		},
		{
			name: "add parameter",
			configs: []*openapi.Config{
				{Components: openapi.Components{Parameters: map[string]*parameter.Ref{}}},
				openapitest.NewConfig("1.0", openapitest.WithComponentParameter("foo", &parameter.Parameter{Description: "foo"})),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Equal(t, "foo", result.Components.Parameters["foo"].Value.Description)
			},
		},
		{
			name: "patch parameter",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithComponentParameter("foo", &parameter.Parameter{Description: "foo"})),
				openapitest.NewConfig("1.0", openapitest.WithComponentParameter("foo", &parameter.Parameter{Description: "bar"})),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Equal(t, "bar", result.Components.Parameters["foo"].Value.Description)
			},
		},
		{
			name: "add example",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0"),
				openapitest.NewConfig("1.0", openapitest.WithComponentExample("foo", &openapi.Example{Description: "foo"})),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Equal(t, "foo", result.Components.Examples["foo"].Value.Description)
			},
		},
		{
			name: "patch example",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithComponentExample("foo", &openapi.Example{Description: "foo"})),
				openapitest.NewConfig("1.0", openapitest.WithComponentExample("foo", &openapi.Example{Description: "bar"})),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Equal(t, "bar", result.Components.Examples["foo"].Value.Description)
			},
		},
		{
			name: "add header",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0"),
				openapitest.NewConfig("1.0", openapitest.WithComponentHeader("foo", &openapi.Header{Parameter: parameter.Parameter{Description: "foo"}})),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Equal(t, "foo", result.Components.Headers["foo"].Value.Description)
			},
		},
		{
			name: "patch header",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithComponentHeader("foo", &openapi.Header{Parameter: parameter.Parameter{Description: "foo"}})),
				openapitest.NewConfig("1.0", openapitest.WithComponentHeader("foo", &openapi.Header{Parameter: parameter.Parameter{Description: "bar"}})),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Equal(t, "bar", result.Components.Headers["foo"].Value.Description)
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
