package openapi_test

import (
	"io"
	"mokapi/export/bruno"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	"mokapi/providers/openapi/schema/schematest"
	"mokapi/schema/json/generator"
	"mokapi/version"
	"net/http"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestConfig_ExportBruno(t *testing.T) {
	testcases := []struct {
		name string
		cfg  *openapi.Config
		opt  openapi.BrunoExportOptions
		test func(t *testing.T, c bruno.Collection, err error)
	}{
		{
			name: "empty",
			cfg:  &openapi.Config{},
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Equal(t, bruno.Collection{
					Version: &version.Version{Major: 1, Minor: 0, Patch: 0},
					Info:    bruno.Info{},
					Bundled: true,
				}, c)
			},
		},
		{
			name: "with info",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithInfo("foo", "1.0", ""),
				openapitest.WithSummary("foo summary"),
			),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Equal(t, bruno.Collection{
					Version: &version.Version{Major: 1, Minor: 0, Patch: 0},
					Info: bruno.Info{
						Name:    "foo",
						Summary: "foo summary",
						Version: "1.0",
					},
					Bundled: true,
				}, c)
			},
		},
		{
			name: "with contact",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithContact("foo", "https://foo.com", "foo@foo.com"),
			),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Equal(t, bruno.Collection{
					Version: &version.Version{Major: 1, Minor: 0, Patch: 0},
					Info: bruno.Info{
						Authors: []bruno.Author{
							{Name: "foo", Url: "https://foo.com", Email: "foo@foo.com"},
						},
					},
					Bundled: true,
				}, c)
			},
		},
		{
			name: "with one server",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithServer("https://foo.com", "foo description"),
			),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.NotNil(t, c.Config)
				require.Equal(t, []bruno.Environment{
					{
						Name:        "foo-description",
						Description: "foo description",
						Variables: []bruno.Variable{
							{Name: "baseUrl", Value: "https://foo.com"},
						},
					},
				}, c.Config.Environments)
				require.Equal(t, &bruno.RequestDefault{
					Variables: []bruno.Variable{
						{Name: "baseUrl", Value: "https://foo.com"},
					},
				}, c.Request)
			},
		},
		{
			name: "with one server just a path",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithServer("/foo", "foo description"),
			),
			opt: openapi.BrunoExportOptions{BaseUrl: "foo.com"},
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.NotNil(t, c.Config)
				require.Equal(t, []bruno.Environment{
					{
						Name:        "foo-description",
						Description: "foo description",
						Variables: []bruno.Variable{
							{Name: "baseUrl", Value: "http://foo.com/foo"},
						},
					},
				}, c.Config.Environments)
				require.Equal(t, &bruno.RequestDefault{
					Variables: []bruno.Variable{
						{Name: "baseUrl", Value: "http://foo.com/foo"},
					},
				}, c.Request)
			},
		},
		{
			name: "with one server port",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithServer("http://:8080/foo", "foo description"),
			),
			opt: openapi.BrunoExportOptions{BaseUrl: "foo.com"},
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.NotNil(t, c.Config)
				require.Equal(t, []bruno.Environment{
					{
						Name:        "foo-description",
						Description: "foo description",
						Variables: []bruno.Variable{
							{Name: "baseUrl", Value: "http://foo.com:8080/foo"},
						},
					},
				}, c.Config.Environments)
				require.Equal(t, &bruno.RequestDefault{
					Variables: []bruno.Variable{
						{Name: "baseUrl", Value: "http://foo.com:8080/foo"},
					},
				}, c.Request)
			},
		},
		{
			name: "with one server host, port and path but no scheme",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithServer("http://foo.api:8080/foo", "foo description"),
			),
			opt: openapi.BrunoExportOptions{BaseUrl: "foo.com"},
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.NotNil(t, c.Config)
				require.Equal(t, []bruno.Environment{
					{
						Name:        "foo-description",
						Description: "foo description",
						Variables: []bruno.Variable{
							{Name: "baseUrl", Value: "http://foo.api:8080/foo"},
						},
					},
				}, c.Config.Environments)
				require.Equal(t, &bruno.RequestDefault{
					Variables: []bruno.Variable{
						{Name: "baseUrl", Value: "http://foo.api:8080/foo"},
					},
				}, c.Request)
			},
		},
		{
			name: "with one server with scheme",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithServer("https://foo.api/foo", "foo description"),
			),
			opt: openapi.BrunoExportOptions{BaseUrl: "foo.com"},
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.NotNil(t, c.Config)
				require.Equal(t, []bruno.Environment{
					{
						Name:        "foo-description",
						Description: "foo description",
						Variables: []bruno.Variable{
							{Name: "baseUrl", Value: "https://foo.api/foo"},
						},
					},
				}, c.Config.Environments)
				require.Equal(t, &bruno.RequestDefault{
					Variables: []bruno.Variable{
						{Name: "baseUrl", Value: "https://foo.api/foo"},
					},
				}, c.Request)
			},
		},
		{
			name: "with two server different domains",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithServer("https://foo.com", "foo description"),
				openapitest.WithServer("https://bar.com", "bar description"),
			),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.NotNil(t, c.Config)
				require.Equal(t, []bruno.Environment{
					{
						Name:        "foo-description",
						Description: "foo description",
						Variables: []bruno.Variable{
							{Name: "baseUrl", Value: "https://foo.com"},
						},
					},
					{
						Name:        "bar-description",
						Description: "bar description",
						Variables: []bruno.Variable{
							{Name: "baseUrl", Value: "https://bar.com"},
						},
					},
				}, c.Config.Environments)
				require.Equal(t, &bruno.RequestDefault{
					Variables: []bruno.Variable{
						{Name: "baseUrl", Value: "https://foo.com"},
					},
				}, c.Request)
			},
		},
		{
			name: "with two server different domains not description",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithServer("https://foo.com", ""),
				openapitest.WithServer("https://bar.com", ""),
			),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.NotNil(t, c.Config)
				require.Equal(t, []bruno.Environment{
					{
						Name: "foo.com",
						Variables: []bruno.Variable{
							{Name: "baseUrl", Value: "https://foo.com"},
						},
					},
					{
						Name: "bar.com",
						Variables: []bruno.Variable{
							{Name: "baseUrl", Value: "https://bar.com"},
						},
					},
				}, c.Config.Environments)
				require.Equal(t, &bruno.RequestDefault{
					Variables: []bruno.Variable{
						{Name: "baseUrl", Value: "https://foo.com"},
					},
				}, c.Request)
			},
		},
		{
			name: "with two server same domains different path not description",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithServer("/foo/bar", ""),
				openapitest.WithServer("/", ""),
			),
			opt: openapi.BrunoExportOptions{BaseUrl: "foo.com"},
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.NotNil(t, c.Config)
				require.Equal(t, []bruno.Environment{
					{
						Name: "foo.com-foo-bar",
						Variables: []bruno.Variable{
							{Name: "baseUrl", Value: "http://foo.com/foo/bar"},
						},
					},
					{
						Name: "foo.com",
						Variables: []bruno.Variable{
							{Name: "baseUrl", Value: "http://foo.com"},
						},
					},
				}, c.Config.Environments)
				require.Equal(t, &bruno.RequestDefault{
					Variables: []bruno.Variable{
						{Name: "baseUrl", Value: "http://foo.com/foo/bar"},
					},
				}, c.Request)
			},
		},
		{
			name: "with path but no operation",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithPath("/foo"),
			),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Len(t, c.Items, 0)
			},
		},
		{
			name: "with path and GET operation",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithPath("/foo",
					openapitest.WithOperation(http.MethodGet),
				),
			),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Len(t, c.Items, 1)
				item := c.Items[0].(bruno.HttpItem)
				require.Equal(t, &bruno.HttpInfo{
					Name:        "GET /foo",
					Description: "",
					Type:        "http",
					Sequence:    1,
				}, item.Info)
				require.Equal(t, &bruno.HttpDetail{
					Method: http.MethodGet,
					Url:    "{{baseUrl}}/foo",
				}, item.Http)
			},
		},
		{
			name: "use operation summary",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithPath("/pet",
					openapitest.WithOperation(
						http.MethodPut,
						openapitest.WithOperationSummary("Update an existing pet"),
						openapitest.WithOperationDescription("Update an existing pet by Id."),
					),
				),
			),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Len(t, c.Items, 1)
				item := c.Items[0].(bruno.HttpItem)
				require.Equal(t, &bruno.HttpInfo{
					Name:        "PUT Update an existing pet",
					Description: "Update an existing pet by Id.",
					Type:        "http",
					Sequence:    1,
				}, item.Info)
				require.Equal(t, &bruno.HttpDetail{
					Method: http.MethodPut,
					Url:    "{{baseUrl}}/pet",
				}, item.Http)
			},
		},
		{
			name: "use path summary and description as fallback",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithPath("/pet",
					openapitest.WithPathSummary("Update an existing pet"),
					openapitest.WithPathDescription("Update an existing pet by Id."),
					openapitest.WithOperation(
						http.MethodPut,
					),
				),
			),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Len(t, c.Items, 1)
				item := c.Items[0].(bruno.HttpItem)
				require.Equal(t, &bruno.HttpInfo{
					Name:        "PUT Update an existing pet",
					Description: "Update an existing pet by Id.",
					Type:        "http",
					Sequence:    1,
				}, item.Info)
				require.Equal(t, &bruno.HttpDetail{
					Method: http.MethodPut,
					Url:    "{{baseUrl}}/pet",
				}, item.Http)
			},
		},
		{
			name: "with path and GET operation description and path parameter",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithPath("/products/{name}",
					openapitest.WithOperation(
						http.MethodGet,
						openapitest.WithOperationDescription("operation description"),
						openapitest.WithOperationParam(
							"name",
							true,
							openapitest.WithParamSchema(schematest.New("string")),
							openapitest.WithParamInfo("param description"),
						),
					),
				),
			),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Len(t, c.Items, 1)
				item := c.Items[0].(bruno.HttpItem)
				require.Equal(t, &bruno.HttpInfo{
					Name:        "GET /products/{name}",
					Description: "operation description",
					Type:        "http",
					Sequence:    1,
				}, item.Info)
				require.Equal(t, &bruno.HttpDetail{
					Method: http.MethodGet,
					Url:    "{{baseUrl}}/products/:name",
					Params: []bruno.HttpRequestParam{
						{
							Name:        "name",
							Value:       "Indispensable%20Trunk",
							Description: "param description",
							Type:        "path",
							Disabled:    false,
						},
					},
				}, item.Http)
			},
		},
		{
			name: "with path and GET operation description and path parameter defined on path",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithPath("/products/{name}",
					openapitest.WithOperation(
						http.MethodGet,
						openapitest.WithOperationDescription("operation description"),
					),
					openapitest.WithPathParam(
						"name",
						openapitest.WithParamSchema(schematest.New("string")),
						openapitest.WithParamInfo("param description"),
					),
				),
			),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Len(t, c.Items, 1)
				item := c.Items[0].(bruno.HttpItem)
				require.Equal(t, &bruno.HttpInfo{
					Name:        "GET /products/{name}",
					Description: "operation description",
					Type:        "http",
					Sequence:    1,
				}, item.Info)
				require.Equal(t, &bruno.HttpDetail{
					Method: http.MethodGet,
					Url:    "{{baseUrl}}/products/:name",
					Params: []bruno.HttpRequestParam{
						{
							Name:        "name",
							Value:       "Indispensable%20Trunk",
							Description: "param description",
							Type:        "path",
							Disabled:    false,
						},
					},
				}, item.Http)
			},
		},
		{
			name: "with path and GET operation description and query parameter",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithPath("/products",
					openapitest.WithOperation(
						http.MethodGet,
						openapitest.WithQueryParam(
							"name",
							false,
							openapitest.WithParamSchema(schematest.New("string")),
						),
					),
				),
			),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Len(t, c.Items, 1)
				item := c.Items[0].(bruno.HttpItem)
				require.Equal(t, &bruno.HttpInfo{
					Name:     "GET /products",
					Type:     "http",
					Sequence: 1,
				}, item.Info)
				require.Equal(t, &bruno.HttpDetail{
					Method: http.MethodGet,
					Url:    "{{baseUrl}}/products",
					Params: []bruno.HttpRequestParam{
						{
							Name:     "name",
							Value:    "Indispensable Trunk",
							Type:     "query",
							Disabled: true,
						},
					},
				}, item.Http)
			},
		},
		{
			name: "with path and GET operation description and header parameter",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithPath("/products",
					openapitest.WithOperation(
						http.MethodGet,
						openapitest.WithHeaderParam(
							"name",
							false,
							openapitest.WithParamSchema(schematest.New("string")),
						),
					),
				),
			),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Len(t, c.Items, 1)
				item := c.Items[0].(bruno.HttpItem)
				require.Equal(t, &bruno.HttpInfo{
					Name:     "GET /products",
					Type:     "http",
					Sequence: 1,
				}, item.Info)
				require.Equal(t, &bruno.HttpDetail{
					Method: http.MethodGet,
					Url:    "{{baseUrl}}/products",
					Headers: []bruno.HttpRequestHeader{
						{
							Name:     "name",
							Value:    "Indispensable Trunk",
							Disabled: true,
						},
					},
				}, item.Http)
			},
		},
		{
			name: "with path and POST operation description and request body",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithPath("/products",
					openapitest.WithOperation(
						http.MethodPost,
						openapitest.WithRequestBody("request body description", true,
							openapitest.WithRequestContent("application/json",
								openapitest.WithSchema(schematest.New("object",
									schematest.WithProperty("name", schematest.New("string")),
								)),
							),
						),
					),
				),
			),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Len(t, c.Items, 1)
				item := c.Items[0].(bruno.HttpItem)
				require.Equal(t, &bruno.HttpInfo{
					Name:     "POST /products",
					Type:     "http",
					Sequence: 1,
				}, item.Info)
				require.Equal(t, &bruno.HttpDetail{
					Method: http.MethodPost,
					Url:    "{{baseUrl}}/products",
					Body: &bruno.HttpRequestBody{
						Body: &bruno.HttpRequestBodyRaw{
							Type: "json",
							Data: `{"name":"Indispensable Trunk"}`,
						},
					},
				}, item.Http)
			},
		},
		{
			name: "with path and POST operation with request body json and plain text",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithPath("/products",
					openapitest.WithOperation(
						http.MethodPost,
						openapitest.WithRequestBody("", true,
							openapitest.WithRequestContent("application/json",
								openapitest.WithSchema(schematest.New("object",
									schematest.WithProperty("name", schematest.New("string")),
								)),
							),
							openapitest.WithRequestContent("text/plain",
								openapitest.WithSchema(schematest.New("string")),
							),
						),
					),
				),
			),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Len(t, c.Items, 1)
				item := c.Items[0].(bruno.HttpItem)
				require.Equal(t, &bruno.HttpInfo{
					Name:     "POST /products",
					Type:     "http",
					Sequence: 1,
				}, item.Info)
				require.Equal(t, &bruno.HttpDetail{
					Method: http.MethodPost,
					Url:    "{{baseUrl}}/products",
					Body: &bruno.HttpRequestBody{
						Variant: []bruno.HttpRequestBodyVariant{
							{
								Title:    "application/json",
								Selected: true,
								Body: bruno.HttpRequestBodyRaw{
									Type: "json",
									Data: `{"name":"Indispensable Trunk"}`,
								},
							},
							{
								Title:    "text/plain",
								Selected: false,
								Body: bruno.HttpRequestBodyRaw{
									Type: "text",
									Data: "yQtLpCUeQyta",
								},
							},
						},
					},
				}, item.Http)
			},
		},
		{
			name: "with path and GET operation using tag",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithTag("foo", "summary", "description"),
				openapitest.WithPath("/foo",
					openapitest.WithOperation(http.MethodGet, openapitest.WithOperationTags("foo")),
				),
			),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Len(t, c.Items, 1)
				folder := c.Items[0].(bruno.FolderItem)
				require.Len(t, folder.Items, 1)
				require.Equal(t, &bruno.FolderInfo{
					Name:        "foo",
					Description: "summary",
					Type:        "folder",
					Sequence:    1,
				}, folder.Info)
				item := folder.Items[0].(bruno.HttpItem)
				require.Equal(t, &bruno.HttpInfo{
					Name:        "GET /foo",
					Description: "",
					Type:        "http",
					Sequence:    1,
				}, item.Info)
			},
		},
		{
			name: "with path and GET operation using tag and summary",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithTag("foo", "summary", ""),
				openapitest.WithPath("/foo",
					openapitest.WithOperation(http.MethodGet, openapitest.WithOperationTags("foo")),
				),
			),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Len(t, c.Items, 1)
				folder := c.Items[0].(bruno.FolderItem)
				require.Len(t, folder.Items, 1)
				require.Equal(t, &bruno.FolderInfo{
					Name:        "foo",
					Description: "summary",
					Type:        "folder",
					Sequence:    1,
				}, folder.Info)
			},
		},
		{
			name: "with path and GET operation using two tags one not used",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithTag("foo", "", "description"),
				openapitest.WithTag("bar", "", "not used"),
				openapitest.WithPath("/foo",
					openapitest.WithOperation(http.MethodGet, openapitest.WithOperationTags("foo")),
				),
			),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Len(t, c.Items, 1)
				folder := c.Items[0].(bruno.FolderItem)
				require.Len(t, folder.Items, 1)
				require.Equal(t, &bruno.FolderInfo{
					Name:        "foo",
					Description: "description",
					Type:        "folder",
					Sequence:    1,
				}, folder.Info)
			},
		},
		{
			name: "with path and two GET operation using two tags, only first tag is used",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithTag("foo", "", "foo description"),
				openapitest.WithTag("bar", "", "bar description"),
				openapitest.WithPath("/foo",
					openapitest.WithOperation(http.MethodGet, openapitest.WithOperationTags("foo", "bar")),
				),
				openapitest.WithPath("/bar",
					openapitest.WithOperation(http.MethodGet, openapitest.WithOperationTags("bar", "foo")),
				),
			),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Len(t, c.Items, 2)
				foo := c.Items[0].(bruno.FolderItem)
				require.Len(t, foo.Items, 1)
				require.Equal(t, &bruno.FolderInfo{
					Name:        "foo",
					Description: "foo description",
					Type:        "folder",
					Sequence:    1,
				}, foo.Info)
				require.Equal(t, "GET /foo", foo.Items[0].(bruno.HttpItem).Info.Name)
				bar := c.Items[1].(bruno.FolderItem)
				require.Len(t, bar.Items, 1)
				require.Equal(t, &bruno.FolderInfo{
					Name:        "bar",
					Description: "bar description",
					Type:        "folder",
					Sequence:    2,
				}, bar.Info)
				require.Equal(t, "GET /bar", bar.Items[0].(bruno.HttpItem).Info.Name)
			},
		},
		{
			name: "using tag tree",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithTag("foo", "", "foo description"),
				openapitest.WithTag("bar", "", "bar description",
					openapitest.WithTagParent("foo"),
				),
				openapitest.WithPath("/foo",
					openapitest.WithOperation(http.MethodGet, openapitest.WithOperationTags("foo")),
				),
				openapitest.WithPath("/bar",
					openapitest.WithOperation(http.MethodGet, openapitest.WithOperationTags("bar")),
				),
			),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Len(t, c.Items, 1)
				require.IsType(t, bruno.FolderItem{}, c.Items[0])
				foo := c.Items[0].(bruno.FolderItem)
				require.Len(t, foo.Items, 2)
				require.Equal(t, &bruno.FolderInfo{
					Name:        "foo",
					Description: "foo description",
					Type:        "folder",
					Sequence:    1,
				}, foo.Info)

				bar := foo.Items[0].(bruno.FolderItem)
				require.Len(t, bar.Items, 1)
				require.Equal(t, &bruno.FolderInfo{
					Name:        "bar",
					Description: "bar description",
					Type:        "folder",
					Sequence:    1,
				}, bar.Info)

				require.Equal(t, "GET /foo", foo.Items[1].(bruno.HttpItem).Info.Name)
				require.Equal(t, "GET /bar", bar.Items[0].(bruno.HttpItem).Info.Name)
			},
		},
		{
			name: "using path as arrangement create folder but not empty ones",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithPath("/pet",
					openapitest.WithOperation(http.MethodPut),
					openapitest.WithOperation(http.MethodPost),
				),
				openapitest.WithPath("/pet/findByStatus",
					openapitest.WithOperation(http.MethodGet),
				),
				openapitest.WithPath("/pet/findByTags",
					openapitest.WithOperation(http.MethodGet),
				),
				openapitest.WithPath("/foo/user/login",
					openapitest.WithOperation(http.MethodGet),
				),
				openapitest.WithPath("/bar/store/inventory",
					openapitest.WithOperation(http.MethodGet),
				),
				openapitest.WithPath("/bar/yuh",
					openapitest.WithOperation(http.MethodGet),
				),
			),
			opt: openapi.BrunoExportOptions{FolderArrangement: openapi.PathFolderArrangement},
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Len(t, c.Items, 3)
				pet := c.Items[0].(bruno.FolderItem)
				require.Equal(t, "pet", pet.Info.Name)
				require.Equal(t, "PUT pet", pet.Items[0].(bruno.HttpItem).Info.Name)
				require.Equal(t, "POST pet", pet.Items[1].(bruno.HttpItem).Info.Name)
				require.Equal(t, "GET findByStatus", pet.Items[2].(bruno.HttpItem).Info.Name)
				require.Equal(t, "GET findByTags", pet.Items[3].(bruno.HttpItem).Info.Name)

				foo := c.Items[1].(bruno.FolderItem)
				require.Equal(t, "foo/user", foo.Info.Name)
				require.Equal(t, "GET login", foo.Items[0].(bruno.HttpItem).Info.Name)

				bar := c.Items[2].(bruno.FolderItem)
				require.Equal(t, "bar", bar.Info.Name)
				require.Equal(t, "GET yuh", bar.Items[1].(bruno.HttpItem).Info.Name)

				store := bar.Items[0].(bruno.FolderItem)
				require.Equal(t, "store", store.Info.Name)
				require.Equal(t, "GET inventory", store.Items[0].(bruno.HttpItem).Info.Name)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			logrus.SetOutput(io.Discard)
			generator.Seed(12345)

			c, err := tc.cfg.ExportBruno(tc.opt)
			tc.test(t, c, err)
		})
	}
}
