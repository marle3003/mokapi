package mcp_test

import (
	"context"
	"mokapi/mcp"
	"mokapi/runtime/runtimetest"
	"net/http/httptest"
	"testing"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	ctx := context.Background()
	defer ctx.Done()

	h := mcp.NewServer(runtimetest.NewApp())
	s := httptest.NewServer(h)
	defer s.Close()

	client := gomcp.NewClient(&gomcp.Implementation{Name: "mcp-client", Version: "v1.0.0"}, nil)

	transport := &gomcp.StreamableClientTransport{
		Endpoint: s.URL,
	}
	session, err := client.Connect(ctx, transport, nil)
	require.NoError(t, err)
	require.NotNil(t, session)
	defer func() { _ = session.Close() }()

	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "test tool list",
			test: func(t *testing.T) {
				list, err := session.ListTools(ctx, &gomcp.ListToolsParams{})
				require.NoError(t, err)
				require.Len(t, list.Tools, 9)
				// alphabetical order
				require.Equal(t, "mokapi_execute_code", list.Tools[0].Name)
				require.Equal(t, "mokapi_generate_http_mock_response", list.Tools[1].Name)
				require.Equal(t, "mokapi_get_api_spec", list.Tools[2].Name)
				require.Equal(t, "mokapi_get_events", list.Tools[3].Name)
				require.Equal(t, "mokapi_get_http_mock_template", list.Tools[4].Name)
				require.Equal(t, "mokapi_get_scenarios", list.Tools[5].Name)
				require.Equal(t, "mokapi_get_typescript_api", list.Tools[6].Name)
				require.Equal(t, "mokapi_produce_kafka_message", list.Tools[7].Name)
				require.Equal(t, "mokapi_send_http_request", list.Tools[8].Name)
			},
		},
		{
			name: "run code",
			test: func(t *testing.T) {
				r, err := session.CallTool(ctx, &gomcp.CallToolParams{
					Name: "mokapi_execute_code",
					Arguments: mcp.RunInput{
						Code: "1+1",
					},
				})
				require.NoError(t, err)
				require.IsType(t, &gomcp.TextContent{}, r.Content[0])
				tc := r.Content[0].(*gomcp.TextContent)
				require.Equal(t, `{"result":2}`, tc.Text)
			},
		},
		{
			name: "get mokapi_execute_code tool information",
			test: func(t *testing.T) {
				list, err := session.ListTools(ctx, &gomcp.ListToolsParams{
					Meta:   nil,
					Cursor: "",
				})
				require.NoError(t, err)
				require.Contains(t, list.Tools[0].Description, "api://execute-types")
			},
		},
		{
			name: "get resource execute-types",
			test: func(t *testing.T) {
				list, err := session.ListResources(ctx, &gomcp.ListResourcesParams{})
				require.NoError(t, err)
				require.Equal(t, "api-docs", list.Resources[0].Name)
				require.Equal(t, "api://execute-types", list.Resources[0].URI)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t)
		})
	}

}
