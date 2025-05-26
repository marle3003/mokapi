package openapi_test

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	"mokapi/providers/openapi/schema/schematest"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type errReader int

func (errReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("TESTING ERROR")
}

func TestRequestBody_UnmarshalJSON(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "request bodies",
			test: func(t *testing.T) {
				r := openapi.RequestBodies{}
				err := json.Unmarshal([]byte(`{ "foo": {"description": "foo"} }`), &r)
				require.NoError(t, err)
				require.Len(t, r, 1)
				require.Contains(t, r, "foo")
				require.Equal(t, "foo", r["foo"].Value.Description)
			},
		},
		{
			name: "request body reference",
			test: func(t *testing.T) {
				r := openapi.RequestBodies{}
				err := json.Unmarshal([]byte(`{ "foo": {"$ref": "foo.yml"} }`), &r)
				require.NoError(t, err)
				require.Len(t, r, 1)
				require.Contains(t, r, "foo")
				require.Equal(t, "foo.yml", r["foo"].Ref)
				require.Nil(t, r["foo"].Value)
			},
		},
		{
			name: "request body",
			test: func(t *testing.T) {
				r := openapi.RequestBody{}
				err := json.Unmarshal([]byte(`{ "description": "foo", "content": {"foo": {}}, "required": true }`), &r)
				require.NoError(t, err)
				require.Equal(t, "foo", r.Description)
				require.NotNil(t, r.Content)
				require.Contains(t, r.Content, "foo")
				require.True(t, r.Required)
			},
		},
	}
	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.test(t)
		})
	}
}

func TestRequestBody_UnmarshalYAML(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "request bodies",
			test: func(t *testing.T) {
				r := openapi.RequestBodies{}
				err := yaml.Unmarshal([]byte(`foo: {description: foo}`), &r)
				require.NoError(t, err)
				require.Len(t, r, 1)
				require.Contains(t, r, "foo")
				require.Equal(t, "foo", r["foo"].Value.Description)
			},
		},
		{
			name: "request body reference",
			test: func(t *testing.T) {
				r := openapi.RequestBodies{}
				err := yaml.Unmarshal([]byte(`foo: {$ref: foo.yml}`), &r)
				require.NoError(t, err)
				require.Len(t, r, 1)
				require.Contains(t, r, "foo")
				require.Equal(t, "foo.yml", r["foo"].Ref)
				require.Nil(t, r["foo"].Value)
			},
		},
		{
			name: "request body",
			test: func(t *testing.T) {
				r := openapi.RequestBody{}
				err := yaml.Unmarshal([]byte(`
description: foo
content: 
  foo: {}
required: true`), &r)
				require.NoError(t, err)
				require.Equal(t, "foo", r.Description)
				require.NotNil(t, r.Content)
				require.Contains(t, r.Content, "foo")
				require.True(t, r.Required)
			},
		},
	}
	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.test(t)
		})
	}
}

func TestBodyFromRequest(t *testing.T) {
	testcases := []struct {
		name      string
		operation *openapi.Operation
		request   func() *http.Request
		test      func(t *testing.T, result *openapi.Body, err error)
	}{
		{
			name: "no Content-Type header with matching MediaType",
			operation: openapitest.NewOperation(
				openapitest.WithRequestBody("foo", true,
					openapitest.WithRequestContent("application/json", openapitest.NewContent()),
				)),
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodPost, "https://foo.bar", strings.NewReader(`{"foo": "bar"}`))
			},
			test: func(t *testing.T, result *openapi.Body, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": "bar"}, result.Value)
			},
		},
		{
			name: "no Content-Type in header, requestBody specified as xml or json, request body is json",
			operation: openapitest.NewOperation(
				openapitest.WithRequestBody("foo", true,
					openapitest.WithRequestContent("application/xml", openapitest.NewContent()),
					openapitest.WithRequestContent("application/json", openapitest.NewContent()),
				)),
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodPost, "https://foo.bar", strings.NewReader(`{"foo": "bar"}`))
			},
			test: func(t *testing.T, result *openapi.Body, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": "bar"}, result.Value)
			},
		},
		{
			name: "no Content-Type in header, requestBody specified as xml or json, request body is xml",
			operation: openapitest.NewOperation(
				openapitest.WithRequestBody("foo", true,
					openapitest.WithRequestContent("application/xml", openapitest.NewContent()),
					openapitest.WithRequestContent("application/json", openapitest.NewContent()),
				)),
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodPost, "https://foo.bar", strings.NewReader(`<root><foo>bar</foo></root>`))
			},
			test: func(t *testing.T, result *openapi.Body, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": "bar"}, result.Value)
			},
		},
		{
			name: "no Content-Type in header, request body does not match",
			operation: openapitest.NewOperation(
				openapitest.WithRequestBody("foo", true,
					openapitest.WithRequestContent("application/json", openapitest.NewContent()),
				)),
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodPost, "https://foo.bar", strings.NewReader(`<root><foo>bar</foo></root>`))
			},
			test: func(t *testing.T, result *openapi.Body, err error) {
				require.EqualError(t, err, "read request body failed: no matching content type defined")
				require.Equal(t, "<root><foo>bar</foo></root>", result.Raw)
			},
		},
		{
			name: "no Content-Type in header, required but missing",
			operation: openapitest.NewOperation(
				openapitest.WithRequestBody("foo", true,
					openapitest.WithRequestContent("application/json", openapitest.NewContent()),
				)),
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodPost, "https://foo.bar", nil)
			},
			test: func(t *testing.T, result *openapi.Body, err error) {
				require.EqualError(t, err, "request body is required")
				require.Nil(t, result)
			},
		},
		{
			name: "no Content-Type in header, error reading body",
			operation: openapitest.NewOperation(
				openapitest.WithRequestBody("foo", true,
					openapitest.WithRequestContent("application/json", openapitest.NewContent()),
				)),
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodPost, "https://foo.bar", errReader(0))
			},
			test: func(t *testing.T, result *openapi.Body, err error) {
				require.EqualError(t, err, "read request body failed: TESTING ERROR")
				require.Nil(t, result)
			},
		},
		{
			name: "Content-Type in header does not match request body",
			operation: openapitest.NewOperation(
				openapitest.WithRequestBody("foo", true,
					openapitest.WithRequestContent("application/json", openapitest.NewContent()),
				)),
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodPost, "https://foo.bar", strings.NewReader(`<root><foo>bar</foo></root>`))
				r.Header.Set("Content-Type", "application/json")
				return r
			},
			test: func(t *testing.T, result *openapi.Body, err error) {
				require.EqualError(t, err, "read request body 'application/json' failed: invalid json format: invalid character '<' looking for beginning of value")
				require.Equal(t, "<root><foo>bar</foo></root>", result.Raw)
			},
		},
		{
			name: "Content-Type in header, error while reading body",
			operation: openapitest.NewOperation(
				openapitest.WithRequestBody("foo", true,
					openapitest.WithRequestContent("application/json", openapitest.NewContent()),
				)),
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodPost, "https://foo.bar", errReader(0))
				r.Header.Set("Content-Type", "application/json")
				return r
			},
			test: func(t *testing.T, result *openapi.Body, err error) {
				require.EqualError(t, err, "read request body 'application/json' failed: TESTING ERROR")
				require.Nil(t, result)
			},
		},
		{
			name: "no matching MediaType for Content-Type",
			operation: openapitest.NewOperation(
				openapitest.WithRequestBody("foo", true,
					openapitest.WithRequestContent("application/json", openapitest.NewContent()),
				)),
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodPost, "https://foo.bar", strings.NewReader("<root><foo>bar</foo></root>"))
				r.Header.Set("Content-Type", "application/xml")
				return r
			},
			test: func(t *testing.T, result *openapi.Body, err error) {
				require.EqualError(t, err, "read request body failed: no matching content type for 'application/xml' defined")
				require.Equal(t, "<root><foo>bar</foo></root>", result.Raw)
			},
		},
		{
			name: "no matching MediaType for Content-Type and error while reading body",
			operation: openapitest.NewOperation(
				openapitest.WithRequestBody("foo", true,
					openapitest.WithRequestContent("application/json", openapitest.NewContent()),
				)),
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodPost, "https://foo.bar", errReader(0))
				r.Header.Set("Content-Type", "application/xml")
				return r
			},
			test: func(t *testing.T, result *openapi.Body, err error) {
				require.EqualError(t, err, "read request body failed: TESTING ERROR")
				require.Nil(t, result)
			},
		},
		{
			name: "Content-Type text/plain MediaType text/*",
			operation: openapitest.NewOperation(
				openapitest.WithRequestBody("foo", true,
					openapitest.WithRequestContent("text/*", openapitest.NewContent()),
				)),
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodPost, "https://foo.bar", strings.NewReader("foobar"))
				r.Header.Set("Content-Type", "text/plain")
				return r
			},
			test: func(t *testing.T, result *openapi.Body, err error) {
				require.NoError(t, err)
				require.Equal(t, "foobar", result.Value)
			},
		},
		{
			name: "Content-Type text/plain MediaType */*",
			operation: openapitest.NewOperation(
				openapitest.WithRequestBody("foo", true,
					openapitest.WithRequestContent("*/*", openapitest.NewContent()),
				)),
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodPost, "https://foo.bar", strings.NewReader("foobar"))
				r.Header.Set("Content-Type", "text/plain")
				return r
			},
			test: func(t *testing.T, result *openapi.Body, err error) {
				require.NoError(t, err)
				require.Equal(t, "foobar", result.Value)
			},
		},
		{
			name: "Content-Type text/plain MediaType */* and application/*",
			operation: openapitest.NewOperation(
				openapitest.WithRequestBody("foo", true,
					openapitest.WithRequestContent("application/*", openapitest.NewContent(openapitest.WithSchema(schematest.New("integer", schematest.WithFormat("int32"))))),
					openapitest.WithRequestContent("*/*", openapitest.NewContent()),
				)),
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodPost, "https://foo.bar", strings.NewReader("12"))
				r.Header.Set("Content-Type", "application/json")
				return r
			},
			test: func(t *testing.T, result *openapi.Body, err error) {
				require.NoError(t, err)
				require.Equal(t, int32(12), result.Value)
			},
		},
		{
			name: "Content-Type text/plain MediaType text/html, text/html; charset=us-ascii and text/html; charset=utf-8, text/*",
			operation: openapitest.NewOperation(
				openapitest.WithRequestBody("foo", true,
					openapitest.WithRequestContent("text/html", openapitest.NewContent()),
					openapitest.WithRequestContent("text/html; charset=utf-8", openapitest.NewContent(
						openapitest.WithSchema(schematest.New("integer", schematest.WithFormat("int32")))),
					),
					openapitest.WithRequestContent("text/html; charset=us-ascii", openapitest.NewContent()),
					openapitest.WithRequestContent("text/*", openapitest.NewContent()),
				)),
			request: func() *http.Request {
				r := httptest.NewRequest(http.MethodPost, "https://foo.bar", strings.NewReader("12"))
				r.Header.Set("Content-Type", "text/html; charset=utf-8")
				return r
			},
			test: func(t *testing.T, result *openapi.Body, err error) {
				require.NoError(t, err)
				require.Equal(t, int32(12), result.Value)
			},
		},
		{
			name: "multipart/form-data",
			operation: openapitest.NewOperation(
				openapitest.WithRequestBody("foo", true,
					openapitest.WithRequestContent("multipart/form-data", openapitest.NewContent(
						openapitest.WithSchema(
							schematest.New("object",
								schematest.WithProperty("id", schematest.New("string", schematest.WithFormat("uuid"))),
								schematest.WithProperty("address", schematest.New("object",
									schematest.WithProperty("street", schematest.New("string")),
									schematest.WithProperty("city", schematest.New("string")),
								)),
								schematest.WithProperty("profileImage", schematest.New("string", schematest.WithFormat("binary"))),
							),
						)),
					))),
			request: func() *http.Request {
				body := strings.NewReader(`
--abcde12345
Content-Disposition: form-data; name="id"
Content-Type: text/plain

123e4567-e89b-12d3-a456-426655440000
--abcde12345
Content-Disposition: form-data; name="address"
Content-Type: application/json

{
  "street": "3, Garden St",
  "city": "Hillsbery, UT"
}
--abcde12345
Content-Disposition: form-data; name="profileImage"; filename="image1.png"
Content-Type: application/octet-stream

foobar
--abcde12345--
`)
				r := httptest.NewRequest(http.MethodPost, "https://foo.bar", body)
				r.Header.Set("Content-Type", "multipart/form-data; boundary=abcde12345")
				return r
			},
			test: func(t *testing.T, result *openapi.Body, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{
					"address": map[string]interface{}{
						"street": "3, Garden St",
						"city":   "Hillsbery, UT",
					},
					"id":           "123e4567-e89b-12d3-a456-426655440000",
					"profileImage": "foobar",
				}, result.Value)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			b, err := openapi.BodyFromRequest(tc.request(), tc.operation)
			tc.test(t, b, err)
		})
	}
}

func TestConfig_Patch_RequestBody(t *testing.T) {
	testcases := []struct {
		name    string
		configs []*openapi.Config
		test    func(t *testing.T, result *openapi.Config)
	}{
		{
			name: "add request body",
			configs: []*openapi.Config{
				{Components: openapi.Components{RequestBodies: map[string]*openapi.RequestBodyRef{}}},
				openapitest.NewConfig("1.0", openapitest.WithComponentRequestBody("foo", &openapi.RequestBody{Description: "foo"})),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Equal(t, "foo", result.Components.RequestBodies["foo"].Value.Description)
			},
		},
		{
			name: "request body ref is nil",
			configs: []*openapi.Config{
				{Components: openapi.Components{RequestBodies: map[string]*openapi.RequestBodyRef{}}},
				openapitest.NewConfig("1.0", openapitest.WithComponentRequestBodyRef("foo", nil)),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Nil(t, result.Components.RequestBodies["foo"])
			},
		},
		{
			name: "request body is nil",
			configs: []*openapi.Config{
				{Components: openapi.Components{RequestBodies: map[string]*openapi.RequestBodyRef{}}},
				openapitest.NewConfig("1.0", openapitest.WithComponentRequestBody("foo", nil)),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Nil(t, result.Components.RequestBodies["foo"])
			},
		},
		{
			name: "source request body is nil",
			configs: []*openapi.Config{
				openapitest.NewConfig("1.0", openapitest.WithComponentRequestBody("foo", nil)),
				openapitest.NewConfig("1.0", openapitest.WithComponentRequestBody("foo", &openapi.RequestBody{Description: "foo"})),
			},
			test: func(t *testing.T, result *openapi.Config) {
				require.Equal(t, "foo", result.Components.RequestBodies["foo"].Value.Description)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			c := tc.configs[0]
			for _, p := range tc.configs[1:] {
				c.Patch(p)
			}
			tc.test(t, c)
		})
	}
}

func TestBodyFromRequest_FormUrlEncoded(t *testing.T) {
	testcases := []struct {
		name string
		mt   *openapi.MediaType
		body string
		test func(t *testing.T, result *openapi.Body, err error)
	}{
		{
			name: "primitive fields",
			mt: openapitest.NewContent(
				openapitest.WithSchema(
					schematest.New("object",
						schematest.WithProperty("name", schematest.New("string")),
						schematest.WithProperty("fav_number", schematest.New("integer")),
					),
				),
			),
			body: "name=Amy+Smith&fav_number=42",
			test: func(t *testing.T, result *openapi.Body, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"name": "Amy Smith", "fav_number": int64(42)}, result.Value)
			},
		},
		{
			name: "array",
			mt: openapitest.NewContent(
				openapitest.WithSchema(
					schematest.New("object",
						schematest.WithProperty("array_name",
							schematest.New("array", schematest.WithItems("string"))),
					),
				),
			),
			body: "array_name=value1&array_name=value2",
			test: func(t *testing.T, result *openapi.Body, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"array_name": []interface{}{"value1", "value2"}}, result.Value)
			},
		},
		{
			name: "array with encoding explode false",
			mt: openapitest.NewContent(
				openapitest.WithEncoding("array_name", &openapi.Encoding{Explode: false}),
				openapitest.WithSchema(
					schematest.New("object",
						schematest.WithProperty("array_name",
							schematest.New("array", schematest.WithItems("string"))),
					),
				),
			),
			body: "array_name=value1,value2",
			test: func(t *testing.T, result *openapi.Body, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"array_name": []interface{}{"value1", "value2"}}, result.Value)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			op := openapitest.NewOperation(
				openapitest.WithRequestBody("foo", true,
					openapitest.WithRequestContent("application/x-www-form-urlencoded", tc.mt),
				))
			r := httptest.NewRequest(http.MethodPost, "https://foo.bar", strings.NewReader(tc.body))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			b, err := openapi.BodyFromRequest(r, op)
			tc.test(t, b, err)
		})
	}
}

func TestBodyFromRequest_MultiPartFormData(t *testing.T) {
	// https://tools.ietf.org/html/rfc2046#section-5.1
	//   The boundary delimiter line is then defined as a line
	//   consisting entirely of two hyphen characters ("-",
	//   decimal value 45) followed by the boundary parameter

	testcases := []struct {
		name     string
		mt       *openapi.MediaType
		boundary string
		body     string
		test     func(t *testing.T, result *openapi.Body, err error)
	}{
		{
			name: "primitive fields",
			mt: openapitest.NewContent(
				openapitest.WithSchema(
					schematest.New("object",
						schematest.WithProperty("meta",
							schematest.New("object",
								schematest.WithProperty("name", schematest.New("string")),
								schematest.WithProperty("fav_number", schematest.New("integer")),
							),
						),
					),
				),
			),
			boundary: "XXX",
			body: `--XXX
Content-Disposition: form-data; name="meta"
Content-Type: application/json

{ "name": "Amy Smith", "fav_number": 42 }
--XXX--

`,
			test: func(t *testing.T, result *openapi.Body, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"meta": map[string]interface{}{"name": "Amy Smith", "fav_number": int64(42)}}, result.Value)
			},
		},
		{

			name: "array",
			mt: openapitest.NewContent(
				openapitest.WithSchema(
					schematest.New("object",
						schematest.WithProperty("names",
							schematest.New("array", schematest.WithItems("string")),
						),
					),
				),
			),
			boundary: "XXX",
			body: `--XXX
Content-Disposition: form-data; name="names"
Content-Type: text/plain

Alice
--XXX
Content-Disposition: form-data; name="names"
Content-Type: text/plain

Bob
--XXX--

`,
			test: func(t *testing.T, result *openapi.Body, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"names": []interface{}{"Alice", "Bob"}}, result.Value)
			},
		},
		{

			name: "encoding",
			mt: openapitest.NewContent(
				openapitest.WithSchema(
					schematest.New("object",
						schematest.WithProperty("names",
							schematest.New("array", schematest.WithItems("string")),
						),
					),
				),
				openapitest.WithEncoding("names", &openapi.Encoding{ContentType: "text/plain"}),
			),
			boundary: "XXX",
			body: `--XXX
Content-Disposition: form-data; name="names"
Content-Type: text/plain

Alice
--XXX
Content-Disposition: form-data; name="names"
Content-Type: text/plain

Bob
--XXX--

`,
			test: func(t *testing.T, result *openapi.Body, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"names": []interface{}{"Alice", "Bob"}}, result.Value)
			},
		},
		{

			name: "encoding error",
			mt: openapitest.NewContent(
				openapitest.WithSchema(
					schematest.New("object",
						schematest.WithProperty("names",
							schematest.New("array", schematest.WithItems("string")),
						),
					),
				),
				openapitest.WithEncoding("names", &openapi.Encoding{ContentType: "text/plain"}),
			),
			boundary: "XXX",
			body: `--XXX
Content-Disposition: form-data; name="names"
Content-Type: text/plain; charset=utf-8

Alice
--XXX
Content-Disposition: form-data; name="names"
Content-Type: application/json

Bob
--XXX--

`,
			test: func(t *testing.T, result *openapi.Body, err error) {
				require.EqualError(t, err, "read request body 'multipart/form-data; boundary=XXX' failed: part 'names' does not match content type: text/plain")
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			op := openapitest.NewOperation(
				openapitest.WithRequestBody("foo", true,
					openapitest.WithRequestContent("multipart/form-data", tc.mt),
				))
			r := httptest.NewRequest(http.MethodPost, "https://foo.bar", strings.NewReader(tc.body))
			r.Header.Set("Content-Type", fmt.Sprintf("multipart/form-data; boundary=%s", tc.boundary))

			b, err := openapi.BodyFromRequest(r, op)
			tc.test(t, b, err)
		})
	}
}
