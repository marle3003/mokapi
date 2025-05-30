package openapi_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	"testing"
)

func TestConfig_Patch_Server(t *testing.T) {
	testcases := []struct {
		name    string
		configs []*openapi.Config
		test    func(t *testing.T, result *openapi.Config)
	}{
		{
			name: "patch without server",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithServer("foo.bar", "description")),
				openapitest.NewConfig("1.0"),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Len(t, result.Servers, 1)
				require.Equal(t, "foo.bar", result.Servers[0].Url)
				require.Equal(t, "description", result.Servers[0].Description)
			},
		},
		{
			name: "patch server",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0"),
				openapitest.NewConfig("1.0", openapitest.WithServer("mokapi.io", "mokapi")),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Len(t, result.Servers, 1)
				require.Equal(t, "mokapi.io", result.Servers[0].Url)
				require.Equal(t, "mokapi", result.Servers[0].Description)
			},
		},
		{
			name: "patch extend servers",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithServer("foo.bar", "description")),
				openapitest.NewConfig("1.0", openapitest.WithServer("mokapi.io", "mokapi")),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Len(t, result.Servers, 2)
				require.Equal(t, "foo.bar", result.Servers[0].Url)
				require.Equal(t, "description", result.Servers[0].Description)
				require.Equal(t, "mokapi.io", result.Servers[1].Url)
				require.Equal(t, "mokapi", result.Servers[1].Description)
			},
		},
		{
			name: "patch server description",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithServer("foo.bar", "")),
				openapitest.NewConfig("1.0", openapitest.WithServer("foo.bar", "foo")),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Len(t, result.Servers, 1)
				require.Equal(t, "foo.bar", result.Servers[0].Url)
				require.Equal(t, "foo", result.Servers[0].Description)
			},
		},
		{
			name: "patch server description is overwritten",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithServer("foo.bar", "description")),
				openapitest.NewConfig("1.0", openapitest.WithServer("foo.bar", "foo")),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Len(t, result.Servers, 1)
				require.Equal(t, "foo.bar", result.Servers[0].Url)
				require.Equal(t, "foo", result.Servers[0].Description)
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

/*func TestServer_Validate(t *testing.T) {
	testdata := []struct {
		name string
		data string
		test func(t *testing.T, err error)
	}{
		{
			name: "server without url",
			data: `
openapi: 3
info:
  title: foo
servers:
 - description: foo
`,
			test: func(t *testing.T, err error) {
				require.EqualError(t, err, "semantic error at info: missing required field 'version', expected schema type=object properties=[title, description, termsOfService, contact, license, version] required=[title version] free-form=false")
			},
		},
	}

	feature.Enable([]string{"openapi-validation"})
	defer feature.Reset()

	for _, tc := range testdata {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			c := &openapi.Config{}
			err := yaml.Unmarshal([]byte(tc.data), c)
			require.NoError(t, err)
			tc.test(t, c.Validate())
		})
	}
}*/
