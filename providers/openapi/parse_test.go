package openapi_test

import (
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	"mokapi/providers/openapi/schema"
	"mokapi/providers/openapi/schema/schematest"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Parse(t *testing.T) {
	testcases := []struct {
		name string
		c    *openapi.Config
		test func(t *testing.T, c *openapi.Config, err error)
	}{
		{
			name: "response schema reference should point to same objects",
			c: openapitest.NewConfig("3.1",
				openapitest.WithPath("/foo",
					openapitest.WithOperation("get",
						openapitest.WithResponse(200, openapitest.UseContent("application/json",
							&openapi.MediaType{
								Schema: &schema.Schema{Ref: "#/components/schemas/Foo"},
							},
						),
						),
					)),
				openapitest.WithComponentSchema("Foo", schematest.New("array", schematest.WithItemsRef("#/components/schemas/Bar"))),
				openapitest.WithComponentSchema("Bar", schematest.New("string")),
			),
			test: func(t *testing.T, c *openapi.Config, err error) {
				require.NoError(t, err)
				response := c.Paths["/foo"].Value.Get.Responses.GetResponse(200).Content["application/json"].Schema
				require.NotNil(t, response)

				foo := c.Components.Schemas.Get("Foo")
				require.Equal(t, foo, response.Sub)
				bar := c.Components.Schemas.Get("Bar")
				require.Equal(t, bar, response.Items.Sub)
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
			name: "response schema reference should point to same objects",
			configs: []*openapi.Config{
				openapitest.NewConfig("3.1",
					openapitest.WithPath("/foo",
						openapitest.WithOperation("get",
							openapitest.WithResponse(200, openapitest.UseContent("application/json",
								&openapi.MediaType{
									Schema: &schema.Schema{Ref: "#/components/schemas/Foo"},
								},
							),
							),
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
				require.NotNil(t, response)
				require.True(t, response.Items.Nullable)
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
				if target == nil {
					target = c
				} else {
					target.Patch(c)
				}
			}

			err := target.Parse(&dynamic.Config{
				Info: dynamictest.NewConfigInfo(),
				Data: target,
			}, &dynamictest.Reader{})
			require.NoError(t, err)

			tc.test(t, target)
		})
	}
}
