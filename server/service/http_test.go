package service

import (
	"fmt"
	"mokapi/lib"
	"mokapi/providers/openapi"
	"mokapi/runtime/events"
	"mokapi/try"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type testHandler struct {
	f func(rw http.ResponseWriter, r *http.Request) *openapi.HttpError
}

func (h *testHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) *openapi.HttpError {
	if h.f != nil {
		return h.f(rw, r)
	} else {
		rw.WriteHeader(200)
		return nil
	}
}

func TestHttpServer_AddOrUpdate(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, h *HttpServer, port string, sm *events.StoreManager)
	}{
		{
			name: "add service on root",
			test: func(t *testing.T, s *HttpServer, port string, sm *events.StoreManager) {
				err := s.AddOrUpdate(&HttpService{
					Url:     mustParseUrl("http://localhost"),
					Handler: &testHandler{},
					Name:    "foo",
				})
				require.NoError(t, err)
				try.GetRequest(t, fmt.Sprintf("http://localhost:%v", port), map[string]string{}, try.HasStatusCode(200))
			},
		},
		{
			name: "add service on path foo",
			test: func(t *testing.T, s *HttpServer, port string, sm *events.StoreManager) {
				err := s.AddOrUpdate(&HttpService{
					Url:     mustParseUrl("http://localhost/foo"),
					Handler: &testHandler{},
					Name:    "foo",
				})
				require.NoError(t, err)
				try.GetRequest(t,
					fmt.Sprintf("http://localhost:%v", port),
					nil,
					try.HasStatusCode(404),
					try.HasHeader("Content-Type", "text/plain; charset=utf-8"),
					try.HasBody(fmt.Sprintf("There was no service listening at http://localhost:%v/\n", port)))
				try.GetRequest(t,
					fmt.Sprintf("http://localhost:%v/foo", port),
					nil,
					try.HasStatusCode(200))

				list := sm.GetEvents(events.NewTraits().WithNamespace("http"))
				require.Len(t, list, 1, "only 404 should be logged")
				data := list[0].Data.(*openapi.HttpLog)
				require.Len(t, data.Response.Headers, 1)
				require.Equal(t, "text/plain; charset=utf-8", data.Response.Headers["Content-Type"])
			},
		},
		{
			name: "add service with empty url",
			test: func(t *testing.T, s *HttpServer, port string, sm *events.StoreManager) {
				err := s.AddOrUpdate(&HttpService{
					Url:     mustParseUrl(""),
					Handler: &testHandler{},
					Name:    "foo",
				})

				require.NoError(t, err)

				// simulate request because lookup foo would not work: dial tcp: lookup foo: no such host
				r := httptest.NewRequest("GET", fmt.Sprintf("http://foo:%v/bar", port), nil)
				rr := httptest.NewRecorder()
				s.ServeHTTP(rr, r)
				require.Equal(t, 200, rr.Code)
			},
		},
		{
			name: "add service with empty host",
			test: func(t *testing.T, s *HttpServer, port string, sm *events.StoreManager) {
				err := s.AddOrUpdate(&HttpService{
					Url:     mustParseUrl("/foo"),
					Handler: &testHandler{},
					Name:    "foo",
				})

				require.NoError(t, err)
				r := httptest.NewRequest("GET", fmt.Sprintf("http://foo:%v/foo", port), nil)
				rr := httptest.NewRecorder()
				s.ServeHTTP(rr, r)
				require.Equal(t, 200, rr.Code)
			},
		},
		{
			name: "nil handler",
			test: func(t *testing.T, s *HttpServer, port string, sm *events.StoreManager) {
				err := s.AddOrUpdate(&HttpService{
					Url:     mustParseUrl(""),
					Handler: nil,
					Name:    "foo",
				})

				require.NoError(t, err)
				require.NoError(t, err)
				try.GetRequest(t,
					fmt.Sprintf("http://localhost:%v", port),
					map[string]string{}, try.HasStatusCode(500),
					try.HasBody(fmt.Sprintf("Handler is nil for http://localhost:%s/\n", port)),
				)
			},
		},
		{
			name: "add service on already used path",
			test: func(t *testing.T, s *HttpServer, port string, sm *events.StoreManager) {
				err := s.AddOrUpdate(&HttpService{
					Url:     mustParseUrl(""),
					Handler: nil,
					Name:    "foo",
				})
				require.NoError(t, err)

				err = s.AddOrUpdate(&HttpService{
					Url:     mustParseUrl(""),
					Handler: nil,
					Name:    "bar",
				})
				require.NoError(t, err)
			},
		},
		{
			name: "update service",
			test: func(t *testing.T, s *HttpServer, port string, sm *events.StoreManager) {
				err := s.AddOrUpdate(&HttpService{
					Url:     mustParseUrl(""),
					Handler: nil,
					Name:    "foo",
				})
				require.NoError(t, err)

				err = s.AddOrUpdate(&HttpService{
					Url:     mustParseUrl(""),
					Handler: &testHandler{},
					Name:    "foo",
				})
				require.NoError(t, err)
				r := httptest.NewRequest("GET", fmt.Sprintf("http://foo:%v/foo", port), nil)
				rr := httptest.NewRecorder()
				s.ServeHTTP(rr, r)
				require.Equal(t, 200, rr.Code)
			},
		},
		{
			name: "add service on same base path with different path",
			test: func(t *testing.T, s *HttpServer, port string, sm *events.StoreManager) {
				err := s.AddOrUpdate(&HttpService{
					Url: mustParseUrl(""),
					Handler: &testHandler{f: func(rw http.ResponseWriter, r *http.Request) *openapi.HttpError {
						if r.URL.Path == "/foo" {
							rw.WriteHeader(http.StatusOK)
							_, _ = rw.Write([]byte("foo"))
							return nil
						}
						return &openapi.HttpError{StatusCode: http.StatusNotFound, Message: fmt.Sprintf("no matching endpoint found: %v %v", strings.ToUpper(r.Method), lib.GetUrl(r))}
					}},
					Name: "foo",
				})
				require.NoError(t, err)

				err = s.AddOrUpdate(&HttpService{
					Url: mustParseUrl(""),
					Handler: &testHandler{f: func(rw http.ResponseWriter, r *http.Request) *openapi.HttpError {
						if r.URL.Path == "/bar" {
							rw.WriteHeader(http.StatusOK)
							_, _ = rw.Write([]byte("bar"))
							return nil
						}
						return &openapi.HttpError{StatusCode: http.StatusNotFound, Message: fmt.Sprintf("no matching endpoint found: %v %v", strings.ToUpper(r.Method), lib.GetUrl(r))}
					}},
					Name: "bar",
				})
				require.NoError(t, err)

				try.GetRequest(t,
					fmt.Sprintf("http://localhost:%v/foo", port),
					nil,
					try.HasStatusCode(http.StatusOK),
					try.HasBody("foo"),
				)
				try.GetRequest(t,
					fmt.Sprintf("http://localhost:%v/bar", port),
					nil,
					try.HasStatusCode(http.StatusOK),
					try.HasBody("bar"),
				)
				try.GetRequest(t,
					fmt.Sprintf("http://localhost:%v/yuh", port),
					nil,
					try.HasStatusCode(http.StatusNotFound),
					try.HasBody(fmt.Sprintf("No matching endpoint found: GET http://localhost:%s/yuh\n", port)),
				)
			},
		},
		{
			name: "http error with header",
			test: func(t *testing.T, s *HttpServer, port string, sm *events.StoreManager) {
				err := s.AddOrUpdate(&HttpService{
					Url: mustParseUrl(""),
					Handler: &testHandler{f: func(rw http.ResponseWriter, r *http.Request) *openapi.HttpError {
						return &openapi.HttpError{
							StatusCode: http.StatusBadRequest,
							Header:     map[string][]string{"Foo": {"bar"}},
						}
					}},
					Name: "foo",
				})
				require.NoError(t, err)

				try.GetRequest(t,
					fmt.Sprintf("http://localhost:%v/foo", port),
					nil,
					try.HasStatusCode(http.StatusBadRequest),
					try.HasHeader("Foo", "bar"),
				)
			},
		},
		{
			name: "add service different base path",
			test: func(t *testing.T, s *HttpServer, port string, sm *events.StoreManager) {
				err := s.AddOrUpdate(&HttpService{
					Url: mustParseUrl("/v1"),
					Handler: &testHandler{f: func(rw http.ResponseWriter, r *http.Request) *openapi.HttpError {
						servicePath, _ := r.Context().Value("servicePath").(string)
						require.Equal(t, "/v1", servicePath)

						if r.URL.Path == "/v1/foo" {
							rw.WriteHeader(http.StatusOK)
							_, _ = rw.Write([]byte("foo"))
							return nil
						}
						return &openapi.HttpError{StatusCode: http.StatusNotFound, Message: fmt.Sprintf("no matching endpoint found: %v %v", strings.ToUpper(r.Method), lib.GetUrl(r))}
					}},
					Name: "foo",
				})
				require.NoError(t, err)

				err = s.AddOrUpdate(&HttpService{
					Url: mustParseUrl("/v2"),
					Handler: &testHandler{f: func(rw http.ResponseWriter, r *http.Request) *openapi.HttpError {
						servicePath, _ := r.Context().Value("servicePath").(string)
						require.Equal(t, "/v2", servicePath)

						if r.URL.Path == "/v2/bar" {
							rw.WriteHeader(http.StatusOK)
							_, _ = rw.Write([]byte("bar"))
							return nil
						}
						return &openapi.HttpError{StatusCode: http.StatusNotFound, Message: fmt.Sprintf("no matching endpoint found: %v %v", strings.ToUpper(r.Method), lib.GetUrl(r))}
					}},
					Name: "bar",
				})
				require.NoError(t, err)

				try.GetRequest(t,
					fmt.Sprintf("http://localhost:%v/v1/foo", port),
					nil,
					try.HasStatusCode(http.StatusOK),
					try.HasBody("foo"),
				)
				try.GetRequest(t,
					fmt.Sprintf("http://localhost:%v/v2/bar", port),
					nil,
					try.HasStatusCode(http.StatusOK),
					try.HasBody("bar"),
				)
				try.GetRequest(t,
					fmt.Sprintf("http://localhost:%v/v1/yuh", port),
					nil,
					try.HasStatusCode(http.StatusNotFound),
					try.HasBody(fmt.Sprintf("No matching endpoint found: GET http://localhost:%s/v1/yuh\n", port)),
				)
			},
		},
		{
			name: "remove one URL",
			test: func(t *testing.T, s *HttpServer, port string, sm *events.StoreManager) {
				err := s.AddOrUpdate(&HttpService{
					Url:     mustParseUrl("/foo"),
					Handler: &testHandler{},
					Name:    "foo",
				})
				require.NoError(t, err)

				err = s.AddOrUpdate(&HttpService{
					Url:     mustParseUrl("/bar"),
					Handler: &testHandler{},
					Name:    "foo",
				})
				require.NoError(t, err)

				s.RemoveUrl(mustParseUrl("/foo"))

				try.GetRequest(t,
					fmt.Sprintf("http://localhost:%v/foo", port),
					nil,
					try.HasStatusCode(http.StatusNotFound),
					try.HasBody(fmt.Sprintf("There was no service listening at http://localhost:%s/foo\n", port)),
				)

				try.GetRequest(t,
					fmt.Sprintf("http://localhost:%v/bar", port),
					nil,
					try.HasStatusCode(http.StatusOK),
				)

				s.RemoveUrl(mustParseUrl("/bar"))

				try.GetRequest(t,
					fmt.Sprintf("http://localhost:%v/bar", port),
					nil,
					try.HasStatusCode(http.StatusNotFound),
					try.HasBody(fmt.Sprintf("There was no service listening at http://localhost:%s/bar\n", port)),
				)
			},
		},
		{
			name: "remove service",
			test: func(t *testing.T, s *HttpServer, port string, sm *events.StoreManager) {
				err := s.AddOrUpdate(&HttpService{
					Url:     mustParseUrl("/foo"),
					Handler: &testHandler{},
					Name:    "foo",
				})
				require.NoError(t, err)

				err = s.AddOrUpdate(&HttpService{
					Url:     mustParseUrl("/bar"),
					Handler: &testHandler{},
					Name:    "foo",
				})
				require.NoError(t, err)

				s.Remove("foo")

				try.GetRequest(t,
					fmt.Sprintf("http://localhost:%v/foo", port),
					nil,
					try.HasStatusCode(http.StatusNotFound),
					try.HasBody(fmt.Sprintf("There was no service listening at http://localhost:%s/foo\n", port)),
				)

				try.GetRequest(t,
					fmt.Sprintf("http://localhost:%v/bar", port),
					nil,
					try.HasStatusCode(http.StatusNotFound),
					try.HasBody(fmt.Sprintf("There was no service listening at http://localhost:%s/bar\n", port)),
				)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			sm := &events.StoreManager{}
			sm.SetStore(20, events.NewTraits().WithNamespace("http"))

			server := NewHttpServer("0", sm)
			s := httptest.NewServer(server)
			defer s.Close()

			u, _ := url.Parse(s.URL)

			tc.test(t, server, u.Port(), sm)
		})

	}
}

func mustParseUrl(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}
