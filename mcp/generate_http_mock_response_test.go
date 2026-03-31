package mcp_test

import (
	"context"
	"mokapi/mcp"
	"mokapi/providers/openapi/openapitest"
	"mokapi/providers/openapi/schema/schematest"
	"mokapi/runtime"
	"mokapi/runtime/runtimetest"
	"mokapi/schema/json/generator"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestService_GenerateHttpResponse(t *testing.T) {
	testcases := []struct {
		name string
		app  *runtime.App
		test func(t *testing.T, s *mcp.Service)
	}{
		{
			name: "Generate response result",
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
				result, err := s.GenerateHttpMockResponse(context.Background(), mcp.GenerateHttpMockResponseInput{
					ApiName:    "foo",
					Path:       "/foo",
					Method:     http.MethodGet,
					StatusCode: http.StatusOK,
				})
				require.NoError(t, err)
				require.Equal(t, mcp.GenerateHttpMockResponseOutput{
					StatusCode: http.StatusOK,
					Data:       "Ln8rnaRqlL",
					Headers:    map[string]any{},
				}, result)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			generator.Seed(12345)

			s := mcp.NewService(tc.app)
			tc.test(t, s)
		})
	}
}
