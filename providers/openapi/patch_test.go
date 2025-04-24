package openapi_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	"mokapi/providers/openapi/schema/schematest"
	"testing"
)

func TestConfig_Patch_Methods_RequestBody(t *testing.T) {
	testcases := []struct {
		name    string
		configs []*openapi.Config
		test    func(t *testing.T, result *openapi.Config)
	}{
		{
			name: "description and required",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(openapitest.WithRequestBody("foo", true)),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Equal(t, "foo", result.Paths["/foo"].Value.Post.RequestBody.Value.Description)
				require.True(t, result.Paths["/foo"].Value.Post.RequestBody.Value.Required)
			},
		},
		{
			name: "patch description and required",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(openapitest.WithRequestBody("", false)),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(openapitest.WithRequestBody("foo", true)),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Equal(t, "foo", result.Paths["/foo"].Value.Post.RequestBody.Value.Description)
				require.True(t, result.Paths["/foo"].Value.Post.RequestBody.Value.Required)
			},
		},
		{
			name: "patch add content type",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithRequestBody("foo", true,
								openapitest.WithRequestContent("text/plain", &openapi.MediaType{}))),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithRequestBody("foo", true,
								openapitest.WithRequestContent("application/json", &openapi.MediaType{}))),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				body := result.Paths["/foo"].Value.Post.RequestBody.Value
				require.Contains(t, body.Content, "text/plain")
				require.Contains(t, body.Content, "application/json")
			},
		},
		{
			name: "add content type schema",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithRequestBody("foo", true,
								openapitest.WithRequestContent("text/plain", &openapi.MediaType{}))),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithRequestBody("foo", true,
								openapitest.WithRequestContent("text/plain",
									openapitest.NewContent(openapitest.WithSchema(schematest.New("number")))))),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				body := result.Paths["/foo"].Value.Post.RequestBody.Value
				require.Len(t, body.Content, 1)
				require.Equal(t, "number", body.Content["text/plain"].Schema.Type.String())
			},
		},
		{
			name: "patch content type schema",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithRequestBody("foo", true,
								openapitest.WithRequestContent("text/plain",
									openapitest.NewContent(openapitest.WithSchema(schematest.New("number")))))),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithRequestBody("foo", true,
								openapitest.WithRequestContent("text/plain",
									openapitest.NewContent(openapitest.WithSchema(schematest.New("number", schematest.WithFormat("double"))))))),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				body := result.Paths["/foo"].Value.Post.RequestBody.Value
				require.Len(t, body.Content, 1)
				require.Equal(t, "number", body.Content["text/plain"].Schema.Type.String())
				require.Equal(t, "double", body.Content["text/plain"].Schema.Format)
			},
		},
		{
			name: "patch content type example",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithRequestBody("foo", true,
								openapitest.WithRequestContent("text/plain", &openapi.MediaType{}))),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithRequestBody("foo", true,
								openapitest.WithRequestContent("text/plain",
									openapitest.NewContent(openapitest.WithExample(12))))),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				body := result.Paths["/foo"].Value.Post.RequestBody.Value
				require.Len(t, body.Content, 1)
				require.Equal(t, 12, body.Content["text/plain"].Example.Value)
			},
		},
		{
			name: "patch config security",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0"),
				openapitest.NewConfig("1.0",
					openapitest.WithGlobalSecurity(map[string][]string{"foo": {}}),
				),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Len(t, result.Security, 1)
				require.Contains(t, result.Security[0], "foo")
			},
		},
		{
			name: "patch config security add scope",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0",
					openapitest.WithGlobalSecurity(map[string][]string{"foo": {"foo"}}),
				),
				openapitest.NewConfig("1.0",
					openapitest.WithGlobalSecurity(map[string][]string{"foo": {"foo", "bar"}}),
				),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Len(t, result.Security, 2)
				require.Equal(t, []string{"foo"}, result.Security[0]["foo"])
				require.Equal(t, []string{"foo", "bar"}, result.Security[1]["foo"])
			},
		},
		{
			name: "patch operation security",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithSecurity(map[string][]string{"foo": {}})),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Len(t, result.Paths["/foo"].Value.Post.Security, 1)
				require.Contains(t, result.Paths["/foo"].Value.Post.Security[0], "foo")
			},
		},
		{
			name: "patch operation security add scope",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithSecurity(map[string][]string{"foo": {"foo"}}),
						),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithPath(
					"/foo", openapitest.NewPath(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithSecurity(map[string][]string{"foo": {"foo", "bar"}})),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Len(t, result.Paths["/foo"].Value.Post.Security, 2)
				require.Equal(t, []string{"foo"}, result.Paths["/foo"].Value.Post.Security[0]["foo"])
				require.Equal(t, []string{"foo", "bar"}, result.Paths["/foo"].Value.Post.Security[1]["foo"])
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
