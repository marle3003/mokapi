package openapi_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	"mokapi/providers/openapi/schema"
	"mokapi/providers/openapi/schema/schematest"
	"mokapi/runtime/events"
	"mokapi/schema/json/generator"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_Response(t *testing.T) {
	getConfig := func(s *schema.Schema, contentType string) *openapi.Config {
		op := openapitest.NewOperation(
			openapitest.WithResponse(http.StatusOK, openapitest.WithContent(contentType,
				openapitest.NewContent(openapitest.WithSchema(s)))),
		)

		return openapitest.NewConfig("3.0",
			openapitest.WithPath("/foo", openapitest.NewPath(openapitest.WithOperation(http.MethodGet, op))))
	}

	testcases := []struct {
		name    string
		config  *openapi.Config
		handler func(event string, req *common.EventRequest, res *common.EventResponse)
		req     func() *http.Request
		test    func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name:   "string as response body",
			config: getConfig(schematest.New("string"), "application/json"),
			handler: func(event string, req *common.EventRequest, res *common.EventResponse) {
				res.Body = "foo"
			},
			req: func() *http.Request {
				return httptest.NewRequest("get", "http://localhost/foo", nil)
			},
			test: func(t *testing.T, rr *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, rr.Code)
				require.Equal(t, "foo", rr.Body.String())
			},
		},
		{
			name:   "invalid string body",
			config: getConfig(schematest.New("string", schematest.WithFormat("date")), "application/json"),
			handler: func(event string, req *common.EventRequest, res *common.EventResponse) {
				res.Data = "foo"
			},
			req: func() *http.Request {
				return httptest.NewRequest("get", "http://localhost/foo", nil)
			},
			test: func(t *testing.T, rr *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, rr.Code)
				require.Equal(t, "encoding data to 'application/json' failed: error count 1:\n\t- #/format: string 'foo' does not match format 'date'\n", rr.Body.String())
			},
		},
		{
			name:   "object with null property",
			config: getConfig(schematest.New("object", schematest.WithProperty("foo", schematest.New("string", schematest.IsNullable(true)))), "application/json"),
			handler: func(event string, req *common.EventRequest, res *common.EventResponse) {
				res.Data = map[string]interface{}{"foo": nil}
			},
			req: func() *http.Request {
				return httptest.NewRequest("get", "http://localhost/foo", nil)
			},
			test: func(t *testing.T, rr *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, rr.Code)
				require.Equal(t, `{"foo":null}`, rr.Body.String())
			},
		},
		{
			name:   "detect content type on byte array",
			config: getConfig(schematest.New("object", schematest.WithProperty("foo", schematest.New("string"))), "*/*"),
			handler: func(event string, req *common.EventRequest, res *common.EventResponse) {
				res.Data = []byte(`{"foo":"bar"}`)
			},
			req: func() *http.Request {
				return httptest.NewRequest("get", "http://localhost/foo", nil)
			},
			test: func(t *testing.T, rr *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, rr.Code)
				require.Equal(t, "application/json", rr.Header().Get("Content-Type"))
				require.Equal(t, `{"foo":"bar"}`, rr.Body.String())
			},
		},
		{
			name:   "application/octet-stream with string",
			config: getConfig(schematest.New("string"), "application/octet-stream"),
			handler: func(event string, req *common.EventRequest, res *common.EventResponse) {
				res.Data = "foo"
			},
			req: func() *http.Request {
				return httptest.NewRequest("get", "http://localhost/foo", nil)
			},
			test: func(t *testing.T, rr *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, rr.Code)
				require.Equal(t, "foo", rr.Body.String())
			},
		},
		{
			name:   "application/octet-stream with object",
			config: getConfig(schematest.New("object"), "application/octet-stream"),
			handler: func(event string, req *common.EventRequest, res *common.EventResponse) {
				res.Data = map[string]interface{}{"foo": "bar"}
			},
			req: func() *http.Request {
				return httptest.NewRequest("get", "http://localhost/foo", nil)
			},
			test: func(t *testing.T, rr *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, rr.Code)
				require.Equal(t, "encoding data to 'application/octet-stream' failed: not supported encoding of content types 'application/octet-stream', except simple data types\n", rr.Body.String())
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			events.SetStore(10, events.NewTraits().WithNamespace("http"))
			defer events.Reset()

			e := &engine{emit: func(event string, args ...interface{}) []*common.Action {
				tc.handler(event, args[0].(*common.EventRequest), args[1].(*common.EventResponse))
				return nil
			}}

			h := openapi.NewHandler(tc.config, e)
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, tc.req())

			tc.test(t, rr)
		})
	}
}

func TestHandler_Response_Context(t *testing.T) {
	testcases := []struct {
		name string
		opt  openapitest.ConfigOptions
		req  func() *http.Request
		test func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name: "use query data in random response",
			opt: openapitest.WithPath("/foo",
				openapitest.NewPath(openapitest.WithOperation(http.MethodGet, openapitest.NewOperation(
					openapitest.WithQueryParam("name", false, openapitest.WithParamSchema(schematest.New("string"))),
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json",
						openapitest.NewContent(openapitest.WithSchema(
							schematest.New("object", schematest.WithProperty("name", schematest.New("string")))),
						),
					)),
				)))),
			req: func() *http.Request {
				return httptest.NewRequest("get", "http://localhost/foo?name=foo", nil)
			},
			test: func(t *testing.T, rr *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, rr.Code)
				require.Equal(t, `{"name":"foo"}`, rr.Body.String())
			},
		},
		{
			name: "use from path parameter",
			opt: openapitest.WithPath("/users/{id}",
				openapitest.NewPath(
					openapitest.WithPathParam("id", openapitest.WithParamSchema(schematest.New("integer"))),
					openapitest.WithOperation(http.MethodGet, openapitest.NewOperation(
						openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json",
							openapitest.NewContent(openapitest.WithSchema(
								schematest.New("object", schematest.WithProperty("id", schematest.New("integer")))),
							),
						)),
					))),
			),
			req: func() *http.Request {
				return httptest.NewRequest("get", "http://localhost/users/123", nil)
			},
			test: func(t *testing.T, rr *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, rr.Code)
				require.Equal(t, `{"id":123}`, rr.Body.String())
			},
		},
		{
			name: "parameter does not match response type",
			opt: openapitest.WithPath("/users/{id}",
				openapitest.NewPath(
					openapitest.WithPathParam("id", openapitest.WithParamSchema(schematest.New("integer"))),
					openapitest.WithOperation(http.MethodGet, openapitest.NewOperation(
						openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json",
							openapitest.NewContent(openapitest.WithSchema(
								schematest.New("object", schematest.WithProperty("id", schematest.New("string")))),
							),
						)),
					))),
			),
			req: func() *http.Request {
				return httptest.NewRequest("get", "http://localhost/users/123", nil)
			},
			test: func(t *testing.T, rr *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, rr.Code)
				require.Equal(t, `{"id":"98173564-6619-4557-888e-65b16bb5def5"}`, rr.Body.String())
			},
		},
		{
			name: "parameter does not match response type but string to int",
			opt: openapitest.WithPath("/users/{id}",
				openapitest.NewPath(
					openapitest.WithPathParam("id", openapitest.WithParamSchema(schematest.New("string"))),
					openapitest.WithOperation(http.MethodGet, openapitest.NewOperation(
						openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json",
							openapitest.NewContent(openapitest.WithSchema(
								schematest.New("object", schematest.WithProperty("id", schematest.New("integer")))),
							),
						)),
					))),
			),
			req: func() *http.Request {
				return httptest.NewRequest("get", "http://localhost/users/123", nil)
			},
			test: func(t *testing.T, rr *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, rr.Code)
				require.Equal(t, `{"id":123}`, rr.Body.String())
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			generator.Seed(11)
			config := openapitest.NewConfig("3.0", tc.opt)

			h := openapi.NewHandler(config, enginetest.NewEngine())
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, tc.req())

			tc.test(t, rr)
		})
	}
}
