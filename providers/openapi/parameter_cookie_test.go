package openapi_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/schema/schematest"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFromRequest_Cookie(t *testing.T) {
	testcases := []struct {
		name    string
		params  openapi.Parameters
		request func() *http.Request
		test    func(t *testing.T, result *openapi.RequestParameters, err error)
	}{
		{
			name: "simple cookie",
			params: openapi.Parameters{{Value: &openapi.Parameter{
				Type:   openapi.ParameterCookie,
				Name:   "debug",
				Schema: schematest.New("integer", schematest.WithEnumValues(0, 1)),
			}}},
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "https://foo.bar", nil)
				r.AddCookie(&http.Cookie{
					Name:  "debug",
					Value: "1",
				})
				return r
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error) {
				require.NoError(t, err)
				cookie := result.Cookie["debug"]
				require.Equal(t, int64(1), cookie.Value)
				require.Equal(t, "1", *cookie.Raw)
			},
		},
		{
			name: "cookie without schema",
			params: openapi.Parameters{{Value: &openapi.Parameter{
				Type: openapi.ParameterCookie,
				Name: "debug",
			}}},
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "https://foo.bar", nil)
				r.AddCookie(&http.Cookie{
					Name:  "debug",
					Value: "1",
				})
				return r
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error) {
				require.NoError(t, err)
				cookie := result.Cookie["debug"]
				require.Equal(t, "1", cookie.Value)
				require.Equal(t, "1", *cookie.Raw)
			},
		},
		{
			name: "not required cookie and not sent",
			params: openapi.Parameters{{Value: &openapi.Parameter{
				Type:     openapi.ParameterCookie,
				Name:     "debug",
				Required: false,
				Schema:   schematest.New("integer", schematest.WithEnumValues(0, 1)),
			}}},
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "https://foo.bar", nil)
				return r
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error) {
				require.NoError(t, err)
				require.Len(t, result.Cookie, 0)
			},
		},
		{
			name: "required cookie but not sent",
			params: openapi.Parameters{{Value: &openapi.Parameter{
				Type:     openapi.ParameterCookie,
				Name:     "debug",
				Required: true,
				Schema:   schematest.New("integer", schematest.WithEnumValues(0, 1)),
			}}},
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "https://foo.bar", nil)
				return r
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error) {
				require.EqualError(t, err, "parse cookie parameter 'debug' failed: parameter is required")
			},
		},
		{
			name: "required cookie but empty",
			params: openapi.Parameters{{Value: &openapi.Parameter{
				Type:     openapi.ParameterCookie,
				Name:     "debug",
				Required: true,
				Schema:   schematest.New("integer", schematest.WithEnumValues(0, 1)),
			}}},
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "https://foo.bar", nil)
				r.AddCookie(&http.Cookie{
					Name:  "debug",
					Value: "",
				})
				return r
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error) {
				require.EqualError(t, err, "parse cookie parameter 'debug' failed: parameter is required")
			},
		},
		{
			name: "invalid value",
			params: openapi.Parameters{{Value: &openapi.Parameter{
				Type:   openapi.ParameterCookie,
				Name:   "debug",
				Schema: schematest.New("integer", schematest.WithEnumValues(0, 1)),
			}}},
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "https://foo.bar", nil)
				r.AddCookie(&http.Cookie{
					Name:  "debug",
					Value: "foo",
				})
				return r
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error) {
				require.EqualError(t, err, "parse cookie parameter 'debug' failed: error count 1:\n\t- #/type: invalid type, expected integer but got string")
			},
		},
		{
			name: "array",
			params: openapi.Parameters{{Value: &openapi.Parameter{
				Type:   openapi.ParameterCookie,
				Name:   "foo",
				Schema: schematest.New("array", schematest.WithItems("integer")),
			}}},
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "https://foo.bar", nil)
				r.AddCookie(&http.Cookie{
					Name:  "foo",
					Value: "1,2,3",
				})
				return r
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error) {
				require.NoError(t, err)
				cookie := result.Cookie["foo"]
				require.Equal(t, []interface{}{int64(1), int64(2), int64(3)}, cookie.Value)
				require.Equal(t, "1,2,3", *cookie.Raw)
			},
		},
		{
			name: "array invalid value",
			params: openapi.Parameters{{Value: &openapi.Parameter{
				Type:   openapi.ParameterCookie,
				Name:   "foo",
				Schema: schematest.New("array", schematest.WithItems("integer")),
			}}},
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "https://foo.bar", nil)
				r.AddCookie(&http.Cookie{
					Name:  "foo",
					Value: "1,foo,3",
				})
				return r
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error) {
				require.EqualError(t, err, "parse cookie parameter 'foo' failed: error count 1:\n\t- #/items/1/type: invalid type, expected integer but got string")
			},
		},
		{
			name: "object",
			params: openapi.Parameters{{Value: &openapi.Parameter{
				Type: openapi.ParameterCookie,
				Name: "foo",
				Schema: schematest.New("object",
					schematest.WithProperty("role", schematest.New("string")),
					schematest.WithProperty("firstName", schematest.New("string")),
				),
			}}},
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "https://foo.bar", nil)
				r.AddCookie(&http.Cookie{
					Name:  "foo",
					Value: "role,admin,firstName,Alex",
				})
				return r
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error) {
				require.NoError(t, err)
				cookie := result.Cookie["foo"]
				require.Equal(t, map[string]interface{}{"firstName": "Alex", "role": "admin"}, cookie.Value)
				require.Equal(t, "role,admin,firstName,Alex", *cookie.Raw)
			},
		},
		{
			name: "object not all properties defined",
			params: openapi.Parameters{{Value: &openapi.Parameter{
				Type: openapi.ParameterCookie,
				Name: "foo",
				Schema: schematest.New("object",
					schematest.WithProperty("role", schematest.New("string")),
				),
			}}},
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "https://foo.bar", nil)
				r.AddCookie(&http.Cookie{
					Name:  "foo",
					Value: "role,admin,firstName,Alex",
				})
				return r
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error) {
				require.NoError(t, err)
				cookie := result.Cookie["foo"]
				require.Equal(t, map[string]interface{}{"role": "admin"}, cookie.Value)
				require.Equal(t, "role,admin,firstName,Alex", *cookie.Raw)
			},
		},
		{
			name: "object invalid property pairs",
			params: openapi.Parameters{{Value: &openapi.Parameter{
				Type: openapi.ParameterCookie,
				Name: "foo",
				Schema: schematest.New("object",
					schematest.WithProperty("role", schematest.New("string")),
					schematest.WithProperty("firstName", schematest.New("string")),
				),
			}}},
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "https://foo.bar", nil)
				r.AddCookie(&http.Cookie{
					Name:  "foo",
					Value: "role,admin,firstName",
				})
				return r
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error) {
				require.EqualError(t, err, "parse cookie parameter 'foo' failed: invalid number of property pairs")
			},
		},
		{
			name: "object invalid property",
			params: openapi.Parameters{{Value: &openapi.Parameter{
				Type: openapi.ParameterCookie,
				Name: "foo",
				Schema: schematest.New("object",
					schematest.WithProperty("role", schematest.New("string")),
					schematest.WithProperty("age", schematest.New("number")),
				),
			}}},
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "https://foo.bar", nil)
				r.AddCookie(&http.Cookie{
					Name:  "foo",
					Value: "role,admin,age,Alex",
				})
				return r
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error) {
				require.EqualError(t, err, "parse cookie parameter 'foo' failed: parse property 'age' failed: error count 1:\n\t- #/type: invalid type, expected number but got string")
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
