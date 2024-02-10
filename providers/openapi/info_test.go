package openapi_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	"testing"
)

func TestConfig_Patch_Info(t *testing.T) {
	testcases := []struct {
		name    string
		configs []*openapi.Config
		test    func(t *testing.T, result *openapi.Config)
	}{
		{
			name: "patch description",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithInfo("", "1.0", "")),
				openapitest.NewConfig("1.0", openapitest.WithInfo("", "1.0", "foo")),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Equal(t, "foo", result.Info.Description)
			},
		},
		{
			name: "patch description is overwritten",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithInfo("", "1.0", "foo")),
				openapitest.NewConfig("1.0", openapitest.WithInfo("", "1.0", "bar")),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Equal(t, "bar", result.Info.Description)
			},
		},
		{
			name: "patch version",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithInfo("", "", "")),
				openapitest.NewConfig("1.0", openapitest.WithInfo("", "1.0", "")),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Equal(t, "1.0", result.Info.Version)
			},
		},
		{
			name: "patch version is overwritten",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithInfo("", "1.0", "")),
				openapitest.NewConfig("1.0", openapitest.WithInfo("", "3.0", "")),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Equal(t, "3.0", result.Info.Version)
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
