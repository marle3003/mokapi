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

func TestService_SendHttpRequest(t *testing.T) {
	testcases := []struct {
		name string
		app  *runtime.App
		test func(t *testing.T, s *mcp.Service)
	}{
		{
			name: "GET request path not specified",
			app: runtimetest.NewHttpApp(
				openapitest.NewConfig("3.1.0",
					openapitest.WithInfo("foo", "", ""),
				),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.SendHttpRequest(context.Background(), mcp.SendHttpRequestInput{APIName: "foo", Method: "GET", Path: "/foo"})
				require.NoError(t, err)
				require.Equal(t, http.StatusNotFound, r.Status)
				require.Equal(t, "path '/foo' not found", r.Body)
			},
		},
		{
			name: "GET request",
			app: runtimetest.NewHttpApp(
				openapitest.NewConfig("3.1.0",
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
				),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.SendHttpRequest(context.Background(), mcp.SendHttpRequestInput{APIName: "foo", Method: "GET", Path: "/foo"})
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, r.Status)
				require.Equal(t, `"Ln8rnaRqlL"`, r.Body)
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
