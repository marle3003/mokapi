package parameter_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/openapi/parameter"
	"mokapi/config/dynamic/openapi/schema"
	"mokapi/config/dynamic/openapi/schema/schematest"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFromRequest_Cookie(t *testing.T) {
	testcases := []struct {
		name    string
		params  parameter.Parameters
		request func() *http.Request
		test    func(t *testing.T, result parameter.RequestParameters, err error)
	}{
		{
			name: "simple cookie",
			params: parameter.Parameters{{Value: &parameter.Parameter{
				Type:   parameter.Cookie,
				Name:   "debug",
				Schema: &schema.Ref{Value: &schema.Schema{Type: "integer", Enum: []interface{}{0, 1}}},
			}}},
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "https://foo.bar", nil)
				r.AddCookie(&http.Cookie{
					Name:  "debug",
					Value: "1",
				})
				return r
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.NoError(t, err)
				cookie := result[parameter.Cookie]["debug"]
				require.Equal(t, int64(1), cookie.Value)
			},
		},
		{
			name: "required cookie but not sent",
			params: parameter.Parameters{{Value: &parameter.Parameter{
				Type:     parameter.Cookie,
				Name:     "debug",
				Required: true,
				Schema:   &schema.Ref{Value: &schema.Schema{Type: "integer", Enum: []interface{}{0, 1}}},
			}}},
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "https://foo.bar", nil)
				return r
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.EqualError(t, err, "cookie parameter 'debug' is required")
			},
		},
		{
			name: "required cookie but empty",
			params: parameter.Parameters{{Value: &parameter.Parameter{
				Type:     parameter.Cookie,
				Name:     "debug",
				Required: true,
				Schema:   &schema.Ref{Value: &schema.Schema{Type: "integer", Enum: []interface{}{0, 1}}},
			}}},
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "https://foo.bar", nil)
				r.AddCookie(&http.Cookie{
					Name:  "debug",
					Value: "",
				})
				return r
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.EqualError(t, err, "cookie parameter 'debug' is required")
			},
		},
		{
			name: "invalid value",
			params: parameter.Parameters{{Value: &parameter.Parameter{
				Type:   parameter.Cookie,
				Name:   "debug",
				Schema: &schema.Ref{Value: &schema.Schema{Type: "integer", Enum: []interface{}{0, 1}}},
			}}},
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "https://foo.bar", nil)
				r.AddCookie(&http.Cookie{
					Name:  "debug",
					Value: "foo",
				})
				return r
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.EqualError(t, err, "parse cookie 'debug' failed: could not parse 'foo' as int, expected schema type=integer")
			},
		},
		{
			name: "array",
			params: parameter.Parameters{{Value: &parameter.Parameter{
				Type: parameter.Cookie,
				Name: "foo",
				Schema: &schema.Ref{
					Value: &schema.Schema{
						Type: "array",
						Items: &schema.Ref{Value: &schema.Schema{
							Type: "integer"}},
					}},
			}}},
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "https://foo.bar", nil)
				r.AddCookie(&http.Cookie{
					Name:  "foo",
					Value: "1,2,3",
				})
				return r
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.NoError(t, err)
				cookie := result[parameter.Cookie]["foo"]
				require.Equal(t, []interface{}{int64(1), int64(2), int64(3)}, cookie.Value)
			},
		}, {
			name: "object",
			params: parameter.Parameters{{Value: &parameter.Parameter{
				Type: parameter.Cookie,
				Name: "foo",
				Schema: &schema.Ref{
					Value: schematest.New("object",
						schematest.WithProperty("role", schematest.New("string")),
						schematest.WithProperty("firstName", schematest.New("string")),
					)},
			}}},
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "https://foo.bar", nil)
				r.AddCookie(&http.Cookie{
					Name:  "foo",
					Value: "role,admin,firstName,Alex",
				})
				return r
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.NoError(t, err)
				cookie := result[parameter.Cookie]["foo"]
				require.Equal(t, map[string]interface{}{"firstName": "Alex", "role": "admin"}, cookie.Value)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			r, err := parameter.FromRequest(tc.params, "", tc.request())
			tc.test(t, r, err)
		})
	}
}
