package openapi_test

import (
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/openapi"
	"mokapi/config/dynamic/openapi/openapitest"
	"mokapi/config/dynamic/openapi/schema/schematest"
	engine2 "mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/runtime/events"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestResponseHandler_ServeHTTP_ResponseBody(t *testing.T) {
	testcases := []struct {
		name   string
		config *openapi.Config
		fn     func(t *testing.T, handler http.Handler)
		check  func(t *testing.T, r *engine2.EventRequest)
	}{
		{
			name: "text/plain",
			config: openapitest.NewConfig("3.0.0",
				openapitest.WithServer("http://localhost", ""),
				openapitest.WithPath("/foo", openapitest.NewPath(
					openapitest.WithOperation("post",
						openapitest.NewOperation(
							openapitest.WithRequestBody("", false,
								openapitest.WithRequestContent(
									"text/plain", openapitest.NewContent(openapitest.WithSchema(schematest.New("string"))))),
							openapitest.WithResponse(200),
						)),
				)),
			),
			fn: func(t *testing.T, handler http.Handler) {
				r := httptest.NewRequest("post", "http://localhost/foo", strings.NewReader("foo"))
				r.Header.Set("Content-Type", "text/plain")
				rr := httptest.NewRecorder()

				handler.ServeHTTP(rr, r)

				require.Equal(t, 200, rr.Code)
			},
			check: func(t *testing.T, r *engine2.EventRequest) {
				require.Equal(t, "foo", r.Body)
			},
		},
		{
			name: "text/*",
			config: openapitest.NewConfig("3.0.0",
				openapitest.WithServer("http://localhost", ""),
				openapitest.WithPath("/foo", openapitest.NewPath(
					openapitest.WithOperation("post",
						openapitest.NewOperation(
							openapitest.WithRequestBody("", false,
								openapitest.WithRequestContent(
									"text/*", openapitest.NewContent(openapitest.WithSchema(schematest.New("string"))))),
							openapitest.WithResponse(200),
						)),
				)),
			),
			fn: func(t *testing.T, handler http.Handler) {
				r := httptest.NewRequest("post", "http://localhost/foo", strings.NewReader("foo"))
				r.Header.Set("Content-Type", "text/plain")
				rr := httptest.NewRecorder()

				handler.ServeHTTP(rr, r)

				require.Equal(t, 200, rr.Code)
			},
			check: func(t *testing.T, r *engine2.EventRequest) {
				require.Equal(t, "foo", r.Body)
			},
		},
		{
			name: "text/* > */*",
			config: openapitest.NewConfig("3.0.0",
				openapitest.WithServer("http://localhost", ""),
				openapitest.WithPath("/foo", openapitest.NewPath(
					openapitest.WithOperation("post",
						openapitest.NewOperation(
							openapitest.WithRequestBody("", false,
								openapitest.WithRequestContent(
									"*/*", openapitest.NewContent(openapitest.WithSchema(schematest.New("number")))),
								openapitest.WithRequestContent(
									"text/*", openapitest.NewContent(openapitest.WithSchema(schematest.New("string"))))),
							openapitest.WithResponse(200),
						)),
				)),
			),
			fn: func(t *testing.T, handler http.Handler) {
				r := httptest.NewRequest("post", "http://localhost/foo", strings.NewReader("foo"))
				r.Header.Set("Content-Type", "text/plain")
				rr := httptest.NewRecorder()

				handler.ServeHTTP(rr, r)

				require.Equal(t, 200, rr.Code)
			},
			check: func(t *testing.T, r *engine2.EventRequest) {
				require.Equal(t, "foo", r.Body)
			},
		},
		{
			name: "application/json free-form",
			config: openapitest.NewConfig("3.0.0",
				openapitest.WithServer("http://localhost", ""),
				openapitest.WithPath("/foo", openapitest.NewPath(
					openapitest.WithOperation("post",
						openapitest.NewOperation(
							openapitest.WithRequestBody("", false,
								openapitest.WithRequestContent(
									"application/json", openapitest.NewContent(openapitest.WithSchema(
										schematest.New("object"),
									)))),
							openapitest.WithResponse(200),
						)),
				)),
			),
			fn: func(t *testing.T, handler http.Handler) {
				r := httptest.NewRequest("post", "http://localhost/foo", strings.NewReader(`{"foo": "abc","bar": 12}`))
				r.Header.Set("Content-Type", "application/json")
				rr := httptest.NewRecorder()

				handler.ServeHTTP(rr, r)

				require.Equal(t, 200, rr.Code)
			},
			check: func(t *testing.T, r *engine2.EventRequest) {
				require.Equal(t, &struct {
					Foo string  `json:"foo"`
					Bar float64 `json:"bar"`
				}{Foo: "abc", Bar: 12}, r.Body)
			},
		},
		{
			name: "application/json free-form",
			config: openapitest.NewConfig("3.0.0",
				openapitest.WithServer("http://localhost", ""),
				openapitest.WithPath("/foo", openapitest.NewPath(
					openapitest.WithOperation("post",
						openapitest.NewOperation(
							openapitest.WithRequestBody("", false,
								openapitest.WithRequestContent(
									"application/json", openapitest.NewContent(openapitest.WithSchema(
										schematest.New("object", schematest.WithProperty("foo", schematest.New("string"))),
									)))),
							openapitest.WithResponse(200),
						)),
				)),
			),
			fn: func(t *testing.T, handler http.Handler) {
				r := httptest.NewRequest("post", "http://localhost/foo", strings.NewReader(`{"foo": "abc","bar": 12}`))
				r.Header.Set("Content-Type", "application/json")
				rr := httptest.NewRecorder()

				handler.ServeHTTP(rr, r)

				require.Equal(t, 500, rr.Code)
			},
			check: func(t *testing.T, r *engine2.EventRequest) {
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			test.NewNullLogger()
			events.SetStore(10, events.NewTraits().WithNamespace("http"))
			defer events.Reset()

			var r *engine2.EventRequest
			e := enginetest.NewEngineWithHandler(func(event string, args ...interface{}) []*engine2.Action {
				r = args[0].(*engine2.EventRequest)
				return nil
			})

			tc.fn(t, openapi.NewHandler(tc.config, e))
			tc.check(t, r)
		})

	}
}
