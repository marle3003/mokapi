package parameter_test

import (
	"context"
	"mokapi/providers/openapi/parameter"
	"mokapi/providers/openapi/schema/schematest"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParsePath(t *testing.T) {
	testcases := []struct {
		name    string
		param   *parameter.Parameter
		route   string
		request func() *http.Request
		test    func(t *testing.T, result parameter.RequestParameters, err error)
	}{
		{
			name: "simple path",
			param: &parameter.Parameter{
				Name:    "foo",
				Type:    parameter.Path,
				Schema:  schematest.New("string"),
				Style:   "",
				Explode: explode(false),
			},
			route: "/{foo}",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar/foo", nil)
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", result[parameter.Path]["foo"].Value)
			},
		},
		{
			name: "path parameter not present in route",
			param: &parameter.Parameter{
				Name:    "foo",
				Type:    parameter.Path,
				Schema:  schematest.New("string"),
				Style:   "",
				Explode: explode(false),
			},
			route: "/foo",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar/foo", nil)
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.EqualError(t, err, "parse path parameter 'foo' failed: path parameter foo not found in route /foo")
				require.Len(t, result[parameter.Path], 0)
			},
		},
		{
			name: "path parameter /v{version}",
			param: &parameter.Parameter{
				Name:    "version",
				Type:    parameter.Path,
				Schema:  schematest.New("string"),
				Style:   "",
				Explode: explode(false),
			},
			route: "/api/v{version}/foo",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar/api/v1/foo", nil)
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, "1", result[parameter.Path]["version"].Value)
			},
		},
		{
			name: "/report.{format}",
			param: &parameter.Parameter{
				Name:    "format",
				Type:    parameter.Path,
				Schema:  schematest.New("string"),
				Style:   "",
				Explode: explode(false),
			},
			route: "/report.{format}",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar/report.xml", nil)
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, "xml", result[parameter.Path]["format"].Value)
			},
		},
		{
			name: "labeled path",
			param: &parameter.Parameter{
				Name:    "foo",
				Type:    parameter.Path,
				Schema:  schematest.New("string"),
				Style:   "label",
				Explode: explode(false),
			},
			route: "/{foo}",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar/.foo", nil)
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", result[parameter.Path]["foo"].Value)
			},
		},
		{
			name: "matrix path",
			param: &parameter.Parameter{
				Name:    "foo",
				Type:    parameter.Path,
				Schema:  schematest.New("string"),
				Style:   "matrix",
				Explode: explode(false),
			},
			route: "/{foo}",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar/;foo", nil)
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", result[parameter.Path]["foo"].Value)
			},
		},
		{
			name: "array",
			param: &parameter.Parameter{

				Name:    "foo",
				Type:    parameter.Path,
				Schema:  schematest.New("array", schematest.WithItems("integer")),
				Style:   "",
				Explode: explode(false),
			},
			route: "/{foo}",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar/3,4,5", nil)
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int64(3), int64(4), int64(5)}, result[parameter.Path]["foo"].Value)
			},
		},
		{
			name: "labeled array",
			param: &parameter.Parameter{
				Name:    "foo",
				Type:    parameter.Path,
				Schema:  schematest.New("array", schematest.WithItems("integer")),
				Style:   "label",
				Explode: explode(false),
			},
			route: "/{foo}",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar/.3,4,5", nil)
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int64(3), int64(4), int64(5)}, result[parameter.Path]["foo"].Value)
			},
		},
		{
			name: "matrix array",
			param: &parameter.Parameter{

				Name:    "foo",
				Type:    parameter.Path,
				Schema:  schematest.New("array", schematest.WithItems("integer")),
				Style:   "matrix",
				Explode: explode(false),
			},
			route: "/{foo}",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar/;3,4,5", nil)
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int64(3), int64(4), int64(5)}, result[parameter.Path]["foo"].Value)
			},
		},
		{
			name: "object",
			param: &parameter.Parameter{
				Name: "foo",
				Type: parameter.Path,
				Schema: schematest.New("object",
					schematest.WithProperty("role", schematest.New("string")),
					schematest.WithProperty("firstName", schematest.New("string")),
				),
				Style:   "",
				Explode: explode(false),
			},
			route: "/{foo}",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar/role,admin,firstName,Alex", nil)
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"role": "admin", "firstName": "Alex"}, result[parameter.Path]["foo"].Value)
			},
		},
		{
			name: "object explode",
			param: &parameter.Parameter{
				Name: "foo",
				Type: parameter.Path,
				Schema: schematest.New("object",
					schematest.WithProperty("role", schematest.New("string")),
					schematest.WithProperty("firstName", schematest.New("string")),
					schematest.WithProperty("msg", schematest.New("string")),
					schematest.WithProperty("foo", schematest.New("string")),
				),
				Style:   "",
				Explode: explode(true),
			},
			route: "/{foo}",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar/role=admin,firstName=Alex,msg=Hello%20World,foo=foo%26bar", nil)
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"role": "admin", "firstName": "Alex", "msg": "Hello World", "foo": "foo&bar"}, result[parameter.Path]["foo"].Value)
			},
		},
		{
			name: "labeled object",
			param: &parameter.Parameter{
				Name: "foo",
				Type: parameter.Path,
				Schema: schematest.New("object",
					schematest.WithProperty("role", schematest.New("string")),
					schematest.WithProperty("firstName", schematest.New("string")),
				),
				Style:   "label",
				Explode: explode(false),
			},
			route: "/{foo}",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar/.role,admin,firstName,Alex", nil)
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"role": "admin", "firstName": "Alex"}, result[parameter.Path]["foo"].Value)
			},
		},
		{
			name: "matrix object",
			param: &parameter.Parameter{
				Name: "foo",
				Type: parameter.Path,
				Schema: schematest.New("object",
					schematest.WithProperty("role", schematest.New("string")),
					schematest.WithProperty("firstName", schematest.New("string")),
				),
				Style:   "matrix",
				Explode: explode(true),
			},
			route: "/{foo}",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar/;role=admin,firstName=Alex", nil)
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"role": "admin", "firstName": "Alex"}, result[parameter.Path]["foo"].Value)
			},
		},
		{
			name: "path parameter and base path",
			param: &parameter.Parameter{
				Name:    "foo",
				Type:    parameter.Path,
				Schema:  schematest.New("string"),
				Style:   "",
				Explode: explode(false),
			},
			route: "/{foo}",
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "https://foo.bar/mokapi/foo/bar", nil)
				return r.WithContext(context.WithValue(r.Context(), "servicePath", "/mokapi/foo"))
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, "bar", result[parameter.Path]["foo"].Value)
			},
		},
		{
			name: "path parameter and trailing slash in request",
			param: &parameter.Parameter{
				Name:    "foo",
				Type:    parameter.Path,
				Schema:  schematest.New("string"),
				Style:   "",
				Explode: explode(false),
			},
			route: "/{foo}",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar/bar/", nil)
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, "bar", result[parameter.Path]["foo"].Value)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			r, err := parameter.FromRequest(parameter.Parameters{{Value: tc.param}}, tc.route, tc.request())
			tc.test(t, r, err)
		})
	}
}
