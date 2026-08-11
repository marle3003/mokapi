package openapi_test

import (
	"io"
	"mokapi/export/bruno"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	"mokapi/providers/openapi/schema/schematest"
	"mokapi/schema/json/generator"
	"net/http"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestConfig_ExportBruno_Parameters(t *testing.T) {
	testcases := []struct {
		name    string
		cfg     *openapi.Config
		baseUrl string
		test    func(t *testing.T, c bruno.Collection, err error)
	}{
		{
			name: "array query parameter",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithPath("/foo",
					openapitest.WithOperation(
						http.MethodGet,
						openapitest.WithQueryParam(
							"status",
							false,
							openapitest.WithStyle("form"),
							openapitest.WithParamSchema(
								schematest.New(
									"array",
									schematest.WithItems("string", schematest.WithEnum([]any{"available", "sold"}))))),
					),
				),
			),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Len(t, c.Items, 1)
				item := c.Items[0].(bruno.HttpItem)
				require.Equal(t, "{{baseUrl}}/foo", item.Http.Url)
				require.Len(t, item.Http.Params, 2)
				require.Equal(t, bruno.HttpRequestParam{
					Name:     "status",
					Value:    "sold",
					Type:     "query",
					Disabled: true,
				}, item.Http.Params[0])
				require.Equal(t, bruno.HttpRequestParam{
					Name:     "status",
					Value:    "available",
					Type:     "query",
					Disabled: true,
				}, item.Http.Params[1])
			},
		},
		{
			name: "array query parameter required",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithPath("/foo",
					openapitest.WithOperation(
						http.MethodGet,
						openapitest.WithQueryParam(
							"status",
							true,
							openapitest.WithStyle("form"),
							openapitest.WithParamSchema(
								schematest.New(
									"array",
									schematest.WithItems("string", schematest.WithEnum([]any{"available", "sold"}))))),
					),
				),
			),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Len(t, c.Items, 1)
				item := c.Items[0].(bruno.HttpItem)
				require.Equal(t, "{{baseUrl}}/foo?status=sold&status=available", item.Http.Url)
				require.Len(t, item.Http.Params, 2)
				require.Equal(t, bruno.HttpRequestParam{
					Name:     "status",
					Value:    "sold",
					Type:     "query",
					Disabled: false,
				}, item.Http.Params[0])
				require.Equal(t, bruno.HttpRequestParam{
					Name:     "status",
					Value:    "available",
					Type:     "query",
					Disabled: false,
				}, item.Http.Params[1])
			},
		},
		{
			name: "array query parameter form not exploded",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithPath("/foo",
					openapitest.WithOperation(
						http.MethodGet,
						openapitest.WithQueryParam(
							"status",
							false,
							openapitest.WithStyle("form"),
							openapitest.WithExplode(false),
							openapitest.WithParamSchema(
								schematest.New(
									"array",
									schematest.WithItems("string", schematest.WithEnum([]any{"available", "sold"}))))),
					),
				),
			),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Len(t, c.Items, 1)
				item := c.Items[0].(bruno.HttpItem)
				require.Equal(t, "{{baseUrl}}/foo", item.Http.Url)
				require.Len(t, item.Http.Params, 1)
				require.Equal(t, bruno.HttpRequestParam{
					Name:     "status",
					Value:    "sold,available",
					Type:     "query",
					Disabled: true,
				}, item.Http.Params[0])
			},
		},
		{
			name: "array query parameter space delimited",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithPath("/foo",
					openapitest.WithOperation(
						http.MethodGet,
						openapitest.WithQueryParam(
							"status",
							false,
							openapitest.WithStyle("spaceDelimited"),
							openapitest.WithParamSchema(
								schematest.New(
									"array",
									schematest.WithItems("string", schematest.WithEnum([]any{"available", "sold"}))))),
					),
				),
			),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Len(t, c.Items, 1)
				item := c.Items[0].(bruno.HttpItem)
				require.Equal(t, "{{baseUrl}}/foo", item.Http.Url)
				require.Len(t, item.Http.Params, 1)
				require.Equal(t, bruno.HttpRequestParam{
					Name:     "status",
					Value:    "sold available",
					Type:     "query",
					Disabled: true,
				}, item.Http.Params[0])
			},
		},
		{
			name: "array query parameter space delimited",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithPath("/foo",
					openapitest.WithOperation(
						http.MethodGet,
						openapitest.WithQueryParam(
							"status",
							false,
							openapitest.WithStyle("spaceDelimited"),
							openapitest.WithParamSchema(
								schematest.New(
									"array",
									schematest.WithItems("string", schematest.WithEnum([]any{"available", "sold"}))))),
					),
				),
			),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Len(t, c.Items, 1)
				item := c.Items[0].(bruno.HttpItem)
				require.Equal(t, "{{baseUrl}}/foo", item.Http.Url)
				require.Len(t, item.Http.Params, 1)
				require.Equal(t, bruno.HttpRequestParam{
					Name:     "status",
					Value:    "sold available",
					Type:     "query",
					Disabled: true,
				}, item.Http.Params[0])
			},
		},
		{
			name: "array query parameter pipe delimited",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithPath("/foo",
					openapitest.WithOperation(
						http.MethodGet,
						openapitest.WithQueryParam(
							"status",
							false,
							openapitest.WithStyle("pipeDelimited"),
							openapitest.WithParamSchema(
								schematest.New(
									"array",
									schematest.WithItems("string", schematest.WithEnum([]any{"available", "sold"}))))),
					),
				),
			),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Len(t, c.Items, 1)
				item := c.Items[0].(bruno.HttpItem)
				require.Equal(t, "{{baseUrl}}/foo", item.Http.Url)
				require.Len(t, item.Http.Params, 1)
				require.Equal(t, bruno.HttpRequestParam{
					Name:     "status",
					Value:    "sold|available",
					Type:     "query",
					Disabled: true,
				}, item.Http.Params[0])
			},
		},
		{
			name: "object query parameter form exploded",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithPath("/foo",
					openapitest.WithOperation(
						http.MethodGet,
						openapitest.WithQueryParam(
							"status",
							true,
							openapitest.WithExplode(true),
							openapitest.WithParamSchema(
								schematest.New(
									"object",
									schematest.WithProperty("foo", schematest.New("string")),
									schematest.WithProperty("bar", schematest.New("string",
										schematest.WithDescription("prop description"),
									)),
									schematest.WithRequired("foo"),
								),
							),
						),
					),
				),
			),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Len(t, c.Items, 1)
				item := c.Items[0].(bruno.HttpItem)
				require.Equal(t, "{{baseUrl}}/foo?foo=Cl", item.Http.Url)
				require.Len(t, item.Http.Params, 2)
				require.Equal(t, bruno.HttpRequestParam{
					Name:     "foo",
					Value:    "Cl",
					Type:     "query",
					Disabled: false,
				}, item.Http.Params[0])
				require.Equal(t, bruno.HttpRequestParam{
					Name:        "bar",
					Value:       "ompxjsbIr",
					Type:        "query",
					Description: "prop description",
					Disabled:    true,
				}, item.Http.Params[1])
			},
		},
		{
			name: "object query parameter form not exploded",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithPath("/foo",
					openapitest.WithOperation(
						http.MethodGet,
						openapitest.WithQueryParam(
							"status",
							false,
							openapitest.WithStyle("form"),
							openapitest.WithExplode(false),
							openapitest.WithParamSchema(
								schematest.New(
									"object",
									schematest.WithProperty("foo", schematest.New("string")),
									schematest.WithProperty("bar", schematest.New("string",
										schematest.WithDescription("prop description"),
									)),
									schematest.WithRequired("foo"),
								),
							),
						),
					),
				),
			),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Len(t, c.Items, 1)
				item := c.Items[0].(bruno.HttpItem)
				require.Equal(t, "{{baseUrl}}/foo", item.Http.Url)
				require.Len(t, item.Http.Params, 1)
				require.Equal(t, bruno.HttpRequestParam{
					Name:     "status",
					Value:    "foo,Cl,bar,ompxjsbIr",
					Type:     "query",
					Disabled: true,
				}, item.Http.Params[0])
			},
		},
		{
			name: "object query parameter deepObject",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithPath("/foo",
					openapitest.WithOperation(
						http.MethodGet,
						openapitest.WithQueryParam(
							"status",
							true,
							openapitest.WithStyle("deepObject"),
							openapitest.WithParamSchema(
								schematest.New(
									"object",
									schematest.WithProperty("foo", schematest.New("string")),
									schematest.WithProperty("bar", schematest.New("string",
										schematest.WithDescription("prop description"),
									)),
									schematest.WithRequired("foo"),
								),
							),
						),
					),
				),
			),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Len(t, c.Items, 1)
				item := c.Items[0].(bruno.HttpItem)
				require.Equal(t, "{{baseUrl}}/foo?status[foo]=Cl", item.Http.Url)
				require.Len(t, item.Http.Params, 2)
				require.Equal(t, bruno.HttpRequestParam{
					Name:     "status[foo]",
					Value:    "Cl",
					Type:     "query",
					Disabled: false,
				}, item.Http.Params[0])
				require.Equal(t, bruno.HttpRequestParam{
					Name:        "status[bar]",
					Value:       "ompxjsbIr",
					Type:        "query",
					Description: "prop description",
					Disabled:    true,
				}, item.Http.Params[1])
			},
		},
		{
			name: "simple path parameter string value",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithPath("/{color}",
					openapitest.WithOperation(
						http.MethodGet,
						openapitest.WithOperationParam(
							"color",
							true,
							openapitest.WithStyle("simple"),
							openapitest.WithParamSchema(
								schematest.New(
									"string",
									schematest.WithEnumValues("blue", "black", "brown"),
								),
							),
						),
					),
				),
			),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Len(t, c.Items, 1)
				item := c.Items[0].(bruno.HttpItem)
				require.Equal(t, "{{baseUrl}}/:color", item.Http.Url)
				require.Len(t, item.Http.Params, 1)
				require.Equal(t, bruno.HttpRequestParam{
					Name:     "color",
					Value:    "brown",
					Type:     "path",
					Disabled: false,
				}, item.Http.Params[0])
			},
		},
		{
			name: "simple path parameter array value",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithPath("/{color}",
					openapitest.WithOperation(
						http.MethodGet,
						openapitest.WithOperationParam(
							"color",
							true,
							openapitest.WithStyle("simple"),
							openapitest.WithParamSchema(
								schematest.New("array",
									schematest.WithItems(
										"string",
										schematest.WithEnumValues("blue", "black", "brown"),
									),
								),
							),
						),
					),
				),
			),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Len(t, c.Items, 1)
				item := c.Items[0].(bruno.HttpItem)
				require.Equal(t, "{{baseUrl}}/:color", item.Http.Url)
				require.Len(t, item.Http.Params, 1)
				require.Equal(t, bruno.HttpRequestParam{
					Name:     "color",
					Value:    "brown,blue",
					Type:     "path",
					Disabled: false,
				}, item.Http.Params[0])
			},
		},
		{
			name: "simple exploded path parameter array value",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithPath("/{color}",
					openapitest.WithOperation(
						http.MethodGet,
						openapitest.WithOperationParam(
							"color",
							true,
							openapitest.WithStyle("simple"),
							openapitest.WithExplode(true),
							openapitest.WithParamSchema(
								schematest.New("array",
									schematest.WithItems(
										"string",
										schematest.WithEnumValues("blue", "black", "brown"),
									),
								),
							),
						),
					),
				),
			),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Len(t, c.Items, 1)
				item := c.Items[0].(bruno.HttpItem)
				require.Equal(t, "{{baseUrl}}/:color", item.Http.Url)
				require.Len(t, item.Http.Params, 1)
				require.Equal(t, bruno.HttpRequestParam{
					Name:     "color",
					Value:    "brown,blue",
					Type:     "path",
					Disabled: false,
				}, item.Http.Params[0])
			},
		},
		{
			name: "simple path parameter object value",
			cfg: func() *openapi.Config {
				color := schematest.New(
					"integer",
					schematest.WithMinimum(0),
					schematest.WithMaximum(255),
				)
				return openapitest.NewConfig("3.2.0",
					openapitest.WithPath("/{color}",
						openapitest.WithOperation(
							http.MethodGet,
							openapitest.WithOperationParam(
								"color",
								true,
								openapitest.WithStyle("simple"),
								openapitest.WithParamSchema(
									schematest.New("object",
										schematest.WithProperty("R", color),
										schematest.WithProperty("G", color),
										schematest.WithProperty("B", color),
										schematest.WithRequired("R", "G", "B"),
									),
								),
							),
						),
					),
				)
			}(),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Len(t, c.Items, 1)
				item := c.Items[0].(bruno.HttpItem)
				require.Equal(t, "{{baseUrl}}/:color", item.Http.Url)
				require.Len(t, item.Http.Params, 1)
				require.Equal(t, bruno.HttpRequestParam{
					Name:     "color",
					Value:    "R,160,G,100,B,83",
					Type:     "path",
					Disabled: false,
				}, item.Http.Params[0])
			},
		},
		{
			name: "simple exploded path parameter array value",
			cfg: func() *openapi.Config {
				color := schematest.New(
					"integer",
					schematest.WithMinimum(0),
					schematest.WithMaximum(255),
				)
				return openapitest.NewConfig("3.2.0",
					openapitest.WithPath("/{color}",
						openapitest.WithOperation(
							http.MethodGet,
							openapitest.WithOperationParam(
								"color",
								true,
								openapitest.WithStyle("simple"),
								openapitest.WithExplode(true),
								openapitest.WithParamSchema(
									schematest.New("object",
										schematest.WithProperty("R", color),
										schematest.WithProperty("G", color),
										schematest.WithProperty("B", color),
										schematest.WithRequired("R", "G", "B"),
									),
								),
							),
						),
					),
				)
			}(),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Len(t, c.Items, 1)
				item := c.Items[0].(bruno.HttpItem)
				require.Equal(t, "{{baseUrl}}/:color", item.Http.Url)
				require.Len(t, item.Http.Params, 1)
				require.Equal(t, bruno.HttpRequestParam{
					Name:     "color",
					Value:    "R=160,G=100,B=83",
					Type:     "path",
					Disabled: false,
				}, item.Http.Params[0])
			},
		},
		{
			name: "matrix path parameter string value",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithPath("/{color}",
					openapitest.WithOperation(
						http.MethodGet,
						openapitest.WithOperationParam(
							"color",
							true,
							openapitest.WithStyle("matrix"),
							openapitest.WithParamSchema(
								schematest.New(
									"string",
									schematest.WithEnumValues("blue", "black", "brown"),
								),
							),
						),
					),
				),
			),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Len(t, c.Items, 1)
				item := c.Items[0].(bruno.HttpItem)
				require.Equal(t, "{{baseUrl}}/:color", item.Http.Url)
				require.Len(t, item.Http.Params, 1)
				require.Equal(t, bruno.HttpRequestParam{
					Name:     "color",
					Value:    ";color=brown",
					Type:     "path",
					Disabled: false,
				}, item.Http.Params[0])
			},
		},
		{
			name: "matrix path parameter array value",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithPath("/{color}",
					openapitest.WithOperation(
						http.MethodGet,
						openapitest.WithOperationParam(
							"color",
							true,
							openapitest.WithStyle("matrix"),
							openapitest.WithParamSchema(
								schematest.New("array",
									schematest.WithItems(
										"string",
										schematest.WithEnumValues("blue", "black", "brown"),
									),
								),
							),
						),
					),
				),
			),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Len(t, c.Items, 1)
				item := c.Items[0].(bruno.HttpItem)
				require.Equal(t, "{{baseUrl}}/:color", item.Http.Url)
				require.Len(t, item.Http.Params, 1)
				require.Equal(t, bruno.HttpRequestParam{
					Name:     "color",
					Value:    ";color=brown,blue",
					Type:     "path",
					Disabled: false,
				}, item.Http.Params[0])
			},
		},
		{
			name: "matrix exploded path parameter array value",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithPath("/{color}",
					openapitest.WithOperation(
						http.MethodGet,
						openapitest.WithOperationParam(
							"color",
							true,
							openapitest.WithStyle("matrix"),
							openapitest.WithExplode(true),
							openapitest.WithParamSchema(
								schematest.New("array",
									schematest.WithItems(
										"string",
										schematest.WithEnumValues("blue", "black", "brown"),
									),
								),
							),
						),
					),
				),
			),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Len(t, c.Items, 1)
				item := c.Items[0].(bruno.HttpItem)
				require.Equal(t, "{{baseUrl}}/:color", item.Http.Url)
				require.Len(t, item.Http.Params, 1)
				require.Equal(t, bruno.HttpRequestParam{
					Name:     "color",
					Value:    ";color=brown;color=blue",
					Type:     "path",
					Disabled: false,
				}, item.Http.Params[0])
			},
		},
		{
			name: "matrix path parameter object value",
			cfg: func() *openapi.Config {
				color := schematest.New(
					"integer",
					schematest.WithMinimum(0),
					schematest.WithMaximum(255),
				)
				return openapitest.NewConfig("3.2.0",
					openapitest.WithPath("/{color}",
						openapitest.WithOperation(
							http.MethodGet,
							openapitest.WithOperationParam(
								"color",
								true,
								openapitest.WithStyle("matrix"),
								openapitest.WithParamSchema(
									schematest.New("object",
										schematest.WithProperty("R", color),
										schematest.WithProperty("G", color),
										schematest.WithProperty("B", color),
										schematest.WithRequired("R", "G", "B"),
									),
								),
							),
						),
					),
				)
			}(),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Len(t, c.Items, 1)
				item := c.Items[0].(bruno.HttpItem)
				require.Equal(t, "{{baseUrl}}/:color", item.Http.Url)
				require.Len(t, item.Http.Params, 1)
				require.Equal(t, bruno.HttpRequestParam{
					Name:     "color",
					Value:    ";color=R,160,G,100,B,83",
					Type:     "path",
					Disabled: false,
				}, item.Http.Params[0])
			},
		},
		{
			name: "matrix exploded path parameter array value",
			cfg: func() *openapi.Config {
				color := schematest.New(
					"integer",
					schematest.WithMinimum(0),
					schematest.WithMaximum(255),
				)
				return openapitest.NewConfig("3.2.0",
					openapitest.WithPath("/{color}",
						openapitest.WithOperation(
							http.MethodGet,
							openapitest.WithOperationParam(
								"color",
								true,
								openapitest.WithStyle("matrix"),
								openapitest.WithExplode(true),
								openapitest.WithParamSchema(
									schematest.New("object",
										schematest.WithProperty("R", color),
										schematest.WithProperty("G", color),
										schematest.WithProperty("B", color),
										schematest.WithRequired("R", "G", "B"),
									),
								),
							),
						),
					),
				)
			}(),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Len(t, c.Items, 1)
				item := c.Items[0].(bruno.HttpItem)
				require.Equal(t, "{{baseUrl}}/:color", item.Http.Url)
				require.Len(t, item.Http.Params, 1)
				require.Equal(t, bruno.HttpRequestParam{
					Name:     "color",
					Value:    ";R=160;G=100;B=83",
					Type:     "path",
					Disabled: false,
				}, item.Http.Params[0])
			},
		},
		{
			name: "label path parameter string value",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithPath("/{color}",
					openapitest.WithOperation(
						http.MethodGet,
						openapitest.WithOperationParam(
							"color",
							true,
							openapitest.WithStyle("label"),
							openapitest.WithParamSchema(
								schematest.New(
									"string",
									schematest.WithEnumValues("blue", "black", "brown"),
								),
							),
						),
					),
				),
			),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Len(t, c.Items, 1)
				item := c.Items[0].(bruno.HttpItem)
				require.Equal(t, "{{baseUrl}}/:color", item.Http.Url)
				require.Len(t, item.Http.Params, 1)
				require.Equal(t, bruno.HttpRequestParam{
					Name:     "color",
					Value:    ".brown",
					Type:     "path",
					Disabled: false,
				}, item.Http.Params[0])
			},
		},
		{
			name: "label path parameter array value",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithPath("/{color}",
					openapitest.WithOperation(
						http.MethodGet,
						openapitest.WithOperationParam(
							"color",
							true,
							openapitest.WithStyle("label"),
							openapitest.WithParamSchema(
								schematest.New("array",
									schematest.WithItems(
										"string",
										schematest.WithEnumValues("blue", "black", "brown"),
									),
								),
							),
						),
					),
				),
			),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Len(t, c.Items, 1)
				item := c.Items[0].(bruno.HttpItem)
				require.Equal(t, "{{baseUrl}}/:color", item.Http.Url)
				require.Len(t, item.Http.Params, 1)
				require.Equal(t, bruno.HttpRequestParam{
					Name:     "color",
					Value:    ".brown,blue",
					Type:     "path",
					Disabled: false,
				}, item.Http.Params[0])
			},
		},
		{
			name: "label exploded path parameter array value",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithPath("/{color}",
					openapitest.WithOperation(
						http.MethodGet,
						openapitest.WithOperationParam(
							"color",
							true,
							openapitest.WithStyle("label"),
							openapitest.WithExplode(true),
							openapitest.WithParamSchema(
								schematest.New("array",
									schematest.WithItems(
										"string",
										schematest.WithEnumValues("blue", "black", "brown"),
									),
								),
							),
						),
					),
				),
			),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Len(t, c.Items, 1)
				item := c.Items[0].(bruno.HttpItem)
				require.Equal(t, "{{baseUrl}}/:color", item.Http.Url)
				require.Len(t, item.Http.Params, 1)
				require.Equal(t, bruno.HttpRequestParam{
					Name:     "color",
					Value:    ".brown.blue",
					Type:     "path",
					Disabled: false,
				}, item.Http.Params[0])
			},
		},
		{
			name: "label path parameter object value",
			cfg: func() *openapi.Config {
				color := schematest.New(
					"integer",
					schematest.WithMinimum(0),
					schematest.WithMaximum(255),
				)
				return openapitest.NewConfig("3.2.0",
					openapitest.WithPath("/{color}",
						openapitest.WithOperation(
							http.MethodGet,
							openapitest.WithOperationParam(
								"color",
								true,
								openapitest.WithStyle("label"),
								openapitest.WithParamSchema(
									schematest.New("object",
										schematest.WithProperty("R", color),
										schematest.WithProperty("G", color),
										schematest.WithProperty("B", color),
										schematest.WithRequired("R", "G", "B"),
									),
								),
							),
						),
					),
				)
			}(),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Len(t, c.Items, 1)
				item := c.Items[0].(bruno.HttpItem)
				require.Equal(t, "{{baseUrl}}/:color", item.Http.Url)
				require.Len(t, item.Http.Params, 1)
				require.Equal(t, bruno.HttpRequestParam{
					Name:     "color",
					Value:    ".R,160,G,100,B,83",
					Type:     "path",
					Disabled: false,
				}, item.Http.Params[0])
			},
		},
		{
			name: "label exploded path parameter array value",
			cfg: func() *openapi.Config {
				color := schematest.New(
					"integer",
					schematest.WithMinimum(0),
					schematest.WithMaximum(255),
				)
				return openapitest.NewConfig("3.2.0",
					openapitest.WithPath("/{color}",
						openapitest.WithOperation(
							http.MethodGet,
							openapitest.WithOperationParam(
								"color",
								true,
								openapitest.WithStyle("label"),
								openapitest.WithExplode(true),
								openapitest.WithParamSchema(
									schematest.New("object",
										schematest.WithProperty("R", color),
										schematest.WithProperty("G", color),
										schematest.WithProperty("B", color),
										schematest.WithRequired("R", "G", "B"),
									),
								),
							),
						),
					),
				)
			}(),
			test: func(t *testing.T, c bruno.Collection, err error) {
				require.NoError(t, err)
				require.Len(t, c.Items, 1)
				item := c.Items[0].(bruno.HttpItem)
				require.Equal(t, "{{baseUrl}}/:color", item.Http.Url)
				require.Len(t, item.Http.Params, 1)
				require.Equal(t, bruno.HttpRequestParam{
					Name:     "color",
					Value:    ".R=160.G=100.B=83",
					Type:     "path",
					Disabled: false,
				}, item.Http.Params[0])
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			logrus.SetOutput(io.Discard)
			generator.Seed(12345)

			c, err := tc.cfg.ExportBruno(tc.baseUrl)
			tc.test(t, c, err)
		})
	}
}
