package openapi

import (
	"github.com/stretchr/testify/require"
	"mokapi/providers/openapi/schema/schematest"
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
		test    func(t *testing.T, p *RequestParameters)
	}{
		{
			name:    "path parameter",
			request: &http.Request{URL: &url.URL{Path: "/api/v1/channels/test/messages/123456743/test"}},
			route:   "/api/v1/channels/{channel}/messages/{id}/test",
			params: Parameters{
				&ParameterRef{Value: &Parameter{
					Name:     "channel",
					Type:     ParameterPath,
					Schema:   schematest.New("string"),
					Required: true,
				}},
				&ParameterRef{Value: &Parameter{
					Name:     "id",
					Type:     ParameterPath,
					Schema:   schematest.New("integer", schematest.WithFormat("int64")),
					Required: true,
				}},
			},
			test: func(t *testing.T, p *RequestParameters) {
				require.Contains(t, p.Path, "channel")
				require.Equal(t, "test", p.Path["channel"].Value)
				require.Contains(t, p.Path, "id")
				require.Equal(t, int64(123456743), p.Path["id"].Value)
			},
		},
		{
			name:    "query parameter",
			request: &http.Request{URL: &url.URL{Path: "/api/v1/search", RawQuery: "limit=10"}},
			route:   "/api/v1/search",
			params: Parameters{
				&ParameterRef{
					Value: &Parameter{
						Name:   "limit",
						Type:   ParameterQuery,
						Schema: schematest.New("integer"),
					},
				},
			},
			test: func(t *testing.T, p *RequestParameters) {
				require.Contains(t, p.Query, "limit")
				require.Equal(t, int64(10), p.Query["limit"].Value)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			params, err := FromRequest(tc.params, tc.route, tc.request)
			require.NoError(t, err)
			tc.test(t, params)
		})
	}
}
