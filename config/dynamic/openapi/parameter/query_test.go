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

func TestParseQuery(t *testing.T) {
	testcases := []struct {
		name    string
		params  parameter.Parameters
		request func() *http.Request
		test    func(t *testing.T, result parameter.RequestParameters, err error)
	}{
		{
			name: "integer",
			params: parameter.Parameters{
				{Value: &parameter.Parameter{
					Name:   "id",
					Type:   parameter.Query,
					Schema: &schema.Ref{Value: &schema.Schema{Type: "integer"}},
					Style:  "form",
				}},
			},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar?id=5", nil)
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(5), result[parameter.Query]["id"].Value)
			},
		},
		{
			name: "no query parameter",
			params: parameter.Parameters{
				{Value: &parameter.Parameter{
					Name:    "id",
					Type:    parameter.Query,
					Schema:  &schema.Ref{Value: &schema.Schema{Type: "integer"}},
					Style:   "form",
					Explode: explode(false),
				}},
			},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar", nil)
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.NoError(t, err)
				require.Len(t, result[parameter.Query], 0)
			},
		},
		{
			name: "no query parameter but required",
			params: parameter.Parameters{
				{Value: &parameter.Parameter{
					Name:     "id",
					Type:     parameter.Query,
					Schema:   &schema.Ref{Value: &schema.Schema{Type: "integer"}},
					Required: true,
					Style:    "form",
				}},
			},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar", nil)
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.EqualError(t, err, "parse query parameter 'id' failed: parameter is required")
				require.Len(t, result[parameter.Query], 0)
			},
		},
		{
			name: "integer array as form and explode",
			params: parameter.Parameters{
				{Value: &parameter.Parameter{
					Name:    "id",
					Type:    parameter.Query,
					Schema:  &schema.Ref{Value: &schema.Schema{Type: "array", Items: &schema.Ref{Value: &schema.Schema{Type: "integer"}}}},
					Style:   "form",
					Explode: explode(true),
				}},
			},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar?id=3&id=4&id=5", nil)
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int64(3), int64(4), int64(5)}, result[parameter.Query]["id"].Value)
			},
		},
		{
			name: "integer array as form and not explode",
			params: parameter.Parameters{
				{Value: &parameter.Parameter{
					Name:    "id",
					Type:    parameter.Query,
					Schema:  &schema.Ref{Value: &schema.Schema{Type: "array", Items: &schema.Ref{Value: &schema.Schema{Type: "integer"}}}},
					Style:   "form",
					Explode: explode(false),
				}},
			},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar?id=3,4,5", nil)
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int64(3), int64(4), int64(5)}, result[parameter.Query]["id"].Value)
			},
		},
		{
			name: "integer array space delimited and explode",
			params: parameter.Parameters{
				{Value: &parameter.Parameter{
					Name:    "id",
					Type:    parameter.Query,
					Schema:  &schema.Ref{Value: &schema.Schema{Type: "array", Items: &schema.Ref{Value: &schema.Schema{Type: "integer"}}}},
					Style:   "spaceDelimited",
					Explode: explode(true),
				}},
			},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar?id=3&id=4&id=5", nil)
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int64(3), int64(4), int64(5)}, result[parameter.Query]["id"].Value)
			},
		},
		{
			name: "integer array space delimited and not explode",
			params: parameter.Parameters{
				{Value: &parameter.Parameter{
					Name:    "id",
					Type:    parameter.Query,
					Schema:  &schema.Ref{Value: &schema.Schema{Type: "array", Items: &schema.Ref{Value: &schema.Schema{Type: "integer"}}}},
					Style:   "spaceDelimited",
					Explode: explode(false),
				}},
			},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar?id=3%204%205", nil)
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int64(3), int64(4), int64(5)}, result[parameter.Query]["id"].Value)
			},
		},
		{
			name: "integer array pipe delimited and explode",
			params: parameter.Parameters{
				{Value: &parameter.Parameter{
					Name:    "id",
					Type:    parameter.Query,
					Schema:  &schema.Ref{Value: &schema.Schema{Type: "array", Items: &schema.Ref{Value: &schema.Schema{Type: "integer"}}}},
					Style:   "pipeDelimited",
					Explode: explode(true),
				}},
			},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar?id=3&id=4&id=5", nil)
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int64(3), int64(4), int64(5)}, result[parameter.Query]["id"].Value)
			},
		},
		{
			name: "integer array pipe delimited and not explode",
			params: parameter.Parameters{
				{Value: &parameter.Parameter{
					Name:    "id",
					Type:    parameter.Query,
					Schema:  &schema.Ref{Value: &schema.Schema{Type: "array", Items: &schema.Ref{Value: &schema.Schema{Type: "integer"}}}},
					Style:   "pipeDelimited",
					Explode: explode(false),
				}},
			},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar?id=3|4|5", nil)
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int64(3), int64(4), int64(5)}, result[parameter.Query]["id"].Value)
			},
		},
		{
			name: "object explode",
			params: parameter.Parameters{
				{Value: &parameter.Parameter{
					Name: "id",
					Type: parameter.Query,
					Schema: &schema.Ref{Value: schematest.New("object",
						schematest.WithProperty("role", schematest.New("string")),
						schematest.WithProperty("firstName", schematest.New("string")),
					)},
					Style:   "form",
					Explode: explode(true),
				}},
			},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar?role=admin&firstName=Alex", nil)
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"role": "admin", "firstName": "Alex"}, result[parameter.Query]["id"].Value)
			},
		},
		{
			name: "object explode and required but no query",
			params: parameter.Parameters{
				{Value: &parameter.Parameter{
					Name: "id",
					Type: parameter.Query,
					Schema: &schema.Ref{Value: schematest.New("object",
						schematest.WithProperty("role", schematest.New("string")),
						schematest.WithProperty("firstName", schematest.New("string")),
					)},
					Required: true,
					Style:    "form",
					Explode:  explode(true),
				}},
			},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar?", nil)
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.EqualError(t, err, "parse query parameter 'id' failed: parameter is required")
				require.Len(t, result[parameter.Query], 0)
			},
		},
		{
			name: "free form object explode",
			params: parameter.Parameters{
				{Value: &parameter.Parameter{
					Name: "id",
					Type: parameter.Query,
					Schema: &schema.Ref{Value: schematest.New("object",
						schematest.WithProperty("role", schematest.New("string")),
					)},
					Style:   "form",
					Explode: explode(true),
				}},
			},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar?role=admin&firstName=Alex", nil)
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"role": "admin", "firstName": "Alex"}, result[parameter.Query]["id"].Value)
			},
		},
		{
			name: "not free form object explode but with extra property",
			params: parameter.Parameters{
				{Value: &parameter.Parameter{
					Name: "id",
					Type: parameter.Query,
					Schema: &schema.Ref{Value: schematest.New("object",
						schematest.WithProperty("role", schematest.New("string")),
						schematest.WithFreeForm(false),
					)},
					Style:   "form",
					Explode: explode(true),
				}},
			},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar?role=admin&firstName=Alex", nil)
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.EqualError(t, err, "parse query parameter 'id' failed: property 'firstName' not defined in schema: schema type=object properties=[role] free-form=false")
			},
		},
		{
			name: "dictionary explode",
			params: parameter.Parameters{
				{Value: &parameter.Parameter{
					Name:    "id",
					Type:    parameter.Query,
					Schema:  &schema.Ref{Value: schematest.New("object", schematest.WithAdditionalProperties(schematest.New("string")))},
					Style:   "form",
					Explode: explode(true),
				}},
			},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar?role=admin&firstName=Alex", nil)
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"role": "admin", "firstName": "Alex"}, result[parameter.Query]["id"].Value)
			},
		},
		{
			name: "object",
			params: parameter.Parameters{
				{Value: &parameter.Parameter{
					Name: "id",
					Type: parameter.Query,
					Schema: &schema.Ref{Value: schematest.New("object",
						schematest.WithProperty("role", schematest.New("string")),
						schematest.WithProperty("firstName", schematest.New("string")),
					)},
					Style:   "form",
					Explode: explode(false),
				}},
			},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar?id=role,admin,firstName,Alex", nil)
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"role": "admin", "firstName": "Alex"}, result[parameter.Query]["id"].Value)
			},
		},
		{
			name: "deepObject",
			params: parameter.Parameters{
				{Value: &parameter.Parameter{
					Name: "id",
					Type: parameter.Query,
					Schema: &schema.Ref{Value: schematest.New("object",
						schematest.WithProperty("role", schematest.New("string")),
						schematest.WithProperty("firstName", schematest.New("string")),
					)},
					Style:   "deepObject",
					Explode: explode(true),
				}},
			},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar?id[role]=admin&id[firstName]=Alex&id[lastName]=Smith", nil)
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"role": "admin", "firstName": "Alex", "lastName": "Smith"}, result[parameter.Query]["id"].Value)
			},
		},
		{
			name: "deepObject but not free-form",
			params: parameter.Parameters{
				{Value: &parameter.Parameter{
					Name: "id",
					Type: parameter.Query,
					Schema: &schema.Ref{Value: schematest.New("object",
						schematest.WithProperty("role", schematest.New("string")),
						schematest.WithProperty("firstName", schematest.New("string")),
						schematest.WithFreeForm(false),
					)},
					Style:   "deepObject",
					Explode: explode(true),
				}},
			},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar?id[role]=admin&id[firstName]=Alex&id[lastName]=Smith", nil)
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.EqualError(t, err, "parse query parameter 'id' failed: property 'lastName' not defined in schema: schema type=object properties=[role, firstName] free-form=false")
				require.Len(t, result[parameter.Query], 0)
			},
		},
		{
			name: "deepObject invalid format",
			params: parameter.Parameters{
				{Value: &parameter.Parameter{
					Name: "id",
					Type: parameter.Query,
					Schema: &schema.Ref{Value: schematest.New("object",
						schematest.WithProperty("role", schematest.New("string")),
						schematest.WithProperty("age", schematest.New("integer")),
					)},
					Style:   "deepObject",
					Explode: explode(true),
				}},
			},
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "https://foo.bar?id[role]=admin&id[age]=foo&id[lastName]=Smith", nil)
			},
			test: func(t *testing.T, result parameter.RequestParameters, err error) {
				require.EqualError(t, err, "parse query parameter 'id' failed: could not parse 'foo' as int, expected schema type=integer")
				require.Len(t, result[parameter.Query], 0)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			r, err := parameter.FromRequest(tc.params, "", tc.request())
			tc.test(t, r, err)
		})

	}
}
