package openapi_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	"mokapi/providers/openapi/schema"
	"mokapi/providers/openapi/schema/schematest"
	"testing"
)

func Test_Parse(t *testing.T) {
	testcases := []struct {
		name string
		c    *openapi.Config
		test func(t *testing.T, c *openapi.Config, err error)
	}{
		{
			name: "schemas reference",
			c: openapitest.NewConfig("3.1",
				openapitest.WithComponentSchema("Foo", schematest.New("array", schematest.WithItemsRef("#/components/schemas/Bar"))),
				openapitest.WithComponentSchema("Bar", schematest.New("string")),
			),
			test: func(t *testing.T, c *openapi.Config, err error) {
				require.NoError(t, err)
				foo := c.Components.Schemas.Get("Foo").Value
				require.Equal(t, "array", foo.Type.String())
				bar := c.Components.Schemas.Get("Bar").Value
				require.Equal(t, "string", bar.Type.String())
				// reference should point to same schema
				require.Equal(t, bar, foo.Items.Value)
			},
		},
		{
			name: "response schema reference should point to same objects",
			c: openapitest.NewConfig("3.1",
				openapitest.WithPath("/foo", openapitest.NewPath(
					openapitest.WithOperation("get", openapitest.NewOperation(
						openapitest.WithResponse(200, openapitest.WithContent("application/json",
							&openapi.MediaType{
								Schema: &schema.Ref{Reference: dynamic.Reference{Ref: "#/components/schemas/Foo"}},
							},
						)),
					)),
				)),
				openapitest.WithComponentSchema("Foo", schematest.New("array", schematest.WithItemsRef("#/components/schemas/Bar"))),
				openapitest.WithComponentSchema("Bar", schematest.New("string")),
			),
			test: func(t *testing.T, c *openapi.Config, err error) {
				require.NoError(t, err)
				response := c.Paths["/foo"].Value.Get.Responses.GetResponse(200).Content["application/json"].Schema
				require.NotNil(t, response.Value)

				foo := c.Components.Schemas.Get("Foo").Value
				require.Equal(t, foo, response.Value)
				bar := c.Components.Schemas.Get("Bar").Value
				require.Equal(t, bar, response.Value.Items.Value)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.c.Parse(&dynamic.Config{
				Info: dynamictest.NewConfigInfo(),
				Data: tc.c,
			}, &dynamictest.Reader{})
			tc.test(t, tc.c, err)
		})
	}
}

func Test_ParseAndPatch(t *testing.T) {
	testcases := []struct {
		name    string
		configs []*openapi.Config
		test    func(t *testing.T, c *openapi.Config)
	}{
		{
			name: "schemas reference",
			configs: []*openapi.Config{
				openapitest.NewConfig("3.1",
					openapitest.WithComponentSchema("Foo", schematest.New("array", schematest.WithItemsRef("#/components/schemas/Bar"))),
					openapitest.WithComponentSchema("Bar", schematest.New("string")),
				),
				openapitest.NewConfig("3.1",
					openapitest.WithComponentSchema("Bar", schematest.New("string", schematest.IsNullable(true))),
				),
			},
			test: func(t *testing.T, c *openapi.Config) {
				foo := c.Components.Schemas.Get("Foo").Value
				require.True(t, foo.Items.Value.Nullable)
			},
		},
		{
			name: "response schema reference should point to same objects",
			configs: []*openapi.Config{
				openapitest.NewConfig("3.1",
					openapitest.WithPath("/foo", openapitest.NewPath(
						openapitest.WithOperation("get", openapitest.NewOperation(
							openapitest.WithResponse(200, openapitest.WithContent("application/json",
								&openapi.MediaType{
									Schema: &schema.Ref{Reference: dynamic.Reference{Ref: "#/components/schemas/Foo"}},
								},
							)),
						)),
					)),
					openapitest.WithComponentSchema("Foo", schematest.New("array", schematest.WithItemsRef("#/components/schemas/Bar"))),
					openapitest.WithComponentSchema("Bar", schematest.New("string")),
				),
				openapitest.NewConfig("3.1",
					openapitest.WithComponentSchema("Bar", schematest.New("string", schematest.IsNullable(true))),
				),
			},
			test: func(t *testing.T, c *openapi.Config) {
				response := c.Paths["/foo"].Value.Get.Responses.GetResponse(200).Content["application/json"].Schema
				require.NotNil(t, response.Value)
				require.True(t, response.Value.Items.Value.Nullable)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var target *openapi.Config
			for _, c := range tc.configs {
				err := c.Parse(&dynamic.Config{
					Info: dynamictest.NewConfigInfo(),
					Data: c,
				}, &dynamictest.Reader{})
				require.NoError(t, err)
				if target == nil {
					target = c
				} else {
					target.Patch(c)
				}
			}

			tc.test(t, target)
		})
	}
}
