package parameter

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"mokapi/providers/openapi/schema"
	"net/http"
	"net/url"
	"testing"
)

func TestParseParam(t *testing.T) {
	testcases := []struct {
		request *http.Request
		route   string
		params  Parameters
		fn      func(RequestParameters) error
	}{
		{
			&http.Request{URL: &url.URL{Path: "/api/v1/channels/test/messages/123456743/test"}},
			"/api/v1/channels/{channel}/messages/{id}/test",
			Parameters{
				&Ref{Value: &Parameter{
					Name: "channel",
					Type: Path,
					Schema: &schema.Ref{
						Value: &schema.Schema{Type: "string"},
					},
					Required: true,
				}},
				&Ref{Value: &Parameter{
					Name: "id",
					Type: Path,
					Schema: &schema.Ref{
						Value: &schema.Schema{Type: "integer", Format: "int64"},
					},
					Required: true,
				}},
			},
			func(p RequestParameters) error {
				if channel, ok := p[Path]["channel"]; !ok {
					return errors.New("expected channel parameter in path")
				} else if channel.Value != "test" {
					return errors.Errorf("got %v; expected id value test", channel)
				}
				if id, ok := p[Path]["id"]; !ok {
					return errors.New("GET: expected id parameter in path")
				} else if id.Value != int64(123456743) {
					return errors.Errorf("got %v; expected id value 123456743", id)
				}
				return nil
			},
		},
		{
			&http.Request{URL: &url.URL{Path: "/api/v1/search", RawQuery: "limit=10"}},
			"/api/v1/search",
			Parameters{
				&Ref{
					Value: &Parameter{
						Name: "limit",
						Type: Query,
						Schema: &schema.Ref{
							Value: &schema.Schema{Type: "integer"},
						},
					},
				},
			},
			func(p RequestParameters) error {
				if limit, ok := p[Query]["limit"]; !ok {
					return errors.New("expected limit parameter in query")
				} else if limit.Value != int64(10) {
					t.Errorf("got %v; expected limit value 10", limit)
				}
				return nil
			},
		},
	}

	for _, testcase := range testcases {
		test := testcase
		t.Run(test.request.URL.String(), func(t *testing.T) {
			p, err := FromRequest(test.params, test.route, test.request)
			require.NoError(t, err)
			err = test.fn(p)
			require.NoError(t, err)
		})
	}
}
