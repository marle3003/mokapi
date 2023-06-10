package openapi_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/openapi"
	"mokapi/config/dynamic/openapi/openapitest"
	"mokapi/config/dynamic/openapi/schema/schematest"
	"strings"
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
			name: "patch server description is not overwritten",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithServer("foo.bar", "description")),
				openapitest.NewConfig("1.0", openapitest.WithServer("foo.bar", "foo")),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Len(t, result.Servers, 1)
				require.Equal(t, "foo.bar", result.Servers[0].Url)
				require.Equal(t, "description", result.Servers[0].Description)
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

func TestConfig_Patch_Info(t *testing.T) {
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
			name: "patch contact name",
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
			name: "patch contact name but not overwritten",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithContact("foo", "foo.bar", "info@foo.bar")),
				openapitest.NewConfig("1.0", openapitest.WithContact("bar", "", "")),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.NotNil(t, result.Info.Contact)
				require.Equal(t, "foo", result.Info.Contact.Name)
				require.Equal(t, "foo.bar", result.Info.Contact.Url)
				require.Equal(t, "info@foo.bar", result.Info.Contact.Email)
			},
		},
		{
			name: "patch contact url",
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
			name: "patch contact url but not overwritten",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithContact("foo", "foo.bar", "info@foo.bar")),
				openapitest.NewConfig("1.0", openapitest.WithContact("", "mokapi.io", "")),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.NotNil(t, result.Info.Contact)
				require.Equal(t, "foo", result.Info.Contact.Name)
				require.Equal(t, "foo.bar", result.Info.Contact.Url)
				require.Equal(t, "info@foo.bar", result.Info.Contact.Email)
			},
		},
		{
			name: "patch contact email",
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
			name: "patch contact email but not overwritten",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithContact("foo", "foo.bar", "info@foo.bar")),
				openapitest.NewConfig("1.0", openapitest.WithContact("", "", "info@mokapi.io")),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.NotNil(t, result.Info.Contact)
				require.Equal(t, "foo", result.Info.Contact.Name)
				require.Equal(t, "foo.bar", result.Info.Contact.Url)
				require.Equal(t, "info@foo.bar", result.Info.Contact.Email)
			},
		},
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
			name: "patch description is not overwritten",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithInfo("", "1.0", "foo")),
				openapitest.NewConfig("1.0", openapitest.WithInfo("", "1.0", "bar")),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Equal(t, "foo", result.Info.Description)
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
			name: "patch version is not overwritten",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithInfo("", "1.0", "")),
				openapitest.NewConfig("1.0", openapitest.WithInfo("", "3.0", "")),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Equal(t, "1.0", result.Info.Version)
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

func TestConfig_Patch_Path(t *testing.T) {
	testcases := []struct {
		name    string
		configs []*openapi.Config
		test    func(t *testing.T, result *openapi.Config)
	}{
		{
			name: "patch without path",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithEndpoint(
					"/foo", openapitest.NewEndpoint(openapitest.WithOperation(
						"post", openapitest.NewOperation(),
					),
					))),
				openapitest.NewConfig("1.0"),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Len(t, result.Paths.Value, 1)
				require.Contains(t, result.Paths.Value, "/foo")
				require.NotNil(t, result.Paths.Value["/foo"].Value.Post)
			},
		},
		{
			name: "patch path",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0"),
				openapitest.NewConfig("1.0", openapitest.WithEndpoint(
					"/foo", openapitest.NewEndpoint(openapitest.WithOperation(
						"post", openapitest.NewOperation(),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Len(t, result.Paths.Value, 1)
				require.Contains(t, result.Paths.Value, "/foo")
				require.NotNil(t, result.Paths.Value["/foo"].Value.Post)
			},
		},
		{
			name: "patch summary and description",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithEndpoint(
					"/foo", openapitest.NewEndpoint())),
				openapitest.NewConfig("1.0", openapitest.WithEndpoint(
					"/foo", openapitest.NewEndpoint(openapitest.WithEndpointInfo("foo", "bar")))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Len(t, result.Paths.Value, 1)
				require.Contains(t, result.Paths.Value, "/foo")
				require.Equal(t, "foo", result.Paths.Value["/foo"].Value.Summary)
				require.Equal(t, "bar", result.Paths.Value["/foo"].Value.Description)
			},
		},
		{
			name: "add parameters",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithEndpoint(
					"/foo", openapitest.NewEndpoint())),
				openapitest.NewConfig("1.0", openapitest.WithEndpoint(
					"/foo", openapitest.NewEndpoint(openapitest.WithEndpointParam("foo", "path", true)))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				e := result.Paths.Value["/foo"].Value
				require.Len(t, e.Parameters, 1)
			},
		},
		{
			name: "patch parameters",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithEndpoint(
					"/foo", openapitest.NewEndpoint(openapitest.WithEndpointParam("foo", "path", true)))),
				openapitest.NewConfig("1.0", openapitest.WithEndpoint(
					"/foo", openapitest.NewEndpoint(openapitest.WithEndpointParam("foo", "path", true, openapitest.WithParamSchema(schematest.New("number")))))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				e := result.Paths.Value["/foo"].Value
				require.Len(t, e.Parameters, 1)
				require.Equal(t, "number", e.Parameters[0].Value.Schema.Value.Type)
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

func TestConfig_Patch_Methods(t *testing.T) {
	methods := []string{"get", "post", "put", "patch", "delete", "head", "options", "trace"}
	type testType struct {
		name    string
		configs []*openapi.Config
		test    func(t *testing.T, result *openapi.Config)
	}
	var testcases []testType
	for _, m := range methods {
		m := m
		testcases = append(testcases, testType{
			name: "patch path with " + m,
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithEndpoint(
					"/foo", openapitest.NewEndpoint())),
				openapitest.NewConfig("1.0", openapitest.WithEndpoint(
					"/foo", openapitest.NewEndpoint(openapitest.WithOperation(m, openapitest.NewOperation())))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Len(t, result.Paths.Value, 1)
				require.Contains(t, result.Paths.Value, "/foo")
				require.Contains(t, result.Paths.Value["/foo"].Value.Operations(), strings.ToUpper(m))
			},
		})
		testcases = append(testcases, testType{
			name: fmt.Sprintf("patch path %v info", m),
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithEndpoint(
					"/foo", openapitest.NewEndpoint(openapitest.WithOperation(m, openapitest.NewOperation())))),
				openapitest.NewConfig("1.0", openapitest.WithEndpoint(
					"/foo", openapitest.NewEndpoint(openapitest.WithOperation(m, openapitest.NewOperation(
						openapitest.WithOperationInfo("foo", "bar", "id", true),
					))))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Len(t, result.Paths.Value, 1)
				require.Contains(t, result.Paths.Value, "/foo")
				o := result.Paths.Value["/foo"].Value.Operations()[strings.ToUpper(m)]
				require.Equal(t, "foo", o.Summary)
				require.Equal(t, "bar", o.Description)
				require.Equal(t, "id", o.OperationId)
				require.True(t, o.Deprecated)
			},
		})
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

func TestConfig_Patch_Methods_RequestBody(t *testing.T) {
	testcases := []struct {
		name    string
		configs []*openapi.Config
		test    func(t *testing.T, result *openapi.Config)
	}{
		{
			name: "description and required",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithEndpoint(
					"/foo", openapitest.NewEndpoint(openapitest.WithOperation(
						"post", openapitest.NewOperation(),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithEndpoint(
					"/foo", openapitest.NewEndpoint(openapitest.WithOperation(
						"post", openapitest.NewOperation(openapitest.WithRequestBody("foo", true)),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Equal(t, "foo", result.Paths.Value["/foo"].Value.Post.RequestBody.Value.Description)
				require.True(t, result.Paths.Value["/foo"].Value.Post.RequestBody.Value.Required)
			},
		},
		{
			name: "patch description and required",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithEndpoint(
					"/foo", openapitest.NewEndpoint(openapitest.WithOperation(
						"post", openapitest.NewOperation(openapitest.WithRequestBody("", false)),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithEndpoint(
					"/foo", openapitest.NewEndpoint(openapitest.WithOperation(
						"post", openapitest.NewOperation(openapitest.WithRequestBody("foo", true)),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Equal(t, "foo", result.Paths.Value["/foo"].Value.Post.RequestBody.Value.Description)
				require.True(t, result.Paths.Value["/foo"].Value.Post.RequestBody.Value.Required)
			},
		},
		{
			name: "patch add content type",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithEndpoint(
					"/foo", openapitest.NewEndpoint(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithRequestBody("foo", true,
								openapitest.WithRequestContent("text/plain"))),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithEndpoint(
					"/foo", openapitest.NewEndpoint(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithRequestBody("foo", true,
								openapitest.WithRequestContent("application/json"))),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				body := result.Paths.Value["/foo"].Value.Post.RequestBody.Value
				require.Contains(t, body.Content, "text/plain")
				require.Contains(t, body.Content, "application/json")
			},
		},
		{
			name: "add content type schema",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithEndpoint(
					"/foo", openapitest.NewEndpoint(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithRequestBody("foo", true,
								openapitest.WithRequestContent("text/plain"))),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithEndpoint(
					"/foo", openapitest.NewEndpoint(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithRequestBody("foo", true,
								openapitest.WithRequestContent("text/plain",
									openapitest.WithSchema(schematest.New("number"))))),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				body := result.Paths.Value["/foo"].Value.Post.RequestBody.Value
				require.Len(t, body.Content, 1)
				require.Equal(t, "number", body.Content["text/plain"].Schema.Value.Type)
			},
		},
		{
			name: "patch content type schema",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithEndpoint(
					"/foo", openapitest.NewEndpoint(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithRequestBody("foo", true,
								openapitest.WithRequestContent("text/plain",
									openapitest.WithSchema(schematest.New("number"))))),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithEndpoint(
					"/foo", openapitest.NewEndpoint(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithRequestBody("foo", true,
								openapitest.WithRequestContent("text/plain",
									openapitest.WithSchema(schematest.New("number", schematest.WithFormat("double")))))),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				body := result.Paths.Value["/foo"].Value.Post.RequestBody.Value
				require.Len(t, body.Content, 1)
				require.Equal(t, "number", body.Content["text/plain"].Schema.Value.Type)
				require.Equal(t, "double", body.Content["text/plain"].Schema.Value.Format)
			},
		},
		{
			name: "patch content type example",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithEndpoint(
					"/foo", openapitest.NewEndpoint(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithRequestBody("foo", true,
								openapitest.WithRequestContent("text/plain"))),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithEndpoint(
					"/foo", openapitest.NewEndpoint(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithRequestBody("foo", true,
								openapitest.WithRequestContent("text/plain",
									openapitest.WithExample(12)))),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				body := result.Paths.Value["/foo"].Value.Post.RequestBody.Value
				require.Len(t, body.Content, 1)
				require.Equal(t, 12, body.Content["text/plain"].Example)
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

func TestConfig_Patch_Methods_Response(t *testing.T) {
	testcases := []struct {
		name    string
		configs []*openapi.Config
		test    func(t *testing.T, result *openapi.Config)
	}{
		{
			name: "add response",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithEndpoint(
					"/foo", openapitest.NewEndpoint(openapitest.WithOperation(
						"post", openapitest.NewOperation(),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithEndpoint(
					"/foo", openapitest.NewEndpoint(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithResponseDescription("foo"))),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				res := result.Paths.Value["/foo"].Value.Post.Responses.GetResponse(200)
				require.Equal(t, "foo", res.Description)
			},
		},
		{
			name: "patch response",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithEndpoint(
					"/foo", openapitest.NewEndpoint(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(204, openapitest.WithResponseDescription("bar"))),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithEndpoint(
					"/foo", openapitest.NewEndpoint(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithResponseDescription("foo"))),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				res := result.Paths.Value["/foo"].Value.Post.Responses.GetResponse(204)
				require.Equal(t, "bar", res.Description)
				res = result.Paths.Value["/foo"].Value.Post.Responses.GetResponse(200)
				require.Equal(t, "foo", res.Description)
			},
		},
		{
			name: "patch description",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithEndpoint(
					"/foo", openapitest.NewEndpoint(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200)),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithEndpoint(
					"/foo", openapitest.NewEndpoint(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithResponseDescription("foo"))),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				res := result.Paths.Value["/foo"].Value.Post.Responses.GetResponse(200)
				require.Equal(t, "foo", res.Description)
			},
		},
		{
			name: "patch add content type",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithEndpoint(
					"/foo", openapitest.NewEndpoint(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/plain"))),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithEndpoint(
					"/foo", openapitest.NewEndpoint(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("application/json"))),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				res := result.Paths.Value["/foo"].Value.Post.Responses.GetResponse(200)
				require.Contains(t, res.Content, "text/plain")
				require.Contains(t, res.Content, "application/json")
			},
		},
		{
			name: "add content type schema",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithEndpoint(
					"/foo", openapitest.NewEndpoint(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/plain"))),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithEndpoint(
					"/foo", openapitest.NewEndpoint(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/plain",
								openapitest.WithSchema(schematest.New("number"))))),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				res := result.Paths.Value["/foo"].Value.Post.Responses.GetResponse(200)
				require.Len(t, res.Content, 1)
				require.Equal(t, "number", res.Content["text/plain"].Schema.Value.Type)
			},
		},
		{
			name: "patch content type schema",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithEndpoint(
					"/foo", openapitest.NewEndpoint(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/plain",
								openapitest.WithSchema(schematest.New("number"))))),
					),
					))),
				openapitest.NewConfig("1.0", openapitest.WithEndpoint(
					"/foo", openapitest.NewEndpoint(openapitest.WithOperation(
						"post", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("text/plain",
								openapitest.WithSchema(schematest.New("number", schematest.WithFormat("double")))))),
					),
					))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				res := result.Paths.Value["/foo"].Value.Post.Responses.GetResponse(200)
				require.Len(t, res.Content, 1)
				require.Equal(t, "number", res.Content["text/plain"].Schema.Value.Type)
				require.Equal(t, "double", res.Content["text/plain"].Schema.Value.Format)
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

func TestConfig_Patch_Components(t *testing.T) {
	testcases := []struct {
		name    string
		configs []*openapi.Config
		test    func(t *testing.T, result *openapi.Config)
	}{
		{
			name: "patch schema",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0"),
				openapitest.NewConfig("1.0", openapitest.WithComponentSchema("foo", schematest.New("number"))),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Equal(t, 1, result.Components.Schemas.Len())
				require.Equal(t, "number", result.Components.Schemas.Get("foo").Value.Type)
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
