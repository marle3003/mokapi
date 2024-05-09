package parameter

import (
	"github.com/stretchr/testify/require"
	jsonSchema "mokapi/json/schema"
	"mokapi/providers/openapi/schema"
	"net/http"
	"net/url"
	"testing"
)

func TestParseParam(t *testing.T) {
	testcases := []struct {
		name    string
		request *http.Request
		route   string
		params  Parameters
		test    func(t *testing.T, p RequestParameters)
	}{
		{
			name:    "path parameter",
			request: &http.Request{URL: &url.URL{Path: "/api/v1/channels/test/messages/123456743/test"}},
			route:   "/api/v1/channels/{channel}/messages/{id}/test",
			params: Parameters{
				&Ref{Value: &Parameter{
					Name: "channel",
					Type: Path,
					Schema: &schema.Ref{
						Value: &schema.Schema{Type: jsonSchema.Types{"string"}},
					},
					Required: true,
				}},
				&Ref{Value: &Parameter{
					Name: "id",
					Type: Path,
					Schema: &schema.Ref{
						Value: &schema.Schema{Type: jsonSchema.Types{"integer"}, Format: "int64"},
					},
					Required: true,
				}},
			},
			test: func(t *testing.T, p RequestParameters) {
				require.Contains(t, p[Path], "channel")
				require.Equal(t, "test", p[Path]["channel"].Value)
				require.Contains(t, p[Path], "id")
				require.Equal(t, int64(123456743), p[Path]["id"].Value)
			},
		},
		{
			name:    "query parameter",
			request: &http.Request{URL: &url.URL{Path: "/api/v1/search", RawQuery: "limit=10"}},
			route:   "/api/v1/search",
			params: Parameters{
				&Ref{
					Value: &Parameter{
						Name: "limit",
						Type: Query,
						Schema: &schema.Ref{
							Value: &schema.Schema{Type: jsonSchema.Types{"integer"}},
						},
					},
				},
			},
			test: func(t *testing.T, p RequestParameters) {
				require.Contains(t, p[Query], "limit")
				require.Equal(t, int64(10), p[Query]["limit"].Value)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p, err := FromRequest(tc.params, tc.route, tc.request)
			require.NoError(t, err)
			tc.test(t, p)
		})
	}
}
