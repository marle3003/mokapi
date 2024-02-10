package openapi_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	"testing"
)

func TestConfig_Patch_Contact(t *testing.T) {
	testcases := []struct {
		name    string
		configs []*openapi.Config
		test    func(t *testing.T, result *openapi.Config)
	}{
		{
			name: "patch without contact",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithContact("foo", "foo.bar", "info@foo.bar")),
				openapitest.NewConfig("1.0"),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.NotNil(t, result.Info.Contact)
				require.Equal(t, "foo", result.Info.Contact.Name)
				require.Equal(t, "foo.bar", result.Info.Contact.Url)
				require.Equal(t, "info@foo.bar", result.Info.Contact.Email)
			},
		},
		{
			name: "patch with contact",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0"),
				openapitest.NewConfig("1.0", openapitest.WithContact("foo", "foo.bar", "info@foo.bar")),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.NotNil(t, result.Info.Contact)
				require.Equal(t, "foo", result.Info.Contact.Name)
				require.Equal(t, "foo.bar", result.Info.Contact.Url)
				require.Equal(t, "info@foo.bar", result.Info.Contact.Email)
			},
		},
		{
			name: "patch name",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithContact("", "foo.bar", "info@foo.bar")),
				openapitest.NewConfig("1.0", openapitest.WithContact("foo", "", "")),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.NotNil(t, result.Info.Contact)
				require.Equal(t, "foo", result.Info.Contact.Name)
				require.Equal(t, "foo.bar", result.Info.Contact.Url)
				require.Equal(t, "info@foo.bar", result.Info.Contact.Email)
			},
		},
		{
			name: "patch name is overwritten",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithContact("foo", "foo.bar", "info@foo.bar")),
				openapitest.NewConfig("1.0", openapitest.WithContact("bar", "", "")),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.NotNil(t, result.Info.Contact)
				require.Equal(t, "bar", result.Info.Contact.Name)
				require.Equal(t, "foo.bar", result.Info.Contact.Url)
				require.Equal(t, "info@foo.bar", result.Info.Contact.Email)
			},
		},
		{
			name: "patch url",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithContact("foo", "", "info@foo.bar")),
				openapitest.NewConfig("1.0", openapitest.WithContact("", "foo.bar", "")),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.NotNil(t, result.Info.Contact)
				require.Equal(t, "foo", result.Info.Contact.Name)
				require.Equal(t, "foo.bar", result.Info.Contact.Url)
				require.Equal(t, "info@foo.bar", result.Info.Contact.Email)
			},
		},
		{
			name: "patch url is overwritten",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithContact("foo", "foo.bar", "info@foo.bar")),
				openapitest.NewConfig("1.0", openapitest.WithContact("", "mokapi.io", "")),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.NotNil(t, result.Info.Contact)
				require.Equal(t, "foo", result.Info.Contact.Name)
				require.Equal(t, "mokapi.io", result.Info.Contact.Url)
				require.Equal(t, "info@foo.bar", result.Info.Contact.Email)
			},
		},
		{
			name: "patch email",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithContact("foo", "foo.bar", "")),
				openapitest.NewConfig("1.0", openapitest.WithContact("", "", "info@foo.bar")),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.NotNil(t, result.Info.Contact)
				require.Equal(t, "foo", result.Info.Contact.Name)
				require.Equal(t, "foo.bar", result.Info.Contact.Url)
				require.Equal(t, "info@foo.bar", result.Info.Contact.Email)
			},
		},
		{
			name: "patch email is overwritten",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithContact("foo", "foo.bar", "info@foo.bar")),
				openapitest.NewConfig("1.0", openapitest.WithContact("", "", "info@mokapi.io")),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.NotNil(t, result.Info.Contact)
				require.Equal(t, "foo", result.Info.Contact.Name)
				require.Equal(t, "foo.bar", result.Info.Contact.Url)
				require.Equal(t, "info@mokapi.io", result.Info.Contact.Email)
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
