package mcp_test

import (
	"context"
	"mokapi/mcp"
	"mokapi/providers/openapi/openapitest"
	"mokapi/providers/openapi/schema"
	"mokapi/providers/openapi/schema/schematest"
	"mokapi/runtime"
	"mokapi/runtime/runtimetest"
	jsonSchema "mokapi/schema/json/schema"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestService_GetHttpResponseSchema(t *testing.T) {
	testcases := []struct {
		name string
		app  *runtime.App
		test func(t *testing.T, s *mcp.Service)
	}{
		{
			name: "Get Response Schema",
			app: runtimetest.NewApp(
				runtimetest.WithHttp(openapitest.NewConfig("3.1.0",
					openapitest.WithInfo("foo", "", ""),
					openapitest.WithPath("/foo",
						openapitest.WithOperation(http.MethodGet,
							openapitest.WithResponse(http.StatusOK,
								openapitest.WithContent("application/json",
									openapitest.WithSchema(schematest.New("string")),
								),
							),
						),
					),
				)),
			),
			test: func(t *testing.T, s *mcp.Service) {
				result, err := s.GetHttpResponseSchema(context.Background(), mcp.GetHttpResponseSchemaInput{
					ApiName:    "foo",
					Path:       "/foo",
					Method:     http.MethodGet,
					StatusCode: http.StatusOK,
				})
				require.NoError(t, err)
				require.IsType(t, &schema.Schema{}, result)
				require.Equal(t, jsonSchema.Types{"string"}, result.(*schema.Schema).Type)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s := mcp.NewService(tc.app)
			tc.test(t, s)
		})
	}
}
