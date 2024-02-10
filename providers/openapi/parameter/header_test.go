package parameter_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/providers/openapi/parameter"
	"mokapi/providers/openapi/schema"
	"mokapi/providers/openapi/schema/schematest"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFromRequest_Header(t *testing.T) {
	testcases := []struct {
		name    string
		params  parameter.Parameters
		request func() *http.Request
		test    func(t *testing.T, result parameter.RequestParameters, err error)
	}{
		{
			name: "simple header",
			params: parameter.Parameters{{Value: &parameter.Parameter{
				Type:   parameter.Header,
				Name:   "debug",
				Schema: &schema.Ref{Value: &schema.Schema{Type: "integer", Enum: []interface{}{0, 1}}},
			}}},
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "https://foo.bar", nil)
				r.Header.Set("debug", "1")
				return r
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.NoError(t, err)
				cookie := result[parameter.Header]["debug"]
				require.Equal(t, int64(1), cookie.Value)
				require.Equal(t, "1", cookie.Raw)
			},
		},
		{
			name: "header without schema",
			params: parameter.Parameters{{Value: &parameter.Parameter{
				Type: parameter.Header,
				Name: "debug",
			}}},
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "https://foo.bar", nil)
				r.Header.Set("debug", "1")
				return r
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.NoError(t, err)
				cookie := result[parameter.Header]["debug"]
				require.Equal(t, "1", cookie.Value)
				require.Equal(t, "1", cookie.Raw)
			},
		},
		{
			name: "not required header and not sent",
			params: parameter.Parameters{{Value: &parameter.Parameter{
				Type:     parameter.Header,
				Name:     "debug",
				Required: false,
				Schema:   &schema.Ref{Value: &schema.Schema{Type: "integer", Enum: []interface{}{0, 1}}},
			}}},
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "https://foo.bar", nil)
				return r
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.NoError(t, err)
				require.Len(t, result[parameter.Header], 0)
			},
		},
		{
			name: "required header but not sent",
			params: parameter.Parameters{{Value: &parameter.Parameter{
				Type:     parameter.Header,
				Name:     "debug",
				Required: true,
				Schema:   &schema.Ref{Value: &schema.Schema{Type: "integer", Enum: []interface{}{0, 1}}},
			}}},
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "https://foo.bar", nil)
				return r
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.EqualError(t, err, "parse header parameter 'debug' failed: parameter is required")
				require.Len(t, result[parameter.Header], 0)
			},
		},
		{
			name: "required header but empty",
			params: parameter.Parameters{{Value: &parameter.Parameter{
				Type:     parameter.Header,
				Name:     "debug",
				Required: true,
				Schema:   &schema.Ref{Value: &schema.Schema{Type: "integer", Enum: []interface{}{0, 1}}},
			}}},
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "https://foo.bar", nil)
				r.Header.Set("debug", "")
				return r
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.EqualError(t, err, "parse header parameter 'debug' failed: parameter is required")
				require.Len(t, result[parameter.Header], 0)
			},
		},
		{
			name: "invalid value",
			params: parameter.Parameters{{Value: &parameter.Parameter{
				Type:   parameter.Header,
				Name:   "debug",
				Schema: &schema.Ref{Value: &schema.Schema{Type: "integer", Enum: []interface{}{0, 1}}},
			}}},
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "https://foo.bar", nil)
				r.Header.Set("debug", "foo")
				return r
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.EqualError(t, err, "parse header parameter 'debug' failed: parse 'foo' failed, expected schema type=integer")
				require.Len(t, result[parameter.Header], 0)
			},
		},
		{
			name: "array",
			params: parameter.Parameters{{Value: &parameter.Parameter{
				Type: parameter.Header,
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
				r.Header.Set("foo", "1,2,3")
				return r
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.NoError(t, err)
				cookie := result[parameter.Header]["foo"]
				require.Equal(t, []interface{}{int64(1), int64(2), int64(3)}, cookie.Value)
				require.Equal(t, "1,2,3", cookie.Raw)
			},
		},
		{
			name: "array invalid value",
			params: parameter.Parameters{{Value: &parameter.Parameter{
				Type: parameter.Header,
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
				r.Header.Set("foo", "1,foo,3")
				return r
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.EqualError(t, err, "parse header parameter 'foo' failed: parse 'foo' failed, expected schema type=integer")
				require.Len(t, result[parameter.Header], 0)
			},
		},
		{
			name: "object",
			params: parameter.Parameters{{Value: &parameter.Parameter{
				Type: parameter.Header,
				Name: "foo",
				Schema: &schema.Ref{
					Value: schematest.New("object",
						schematest.WithProperty("role", schematest.New("string")),
						schematest.WithProperty("firstName", schematest.New("string")),
					)},
			}}},
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "https://foo.bar", nil)
				r.Header.Set("foo", "role,admin,firstName,Alex")
				return r
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.NoError(t, err)
				cookie := result[parameter.Header]["foo"]
				require.Equal(t, map[string]interface{}{"firstName": "Alex", "role": "admin"}, cookie.Value)
				require.Equal(t, "role,admin,firstName,Alex", cookie.Raw)
			},
		},
		{
			name: "object not all properties defined",
			params: parameter.Parameters{{Value: &parameter.Parameter{
				Type: parameter.Header,
				Name: "foo",
				Schema: &schema.Ref{
					Value: schematest.New("object",
						schematest.WithProperty("role", schematest.New("string")),
					)},
			}}},
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "https://foo.bar", nil)
				r.Header.Set("foo", "role,admin,firstName,Alex")
				return r
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.NoError(t, err)
				cookie := result[parameter.Header]["foo"]
				require.Equal(t, map[string]interface{}{"role": "admin"}, cookie.Value)
				require.Equal(t, "role,admin,firstName,Alex", cookie.Raw)
			},
		},
		{
			name: "object invalid property pairs",
			params: parameter.Parameters{{Value: &parameter.Parameter{
				Type: parameter.Header,
				Name: "foo",
				Schema: &schema.Ref{
					Value: schematest.New("object",
						schematest.WithProperty("role", schematest.New("string")),
						schematest.WithProperty("firstName", schematest.New("string")),
					)},
			}}},
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "https://foo.bar", nil)
				r.Header.Set("foo", "role,admin,firstName")
				return r
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.EqualError(t, err, "parse header parameter 'foo' failed: invalid number of property pairs")
				require.Len(t, result[parameter.Cookie], 0)
			},
		},
		{
			name: "object invalid property",
			params: parameter.Parameters{{Value: &parameter.Parameter{
				Type: parameter.Header,
				Name: "foo",
				Schema: &schema.Ref{
					Value: schematest.New("object",
						schematest.WithProperty("role", schematest.New("string")),
						schematest.WithProperty("age", schematest.New("number")),
					)},
			}}},
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "https://foo.bar", nil)
				r.Header.Set("foo", "role,admin,age,Alex")
				return r
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.EqualError(t, err, "parse header parameter 'foo' failed: parse property 'age' failed: parse 'Alex' failed, expected schema type=number")
				require.Len(t, result[parameter.Header], 0)
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
