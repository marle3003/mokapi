package mcp_test

import (
	"context"
	"mokapi/mcp"
	"mokapi/providers/openapi/openapitest"
	"mokapi/runtime"
	"mokapi/runtime/runtimetest"
	"mokapi/schema/json/generator"
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
			name: "run JavaScript code",
			app:  runtimetest.NewApp(),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `1+1`,
					},
				)
				require.NoError(t, err)
				require.Equal(t, int64(2), r.Result)
			},
		},
		{
			name: "JSON.parse()",
			app:  runtimetest.NewApp(),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `JSON.parse('{"foo":"bar"}')`,
					},
				)
				require.NoError(t, err)
				require.Equal(t, map[string]any{"foo": "bar"}, r.Result)
			},
		},
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
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			generator.Seed(123456)

			s := mcp.NewService(tc.app)
			tc.test(t, s)
		})
	}
}
