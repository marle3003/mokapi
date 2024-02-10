package service

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/providers/openapi"
	"mokapi/runtime/events"
	"mokapi/try"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type testHandler struct {
	f func(rw http.ResponseWriter, r *http.Request)
}

func (h *testHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if h.f != nil {
		h.f(rw, r)
	} else {
		rw.WriteHeader(200)
	}
}

func TestHttpServer_AddOrUpdate(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, h *HttpServer, port string)
	}{
		{
			name: "add service on root",
			test: func(t *testing.T, s *HttpServer, port string) {
				err := s.AddOrUpdate(&HttpService{
					Url:     mustParseUrl("http://localhost"),
					Handler: &testHandler{},
					Name:    "foo",
				})
				require.NoError(t, err)
				try.GetRequest(t, fmt.Sprintf("http://localhost:%v", port), map[string]string{}, try.HasStatusCode(200))
			}},
		{
			name: "add service on path foo",
			test: func(t *testing.T, s *HttpServer, port string) {
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

				list := events.GetEvents(events.NewTraits().WithNamespace("http"))
				require.Len(t, list, 1, "only 404 should be logged")
				data := list[0].Data.(*openapi.HttpLog)
				require.Len(t, data.Response.Headers, 1)
				require.Equal(t, "text/plain; charset=utf-8", data.Response.Headers["Content-Type"])
			}},
		{
			name: "add service with empty url",
			test: func(t *testing.T, s *HttpServer, port string) {
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
			}},
		{
			name: "add service with empty host",
			test: func(t *testing.T, s *HttpServer, port string) {
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
			}},
		{
			name: "nil handler",
			test: func(t *testing.T, s *HttpServer, port string) {
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
					try.HasBody("handler is nil\n"))
			}},
		{
			name: "add service on already used path",
			test: func(t *testing.T, s *HttpServer, port string) {
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
				require.Error(t, err, "service 'foo' is already defined on path ''")
			}},
		{
			name: "update service",
			test: func(t *testing.T, s *HttpServer, port string) {
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
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			events.SetStore(20, events.NewTraits().WithNamespace("http"))
			defer events.Reset()
			server := NewHttpServer("0")
			s := httptest.NewServer(server)
			defer s.Close()

			u, _ := url.Parse(s.URL)

			tc.test(t, server, u.Port())
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
