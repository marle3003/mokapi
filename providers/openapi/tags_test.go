package openapi_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	"testing"
)

func TestConfig_Patch_Tag(t *testing.T) {
	testcases := []struct {
		name    string
		configs []*openapi.Config
		test    func(t *testing.T, result *openapi.Config)
	}{
		{
			name: "patch summary",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithTag("foo", "", "")),
				openapitest.NewConfig("1.0", openapitest.WithTag("foo", "bar", "")),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Equal(t, "bar", result.Tags[0].Summary)
			},
		},
		{
			name: "patch summary is overwritten",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithTag("", "bar", "")),
				openapitest.NewConfig("1.0", openapitest.WithTag("", "yuh", "")),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Equal(t, "yuh", result.Tags[0].Summary)
			},
		},
		{
			name: "adds tag when name not matches",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithTag("", "bar", "foo")),
				openapitest.NewConfig("1.0", openapitest.WithTag("bar", "yuh", "")),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Equal(t, "bar", result.Tags[0].Summary)
				require.Equal(t, "yuh", result.Tags[1].Summary)
			},
		},
		{
			name: "patch description",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithTag("", "", "")),
				openapitest.NewConfig("1.0", openapitest.WithTag("", "", "foo")),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Equal(t, "foo", result.Tags[0].Description)
			},
		},
		{
			name: "patch description is overwritten",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithTag("", "", "foo")),
				openapitest.NewConfig("1.0", openapitest.WithTag("", "", "bar")),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Equal(t, "bar", result.Tags[0].Description)
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
