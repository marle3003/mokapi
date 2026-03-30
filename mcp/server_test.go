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

	list, err := session.ListTools(ctx, &gomcp.ListToolsParams{})
	require.NoError(t, err)
	require.Len(t, list.Tools, 6)
	// alphabetical order
	require.Equal(t, "get_api_list", list.Tools[0].Name)
	require.Equal(t, "get_api_spec", list.Tools[1].Name)
	require.Equal(t, "get_events", list.Tools[2].Name)
	require.Equal(t, "get_mokapi_js_api", list.Tools[3].Name)
	require.Equal(t, "produce_kafka_message", list.Tools[4].Name)
	require.Equal(t, "send_http_request", list.Tools[5].Name)
}
