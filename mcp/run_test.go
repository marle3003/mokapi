package mcp_test

import (
	"context"
	"encoding/json"
	"mokapi/mcp"
	"mokapi/providers/openapi/openapitest"
	"mokapi/runtime"
	"mokapi/runtime/runtimetest"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestService_Run(t *testing.T) {
	testcases := []struct {
		name string
		app  *runtime.App
		test func(t *testing.T, s *mcp.Service)
	}{
		{
			name: "List APIs skip empty name",
			app: runtimetest.NewHttpApp(
				openapitest.NewConfig("3.1.0"),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `mokapi.getApis()`,
					},
				)
				require.NoError(t, err)
				require.Len(t, r.Result, 0)
			},
		},
		{
			name: "get HTTP APIs",
			app: runtimetest.NewHttpApp(
				openapitest.NewConfig("3.1.0",
					openapitest.WithInfo("foo", "", ""),
				),
				openapitest.NewConfig("3.1.0",
					openapitest.WithInfo("bar", "", ""),
				),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `mokapi.getApis()`,
					},
				)
				require.NoError(t, err)
				require.Equal(t, []mcp.ApiSummary{
					{Name: "bar", Type: "http"},
					{Name: "foo", Type: "http"},
				}, r.Result)
			},
		},
		{
			name: "get specific HTTP API",
			app: runtimetest.NewHttpApp(
				openapitest.NewConfig("3.1.0",
					openapitest.WithInfo("foo", "", ""),
				),
				openapitest.NewConfig("3.1.0",
					openapitest.WithInfo("bar", "", ""),
				),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `mokapi.getApi('foo')`,
					},
				)
				require.NoError(t, err)
				b, err := json.Marshal(r.Result)
				require.NoError(t, err)
				require.Equal(t, `{"name":"foo","type":"http"}`, string(b))
			},
		},
		{
			name: "get API's operation list",
			app: runtimetest.NewHttpApp(
				openapitest.NewConfig("3.1.0",
					openapitest.WithInfo("foo", "", ""),
					openapitest.WithPath("/pets",
						openapitest.WithOperation(http.MethodGet,
							openapitest.WithOperationSummary("GET summary"),
						),
						openapitest.WithOperation(http.MethodPut,
							openapitest.WithOperationSummary("PUT summary"),
						),
					),
					openapitest.WithPath("/users",
						openapitest.WithPathInfo("path summary", ""),
						openapitest.WithOperation(http.MethodPost),
					),
				),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `mokapi.getApi('foo').getOperations()`,
					},
				)
				require.NoError(t, err)
				require.Equal(t, []mcp.OperationSummary{
					{Method: "GET", Path: "/pets", Summary: "GET summary"},
					{Method: "PUT", Path: "/pets", Summary: "PUT summary"},
					{Method: "POST", Path: "/users", Summary: "path summary"}},
					r.Result)
			},
		},
		{
			name: "get API's operation details",
			app: runtimetest.NewHttpApp(
				openapitest.NewConfig("3.1.0",
					openapitest.WithInfo("foo", "", ""),
					openapitest.WithPath("/pets",
						openapitest.WithOperation(http.MethodGet,
							openapitest.WithOperationSummary("GET summary"),
						),
						openapitest.WithOperation(http.MethodPut,
							openapitest.WithOperationSummary("PUT summary"),
						),
					),
				),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `mokapi.getApi('foo').getOperationDetails('/pets', 'GET')`,
					},
				)
				require.NoError(t, err)
				require.Equal(t, &mcp.OperationDetails{
					OperationId: "",
					Method:      "GET",
					Path:        "/pets",
					Summary:     "GET summary",
					Description: ""},
					r.Result)
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
