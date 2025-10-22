package openapi_test

import (
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/schema/schematest"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseQuery(t *testing.T) {
	testcases := []struct {
		name    string
		params  openapi.Parameters
		request func() *http.Request
		test    func(t *testing.T, result *openapi.RequestParameters, err error)
	}{
		{
			name: "integer",
			params: openapi.Parameters{
				{Value: &openapi.Parameter{
					Name:   "id",
					Type:   openapi.ParameterQuery,
					Schema: schematest.New("integer"),
					Style:  "form",
				}},
			},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar?id=5", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(5), result.Query["id"].Value)
				require.Equal(t, "5", *result.Query["id"].Raw)
			},
		},
		{
			name: "string with whitespace",
			params: openapi.Parameters{
				{Value: &openapi.Parameter{
					Name:   "foo",
					Type:   openapi.ParameterQuery,
					Schema: schematest.New("string"),
					Style:  "form",
				}},
			},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar?foo=Hello%20World", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, "Hello World", result.Query["foo"].Value)
			},
		},
		{
			name: "no query parameter",
			params: openapi.Parameters{
				{Value: &openapi.Parameter{
					Name:    "id",
					Type:    openapi.ParameterQuery,
					Schema:  schematest.New("integer"),
					Style:   "form",
					Explode: explode(false),
				}},
			},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error) {
				require.NoError(t, err)
				require.Len(t, result.Query, 1)
				require.Nil(t, result.Query["id"].Value)
				require.Nil(t, result.Query["id"].Raw)
			},
		},
		{
			name: "no query parameter but required",
			params: openapi.Parameters{
				{Value: &openapi.Parameter{
					Name:     "id",
					Type:     openapi.ParameterQuery,
					Schema:   schematest.New("integer"),
					Required: true,
					Style:    "form",
				}},
			},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error) {
				require.EqualError(t, err, "parse query parameter 'id' failed: parameter is required")
			},
		},
		{
			name: "integer array as form and explode",
			params: openapi.Parameters{
				{Value: &openapi.Parameter{
					Name:    "id",
					Type:    openapi.ParameterQuery,
					Schema:  schematest.New("array", schematest.WithItems("integer")),
					Style:   "form",
					Explode: explode(true),
				}},
			},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar?id=3&id=4&id=5", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int64(3), int64(4), int64(5)}, result.Query["id"].Value)
				require.Equal(t, "3,4,5", *result.Query["id"].Raw)
			},
		},
		{
			name: "integer array as form and not explode",
			params: openapi.Parameters{
				{Value: &openapi.Parameter{
					Name:    "id",
					Type:    openapi.ParameterQuery,
					Schema:  schematest.New("array", schematest.WithItems("integer")),
					Style:   "form",
					Explode: explode(false),
				}},
			},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar?id=3,4,5", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int64(3), int64(4), int64(5)}, result.Query["id"].Value)
			},
		},
		{
			name: "integer array space delimited and explode",
			params: openapi.Parameters{
				{Value: &openapi.Parameter{
					Name:    "id",
					Type:    openapi.ParameterQuery,
					Schema:  schematest.New("array", schematest.WithItems("integer")),
					Style:   "spaceDelimited",
					Explode: explode(true),
				}},
			},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar?id=3&id=4&id=5", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int64(3), int64(4), int64(5)}, result.Query["id"].Value)
			},
		},
		{
			name: "integer array space delimited and not explode",
			params: openapi.Parameters{
				{Value: &openapi.Parameter{
					Name:    "id",
					Type:    openapi.ParameterQuery,
					Schema:  schematest.New("array", schematest.WithItems("integer")),
					Style:   "spaceDelimited",
					Explode: explode(false),
				}},
			},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar?id=3%204%205", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int64(3), int64(4), int64(5)}, result.Query["id"].Value)
			},
		},
		{
			name: "integer array pipe delimited and explode",
			params: openapi.Parameters{
				{Value: &openapi.Parameter{
					Name:    "id",
					Type:    openapi.ParameterQuery,
					Schema:  schematest.New("array", schematest.WithItems("integer")),
					Style:   "pipeDelimited",
					Explode: explode(true),
				}},
			},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar?id=3&id=4&id=5", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int64(3), int64(4), int64(5)}, result.Query["id"].Value)
			},
		},
		{
			name: "integer array pipe delimited and not explode",
			params: openapi.Parameters{
				{Value: &openapi.Parameter{
					Name:    "id",
					Type:    openapi.ParameterQuery,
					Schema:  schematest.New("array", schematest.WithItems("integer")),
					Style:   "pipeDelimited",
					Explode: explode(false),
				}},
			},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar?id=3|4|5", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int64(3), int64(4), int64(5)}, result.Query["id"].Value)
			},
		},
		{
			name: "object explode",
			params: openapi.Parameters{
				{Value: &openapi.Parameter{
					Name: "id",
					Type: openapi.ParameterQuery,
					Schema: schematest.New("object",
						schematest.WithProperty("role", schematest.New("string")),
						schematest.WithProperty("firstName", schematest.New("string")),
						schematest.WithProperty("msg", schematest.New("string")),
						schematest.WithProperty("foo", schematest.New("string")),
					),
					Style:   "form",
					Explode: explode(true),
				}},
			},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar?role=admin&firstName=Alex&msg=Hello%20World&foo=foo%26bar", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"role": "admin", "firstName": "Alex", "msg": "Hello World", "foo": "foo&bar"}, result.Query["id"].Value)
				require.Equal(t, "role=admin&firstName=Alex&msg=Hello%20World&foo=foo%26bar", *result.Query["id"].Raw)
			},
		},
		{
			name: "object explode and required but no query",
			params: openapi.Parameters{
				{Value: &openapi.Parameter{
					Name: "id",
					Type: openapi.ParameterQuery,
					Schema: schematest.New("object",
						schematest.WithProperty("role", schematest.New("string")),
						schematest.WithProperty("firstName", schematest.New("string")),
					),
					Required: true,
					Style:    "form",
					Explode:  explode(true),
				}},
			},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error) {
				require.EqualError(t, err, "parse query parameter 'id' failed: parameter is required")
			},
		},
		{
			name: "object not explode and required but no query",
			params: openapi.Parameters{
				{Value: &openapi.Parameter{
					Name: "id",
					Type: openapi.ParameterQuery,
					Schema: schematest.New("object",
						schematest.WithProperty("role", schematest.New("string")),
						schematest.WithProperty("firstName", schematest.New("string")),
					),
					Required: true,
					Style:    "form",
					Explode:  explode(false),
				}},
			},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error) {
				require.EqualError(t, err, "parse query parameter 'id' failed: parameter is required")
			},
		},
		{
			name: "free form object explode",
			params: openapi.Parameters{
				{Value: &openapi.Parameter{
					Name: "id",
					Type: openapi.ParameterQuery,
					Schema: schematest.New("object",
						schematest.WithProperty("role", schematest.New("string")),
					),
					Style:   "form",
					Explode: explode(true),
				}},
			},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar?role=admin&firstName=Alex", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"role": "admin", "firstName": "Alex"}, result.Query["id"].Value)
			},
		},
		{
			name: "not free form object explode but with extra property",
			params: openapi.Parameters{
				{Value: &openapi.Parameter{
					Name: "id",
					Type: openapi.ParameterQuery,
					Schema: schematest.New("object",
						schematest.WithProperty("role", schematest.New("string")),
						schematest.WithFreeForm(false),
					),
					Style:   "form",
					Explode: explode(true),
				}},
			},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar?role=admin&firstName=Alex", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error) {
				require.EqualError(t, err, "parse query parameter 'id' failed: property 'firstName' not defined in schema: schema type=object properties=[role] free-form=false")
			},
		},
		{
			name: "dictionary explode",
			params: openapi.Parameters{
				{Value: &openapi.Parameter{
					Name:    "id",
					Type:    openapi.ParameterQuery,
					Schema:  schematest.New("object", schematest.WithAdditionalProperties(schematest.New("string"))),
					Style:   "form",
					Explode: explode(true),
				}},
			},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar?role=admin&firstName=Alex", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"role": "admin", "firstName": "Alex"}, result.Query["id"].Value)
			},
		},
		{
			name: "object",
			params: openapi.Parameters{
				{Value: &openapi.Parameter{
					Name: "id",
					Type: openapi.ParameterQuery,
					Schema: schematest.New("object",
						schematest.WithProperty("role", schematest.New("string")),
						schematest.WithProperty("firstName", schematest.New("string")),
					),
					Style:   "form",
					Explode: explode(false),
				}},
			},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar?id=role,admin,firstName,Alex", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"role": "admin", "firstName": "Alex"}, result.Query["id"].Value)
			},
		},
		{
			name: "deepObject",
			params: openapi.Parameters{
				{Value: &openapi.Parameter{
					Name: "id",
					Type: openapi.ParameterQuery,
					Schema: schematest.New("object",
						schematest.WithProperty("role", schematest.New("string")),
						schematest.WithProperty("firstName", schematest.New("string")),
					),
					Style:   "deepObject",
					Explode: explode(true),
				}},
			},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar?id[role]=admin&id[firstName]=Alex&id[lastName]=Smith", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"role": "admin", "firstName": "Alex", "lastName": "Smith"}, result.Query["id"].Value)
			},
		},
		{
			name: "deepObject but not free-form",
			params: openapi.Parameters{
				{Value: &openapi.Parameter{
					Name: "id",
					Type: openapi.ParameterQuery,
					Schema: schematest.New("object",
						schematest.WithProperty("role", schematest.New("string")),
						schematest.WithProperty("firstName", schematest.New("string")),
						schematest.WithFreeForm(false),
					),
					Style:   "deepObject",
					Explode: explode(true),
				}},
			},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar?id[role]=admin&id[firstName]=Alex&id[lastName]=Smith", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error) {
				require.EqualError(t, err, "parse query parameter 'id' failed: property 'lastName' not defined in schema: schema type=object properties=[role, firstName] free-form=false")
			},
		},
		{
			name: "deepObject invalid format",
			params: openapi.Parameters{
				{Value: &openapi.Parameter{
					Name: "id",
					Type: openapi.ParameterQuery,
					Schema: schematest.New("object",
						schematest.WithProperty("role", schematest.New("string")),
						schematest.WithProperty("age", schematest.New("integer")),
					),
					Style:   "deepObject",
					Explode: explode(true),
				}},
			},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar?id[role]=admin&id[age]=foo&id[lastName]=Smith", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error) {
				require.EqualError(t, err, "parse query parameter 'id' failed: error count 1:\n\t- #/type: invalid type, expected integer but got string")
			},
		},
		{
			name: "deepObject required but no query",
			params: openapi.Parameters{
				{Value: &openapi.Parameter{
					Name: "id",
					Type: openapi.ParameterQuery,
					Schema: schematest.New("object",
						schematest.WithProperty("role", schematest.New("string")),
						schematest.WithProperty("firstName", schematest.New("string")),
					),
					Required: true,
					Style:    "deepObject",
					Explode:  explode(true),
				}},
			},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error) {
				require.EqualError(t, err, "parse query parameter 'id' failed: parameter is required")
			},
		},
		{
			name: "boolean value true",
			params: openapi.Parameters{
				{Value: &openapi.Parameter{
					Name:     "enabled",
					Type:     openapi.ParameterQuery,
					Schema:   schematest.New("boolean"),
					Required: true,
				}},
			},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar?enabled=true", nil)
			},
			test: func(t *testing.T, result *openapi.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, true, result.Query["enabled"].Value)
				require.Equal(t, "true", *(result.Query["enabled"].Raw))
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			r, err := openapi.FromRequest(tc.params, "", tc.request())
			tc.test(t, r, err)
		})

	}
}
