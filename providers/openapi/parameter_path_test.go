package openapi_test

import (
	"context"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/schema/schematest"
	"net/http"
	"net/http/httptest"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestParsePath(t *testing.T) {
	testcases := []struct {
		name    string
		param   []*openapi.Parameter
		route   string
		request func() *http.Request
		test    func(t *testing.T, result *openapi.RequestParameters, err error, hook *test.Hook)
	}{
		{
			name: "simple path",
			param: []*openapi.Parameter{
				{
					Name:    "foo",
					Type:    openapi.ParameterPath,
					Schema:  schematest.New("string"),
					Style:   "",
					Explode: new(false),
				},
			},
			route: "/{foo}",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar/foo", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error, _ *test.Hook) {
				require.NoError(t, err)
				require.Equal(t, "foo", result.Path["foo"].Value)
			},
		},
		{
			name: "path parameter not present in route",
			param: []*openapi.Parameter{
				{
					Name:    "foo",
					Type:    openapi.ParameterPath,
					Schema:  schematest.New("string"),
					Style:   "",
					Explode: new(false),
				},
			},
			route: "/foo",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar/foo", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error, _ *test.Hook) {
				require.EqualError(t, err, "parse path parameter 'foo' failed: path parameter foo not found in route /foo")
			},
		},
		{
			name: "path parameter /v{version}",
			param: []*openapi.Parameter{
				{
					Name:    "version",
					Type:    openapi.ParameterPath,
					Schema:  schematest.New("string"),
					Style:   "",
					Explode: new(false),
				},
			},
			route: "/api/v{version}/foo",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar/api/v1/foo", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error, _ *test.Hook) {
				require.NoError(t, err)
				require.Equal(t, "1", result.Path["version"].Value)
			},
		},
		{
			name: "/report.{format}",
			param: []*openapi.Parameter{
				{
					Name:    "format",
					Type:    openapi.ParameterPath,
					Schema:  schematest.New("string"),
					Style:   "",
					Explode: new(false),
				},
			},
			route: "/report.{format}",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar/report.xml", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error, _ *test.Hook) {
				require.NoError(t, err)
				require.Equal(t, "xml", result.Path["format"].Value)
			},
		},
		{
			name: "labeled path",
			param: []*openapi.Parameter{
				{
					Name:    "foo",
					Type:    openapi.ParameterPath,
					Schema:  schematest.New("string"),
					Style:   "label",
					Explode: new(false),
				},
			},
			route: "/{foo}",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar/.foo", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error, _ *test.Hook) {
				require.NoError(t, err)
				require.Equal(t, "foo", result.Path["foo"].Value)
			},
		},
		{
			name: "matrix path two parameters in same segment",
			param: []*openapi.Parameter{
				{
					Name:    "color",
					Type:    openapi.ParameterPath,
					Schema:  schematest.New("string"),
					Style:   "matrix",
					Explode: new(false),
				},
				{
					Name:    "size",
					Type:    openapi.ParameterPath,
					Schema:  schematest.New("string"),
					Style:   "matrix",
					Explode: new(false),
				},
			},
			route: "/{color}{size}",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar/;color=blue;size=M", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error, _ *test.Hook) {
				require.NoError(t, err)
				require.Equal(t, "blue", result.Path["color"].Value)
				require.Equal(t, "M", result.Path["size"].Value)
			},
		},
		{
			name: "label path two parameters in same segment",
			param: []*openapi.Parameter{
				{
					Name:    "color",
					Type:    openapi.ParameterPath,
					Schema:  schematest.New("string"),
					Style:   "label",
					Explode: new(false),
				},
				{
					Name:    "size",
					Type:    openapi.ParameterPath,
					Schema:  schematest.New("string"),
					Style:   "label",
					Explode: new(false),
				},
			},
			route: "/{color}{size}",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar/.blue.M", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error, _ *test.Hook) {
				require.NoError(t, err)
				require.Equal(t, "blue", result.Path["color"].Value)
				require.Equal(t, "M", result.Path["size"].Value)
			},
		},
		{
			name: "array",
			param: []*openapi.Parameter{
				{
					Name:    "foo",
					Type:    openapi.ParameterPath,
					Schema:  schematest.New("array", schematest.WithItems("integer")),
					Style:   "",
					Explode: new(false),
				},
			},
			route: "/{foo}",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar/3,4,5", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error, _ *test.Hook) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int64(3), int64(4), int64(5)}, result.Path["foo"].Value)
			},
		},
		{
			name: "labeled array",
			param: []*openapi.Parameter{
				{
					Name:    "foo",
					Type:    openapi.ParameterPath,
					Schema:  schematest.New("array", schematest.WithItems("integer")),
					Style:   "label",
					Explode: new(false),
				},
			},
			route: "/{foo}",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar/.3,4,5", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error, _ *test.Hook) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int64(3), int64(4), int64(5)}, result.Path["foo"].Value)
			},
		},
		{
			name: "labeled array exploded",
			param: []*openapi.Parameter{
				{
					Name:    "foo",
					Type:    openapi.ParameterPath,
					Schema:  schematest.New("array", schematest.WithItems("integer")),
					Style:   "label",
					Explode: new(true),
				},
			},
			route: "/{foo}",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar/.3.4.5", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error, _ *test.Hook) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int64(3), int64(4), int64(5)}, result.Path["foo"].Value)
			},
		},
		{
			name: "matrix array",
			param: []*openapi.Parameter{
				{
					Name:    "color",
					Type:    openapi.ParameterPath,
					Schema:  schematest.New("array", schematest.WithItems("string")),
					Style:   "matrix",
					Explode: new(false),
				},
			},
			route: "/{color}",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar/;color=blue,black,brown", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error, _ *test.Hook) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"blue", "black", "brown"}, result.Path["color"].Value)
			},
		},
		{
			name: "matrix array exploded",
			param: []*openapi.Parameter{
				{
					Name:    "color",
					Type:    openapi.ParameterPath,
					Schema:  schematest.New("array", schematest.WithItems("string")),
					Style:   "matrix",
					Explode: new(true),
				},
			},
			route: "/{color}",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar/;color=blue;color=black;color=brown", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error, _ *test.Hook) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"blue", "black", "brown"}, result.Path["color"].Value)
			},
		},
		{
			name: "object",
			param: []*openapi.Parameter{
				{
					Name: "foo",
					Type: openapi.ParameterPath,
					Schema: schematest.New("object",
						schematest.WithProperty("role", schematest.New("string")),
						schematest.WithProperty("firstName", schematest.New("string")),
					),
					Style:   "",
					Explode: new(false),
				},
			},
			route: "/{foo}",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar/role,admin,firstName,Alex", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error, _ *test.Hook) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"role": "admin", "firstName": "Alex"}, result.Path["foo"].Value)
			},
		},
		{
			name: "object explode",
			param: []*openapi.Parameter{
				{
					Name: "foo",
					Type: openapi.ParameterPath,
					Schema: schematest.New("object",
						schematest.WithProperty("role", schematest.New("string")),
						schematest.WithProperty("firstName", schematest.New("string")),
						schematest.WithProperty("msg", schematest.New("string")),
						schematest.WithProperty("foo", schematest.New("string")),
					),
					Style:   "",
					Explode: new(true),
				},
			},
			route: "/{foo}",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar/role=admin,firstName=Alex,msg=Hello%20World,foo=foo%26bar", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error, _ *test.Hook) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"role": "admin", "firstName": "Alex", "msg": "Hello World", "foo": "foo&bar"}, result.Path["foo"].Value)
			},
		},
		{
			name: "labeled object",
			param: []*openapi.Parameter{
				{
					Name: "foo",
					Type: openapi.ParameterPath,
					Schema: schematest.New("object",
						schematest.WithProperty("role", schematest.New("string")),
						schematest.WithProperty("firstName", schematest.New("string")),
					),
					Style:   "label",
					Explode: new(false),
				},
			},
			route: "/{foo}",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar/.role,admin,firstName,Alex", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error, _ *test.Hook) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"role": "admin", "firstName": "Alex"}, result.Path["foo"].Value)
			},
		},
		{
			name: "labeled object exploded",
			param: []*openapi.Parameter{
				{
					Name: "foo",
					Type: openapi.ParameterPath,
					Schema: schematest.New("object",
						schematest.WithProperty("role", schematest.New("string")),
						schematest.WithProperty("firstName", schematest.New("string")),
					),
					Style:   "label",
					Explode: new(true),
				},
			},
			route: "/{foo}",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar/.role=admin.firstName=Alex", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error, _ *test.Hook) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"role": "admin", "firstName": "Alex"}, result.Path["foo"].Value)
			},
		},
		{
			name: "matrix object",
			param: []*openapi.Parameter{
				{
					Name: "color",
					Type: openapi.ParameterPath,
					Schema: schematest.New("object",
						schematest.WithProperty("R", schematest.New("string")),
						schematest.WithProperty("G", schematest.New("string")),
						schematest.WithProperty("B", schematest.New("string")),
					),
					Style:   "matrix",
					Explode: new(false),
				},
			},
			route: "/{color}",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar/;color=R,100,G,200,B,150", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error, _ *test.Hook) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"B": "150", "G": "200", "R": "100"}, result.Path["color"].Value)
			},
		},
		{
			name: "matrix object with additional characters in segment",
			param: []*openapi.Parameter{
				{
					Name: "color",
					Type: openapi.ParameterPath,
					Schema: schematest.New("object",
						schematest.WithProperty("R", schematest.New("string")),
						schematest.WithProperty("G", schematest.New("string")),
						schematest.WithProperty("B", schematest.New("string")),
					),
					Style:   "matrix",
					Explode: new(false),
				},
			},
			route: "/foo{color}bar",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar/foo;color=R,100,G,200,B,150;bar", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error, _ *test.Hook) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"B": "150", "G": "200", "R": "100"}, result.Path["color"].Value)
			},
		},
		{
			name: "matrix object exploded",
			param: []*openapi.Parameter{
				{
					Name: "color",
					Type: openapi.ParameterPath,
					Schema: schematest.New("object",
						schematest.WithProperty("R", schematest.New("string")),
						schematest.WithProperty("G", schematest.New("string")),
						schematest.WithProperty("B", schematest.New("string")),
					),
					Style:   "matrix",
					Explode: new(true),
				},
			},
			route: "/{color}",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar/;R=100;G=200;B=150", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error, _ *test.Hook) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"B": "150", "G": "200", "R": "100"}, result.Path["color"].Value)
			},
		},
		{
			name: "path parameter and base path",
			param: []*openapi.Parameter{
				{
					Name:    "foo",
					Type:    openapi.ParameterPath,
					Schema:  schematest.New("string"),
					Style:   "",
					Explode: new(false),
				},
			},
			route: "/{foo}",
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "https://foo.bar/mokapi/foo/bar", nil)
				return r.WithContext(context.WithValue(r.Context(), "servicePath", "/mokapi/foo"))
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error, _ *test.Hook) {
				require.NoError(t, err)
				require.Equal(t, "bar", result.Path["foo"].Value)
			},
		},
		{
			name: "path parameter and trailing slash in request",
			param: []*openapi.Parameter{
				{
					Name:    "foo",
					Type:    openapi.ParameterPath,
					Schema:  schematest.New("string"),
					Style:   "",
					Explode: new(false),
				},
			},
			route: "/{foo}",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar/bar/", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error, _ *test.Hook) {
				require.NoError(t, err)
				require.Equal(t, "bar", result.Path["foo"].Value)
			},
		},
		{
			name: "to simple path parameter in same segment without separation should warn",
			param: []*openapi.Parameter{
				{
					Name:    "foo",
					Type:    openapi.ParameterPath,
					Schema:  schematest.New("string"),
					Explode: new(false),
				},
				{
					Name:    "bar",
					Type:    openapi.ParameterPath,
					Schema:  schematest.New("string"),
					Explode: new(false),
				},
			},
			route: "/{foo}{bar}",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar/bar.json", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error, _ *test.Hook) {
				require.NoError(t, err)
				require.Equal(t, "bar.json", result.Path["foo"].Value)
				require.Equal(t, "", result.Path["bar"].Value)
			},
		},
		{
			name: "to simple path parameter in same segment with separation should warn",
			param: []*openapi.Parameter{
				{
					Name:    "foo",
					Type:    openapi.ParameterPath,
					Schema:  schematest.New("string"),
					Explode: new(false),
				},
				{
					Name:    "bar",
					Type:    openapi.ParameterPath,
					Schema:  schematest.New("string"),
					Explode: new(false),
				},
			},
			route: "/{foo}.{bar}",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar/bar.json", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error, _ *test.Hook) {
				require.NoError(t, err)
				require.Equal(t, "bar", result.Path["foo"].Value)
				require.Equal(t, "json", result.Path["bar"].Value)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			log.SetLevel(log.WarnLevel)
			hook := test.NewGlobal()

			var params openapi.Parameters
			for _, p := range tc.param {
				params = append(params, &openapi.ParameterRef{Value: p})
			}
			r, err := openapi.FromRequest(params, tc.route, tc.request())
			tc.test(t, r, err, hook)
		})
	}
}
