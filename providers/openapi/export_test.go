package openapi_test

import (
	"encoding/json"
	"mokapi/config/dynamic"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	"mokapi/providers/openapi/schema"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExport(t *testing.T) {
	testcases := []struct {
		name string
		cfg  *openapi.Config
		test func(t *testing.T, jsonString, yamlString string, err error)
	}{
		{
			name: "path with global ref",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithPathRef("/foo",
					&openapi.PathRef{
						Reference: dynamic.Reference[*openapi.PathRef]{
							Ref: "/foo",
						},
						Value: &openapi.Path{Description: "desc"},
					},
				),
			),
			test: func(t *testing.T, jsonString, yamlString string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"openapi":"3.2.0","info":{"title":"","version":""},"paths":{"/foo":{"$ref":"#/components/pathItems/foo"}},"components":{"pathItems":{"foo":{"description":"desc"}}}}`, jsonString)
			},
		},
		{
			name: "path with local ref",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithPathRef("/foo",
					&openapi.PathRef{
						Reference: dynamic.Reference[*openapi.PathRef]{
							Ref: "#/components/pathItems/foo",
						},
						Value: &openapi.Path{Description: "desc"},
					},
				),
				openapitest.WithComponentPathItem("foo", &openapi.Path{Description: "desc"}),
			),
			test: func(t *testing.T, jsonString, yamlString string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"openapi":"3.2.0","info":{"title":"","version":""},"paths":{"/foo":{"$ref":"#/components/pathItems/foo"}},"components":{"pathItems":{"foo":{"description":"desc"}}}}`, jsonString)
			},
		},
		{
			name: "path with parameter local ref",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithPathRef("/foo",
					&openapi.PathRef{
						Value: &openapi.Path{
							Parameters: []*openapi.ParameterRef{
								{
									Reference: dynamic.Reference[*openapi.ParameterRef]{
										Ref: "#/components/parameters/foo",
									},
									Value: &openapi.Parameter{
										Name:   "foo",
										Type:   openapi.ParameterCookie,
										Schema: &schema.Schema{Description: "desc"},
									},
								},
							},
						},
					},
				),
				openapitest.WithComponentParameter("foo", &openapi.Parameter{Description: "desc"}),
			),
			test: func(t *testing.T, jsonString, yamlString string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"openapi":"3.2.0","info":{"title":"","version":""},"paths":{"/foo":{"parameters":[{"$ref":"#/components/parameters/foo"}]}},"components":{"parameters":{"foo":{"name":"foo","in":"cookie","schema":{"description":"desc"}}}}}`, jsonString)
			},
		},
		{
			name: "path with parameter local ref in operation",
			cfg: openapitest.NewConfig("3.2.0",
				openapitest.WithPathRef("/foo",
					&openapi.PathRef{
						Value: &openapi.Path{
							Get: &openapi.Operation{
								Parameters: []*openapi.ParameterRef{
									{
										Reference: dynamic.Reference[*openapi.ParameterRef]{
											Ref: "#/components/parameters/foo",
										},
										Value: &openapi.Parameter{Name: "foo", Type: openapi.ParameterCookie},
									},
								},
							},
						},
					},
				),
				openapitest.WithComponentParameter("foo", &openapi.Parameter{Description: "desc"}),
			),
			test: func(t *testing.T, jsonString, yamlString string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"openapi":"3.2.0","info":{"title":"","version":""},"paths":{"/foo":{"get":{"parameters":[{"$ref":"#/components/parameters/foo"}]}}},"components":{"parameters":{"foo":{"name":"foo","in":"cookie"}}}}`, jsonString)
			},
		},
		{
			name: "operation with request body global ref",
			cfg: func() *openapi.Config {
				rb := &openapi.RequestBody{
					Content: map[string]*openapi.MediaType{
						"application/json": {
							Schema: &schema.Schema{
								Reference: dynamic.Reference[*schema.Schema]{
									Ref: "#/components/schemas/foo",
								},
								Description: "desc",
							},
						},
					},
				}

				return openapitest.NewConfig("3.2.0",
					openapitest.WithPathRef("/foo",
						&openapi.PathRef{
							Value: &openapi.Path{
								Get: &openapi.Operation{
									RequestBody: &openapi.RequestBodyRef{
										Reference: dynamic.Reference[*openapi.RequestBodyRef]{
											Ref: "#/components/requestBodies/foo",
										},
										Value: rb,
									},
								},
							},
						},
					),
					openapitest.WithComponentRequestBody("foo", rb),
					openapitest.WithComponentSchema("foo", &schema.Schema{Description: "desc"}),
				)
			}(),
			test: func(t *testing.T, jsonString, yamlString string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"openapi":"3.2.0","info":{"title":"","version":""},"paths":{"/foo":{"get":{"requestBody":{"$ref":"#/components/requestBodies/foo"}}}},"components":{"schemas":{"foo":{"description":"desc"}},"requestBodies":{"foo":{"content":{"application/json":{"schema":{"$ref":"#/components/schemas/foo"}}}}}}}`, jsonString)
			},
		},
		{
			name: "operation with response local ref",
			cfg: func() *openapi.Config {
				resList := &openapi.Responses{}
				res := &openapi.Response{
					Content: map[string]*openapi.MediaType{
						"application/json": {
							Schema: &schema.Schema{Description: "desc"},
						},
					},
				}
				resList.Set("foo", &openapi.ResponseRef{
					Reference: dynamic.Reference[*openapi.ResponseRef]{
						Ref: "#/components/responses/foo",
					},
					Value: res,
				})

				return openapitest.NewConfig("3.2.0",
					openapitest.WithPathRef("/foo",
						&openapi.PathRef{
							Value: &openapi.Path{
								Get: &openapi.Operation{
									Responses: resList,
								},
							},
						},
					),
					openapitest.WithComponentResponse("foo", res),
				)
			}(),
			test: func(t *testing.T, jsonString, yamlString string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"openapi":"3.2.0","info":{"title":"","version":""},"paths":{"/foo":{"get":{"responses":{"foo":{"$ref":"#/components/responses/foo"}}}}},"components":{"responses":{"foo":{"content":{"application/json":{"schema":{"description":"desc"}}}}}}}`, jsonString)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			e := tc.cfg.Export()
			b, err := json.Marshal(e)
			tc.test(t, string(b), "", err)
		})
	}
}
