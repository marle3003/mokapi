package openapi_test

import (
	"bytes"
	"io"
	engine2 "mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	"mokapi/providers/openapi/schema/schematest"
	"mokapi/runtime/events"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/blevesearch/bleve/v2"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestResponseHandler_ServeHTTP_ResponseBody(t *testing.T) {
	testcases := []struct {
		name   string
		config *openapi.Config
		fn     func(t *testing.T, handler openapi.Handler)
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
			fn: func(t *testing.T, handler openapi.Handler) {
				r := httptest.NewRequest("post", "http://localhost/foo", strings.NewReader("foo"))
				r.Header.Set("Content-Type", "text/plain")
				rr := httptest.NewRecorder()

				err := handler.ServeHTTP(rr, r)
				require.Nil(t, err)

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
			fn: func(t *testing.T, handler openapi.Handler) {
				r := httptest.NewRequest("post", "http://localhost/foo", strings.NewReader("foo"))
				r.Header.Set("Content-Type", "text/plain")
				rr := httptest.NewRecorder()

				err := handler.ServeHTTP(rr, r)
				require.Nil(t, err)

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
			fn: func(t *testing.T, handler openapi.Handler) {
				r := httptest.NewRequest("post", "http://localhost/foo", strings.NewReader("foo"))
				r.Header.Set("Content-Type", "text/plain")
				rr := httptest.NewRecorder()

				err := handler.ServeHTTP(rr, r)
				require.Nil(t, err)

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
			fn: func(t *testing.T, handler openapi.Handler) {
				r := httptest.NewRequest("post", "http://localhost/foo", strings.NewReader(`{"foo": "abc","bar": 12}`))
				r.Header.Set("Content-Type", "application/json")
				rr := httptest.NewRecorder()

				err := handler.ServeHTTP(rr, r)
				require.Nil(t, err)

				require.Equal(t, 200, rr.Code)
			},
			check: func(t *testing.T, r *engine2.EventRequest) {
				require.Equal(t, map[string]interface{}{"bar": float64(12), "foo": "abc"}, r.Body)
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
			fn: func(t *testing.T, handler openapi.Handler) {
				r := httptest.NewRequest("post", "http://localhost/foo", strings.NewReader(`{"foo": "abc","bar": 12}`))
				r.Header.Set("Content-Type", "application/json")
				rr := httptest.NewRecorder()

				err := handler.ServeHTTP(rr, r)
				require.Nil(t, err)

				require.Equal(t, http.StatusOK, rr.Code)
			},
			check: func(t *testing.T, r *engine2.EventRequest) {
			},
		},
		{
			name: "no request body specified but request contains body",
			config: openapitest.NewConfig("3.0.0",
				openapitest.WithServer("http://localhost", ""),
				openapitest.WithPath("/foo", openapitest.NewPath(
					openapitest.WithOperation("post",
						openapitest.NewOperation(
							openapitest.WithResponse(200),
						)),
				)),
			),
			fn: func(t *testing.T, handler openapi.Handler) {
				spy := &spyBody{Reader: bytes.NewBufferString(`{"foo": "abc","bar": 12}`)}

				r := httptest.NewRequest("post", "http://localhost/foo", spy)
				r.Header.Set("Content-Type", "application/json")
				rr := httptest.NewRecorder()

				err := handler.ServeHTTP(rr, r)
				require.Nil(t, err)

				require.Equal(t, http.StatusOK, rr.Code)
				require.True(t, spy.readCalled, "server needs to read body")
			},
			check: func(t *testing.T, r *engine2.EventRequest) {
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			test.NewNullLogger()

			var r *engine2.EventRequest
			e := enginetest.NewEngineWithHandler(func(event string, args ...interface{}) []*engine2.Action {
				r = args[0].(*engine2.EventRequest)
				return nil
			})

			idx, err := bleve.NewMemOnly(bleve.NewIndexMapping())
			require.NoError(t, err)
			store := events.NewStoreManager(idx)
			store.SetStore(10, events.NewTraits().WithNamespace("http"))

			tc.fn(t, openapi.NewHandler(tc.config, e, store))
			tc.check(t, r)
		})

	}
}

type spyBody struct {
	io.Reader
	readCalled bool
}

func (s *spyBody) Read(p []byte) (n int, err error) {
	s.readCalled = true
	return s.Reader.Read(p)
}

func (s *spyBody) Close() error {
	return nil
}
