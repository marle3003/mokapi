package openapi_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/media"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/schema/schematest"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFromRequest_QueryString(t *testing.T) {
	testcases := []struct {
		name    string
		params  openapi.Parameters
		request func() *http.Request
		test    func(t *testing.T, result *openapi.RequestParameters, err error)
	}{
		{
			name: "simple query string",
			params: openapi.Parameters{
				{Value: &openapi.Parameter{
					Type: openapi.ParameterQueryString,
					Content: map[string]*openapi.MediaType{
						"application/x-www-form-urlencoded": {
							Schema: schematest.New("object",
								schematest.WithProperty("foo", schematest.New("string")),
								schematest.WithProperty("bar", schematest.New("integer")),
							),
							ContentType: media.ParseContentType("application/x-www-form-urlencoded"),
						},
					},
				}}},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar?foo=hello%20world&bar=123", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]any{"bar": int64(123), "foo": "hello world"}, result.QueryString.Value)
				require.Equal(t, "foo=hello%20world&bar=123", *result.QueryString.Raw)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			r, err := openapi.FromRequest(tc.params, "", tc.request())
			tc.test(t, r, err)
		})
	}
}
