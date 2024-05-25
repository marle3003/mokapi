package openapi_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/engine/common"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	"mokapi/providers/openapi/schema"
	"mokapi/providers/openapi/schema/schematest"
	"mokapi/runtime/events"
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
				require.Equal(t, "encoding data to 'application/json' failed: value 'foo' does not match format 'date' (RFC3339), expected schema type=string format=date\n", rr.Body.String())
			},
		},
		{
			name:   "object with null property",
			config: getConfig(schematest.New("object", schematest.WithProperty("foo", schematest.New("string", schematest.IsNullable(true)))), "application/json"),
			handler: func(event string, req *common.EventRequest, res *common.EventResponse) {
				res.Data = &struct {
					Foo interface{}
				}{
					Foo: nil,
				}
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
