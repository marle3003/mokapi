package service

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/try"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type testHandler struct{}

func (h *testHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	rw.WriteHeader(200)
}

func TestHttpServer_AddOrUpdate(t *testing.T) {
	testdata := []struct {
		name string
		fn   func(t *testing.T, h *HttpServer, port string)
	}{
		{"root path",
			func(t *testing.T, s *HttpServer, port string) {
				err := s.AddOrUpdate(&HttpService{
					Url:     mustParseUrl("http://localhost"),
					Handler: &testHandler{},
					Name:    "foo",
				})
				require.NoError(t, err)
				try.GetRequest(t, fmt.Sprintf("http://localhost:%v", port), map[string]string{}, try.HasStatusCode(200))
			}},
		{"foo path",
			func(t *testing.T, s *HttpServer, port string) {
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
					try.HasBody(fmt.Sprintf("There was no service listening at http://localhost:%v/\n", port)))
				try.GetRequest(t,
					fmt.Sprintf("http://localhost:%v/foo", port),
					nil,
					try.HasStatusCode(200))
			}},
		{"empty url",
			func(t *testing.T, s *HttpServer, port string) {
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
		{"empty host",
			func(t *testing.T, s *HttpServer, port string) {
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
		{"nil handler",
			func(t *testing.T, s *HttpServer, port string) {
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
		{"add on same path",
			func(t *testing.T, s *HttpServer, port string) {
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
	}

	for _, data := range testdata {
		t.Run(data.name, func(t *testing.T) {
			server := NewHttpServer("0", NewServerAliases())
			s := httptest.NewServer(server)
			defer s.Close()

			u, _ := url.Parse(s.URL)

			data.fn(t, server, u.Port())
		})

	}
}

func TestHttpServer_ServerAlias2(t *testing.T) {
	for _, tc := range []struct {
		name                 string
		serverName           string
		aliases              []string
		request              string
		shouldHaveStatusCode int
	}{
		{
			name:                 "mokapi.com with alias localhost",
			serverName:           "mokapi.com",
			aliases:              []string{"mokapi.com=localhost"},
			request:              "http://localhost",
			shouldHaveStatusCode: http.StatusOK,
		},
		{
			name:                 "mokapi.com with alias mokapi.io",
			serverName:           "mokapi.com",
			aliases:              []string{"mokapi.com=mokapi.io"},
			request:              "http://mokapi.io",
			shouldHaveStatusCode: http.StatusOK,
		},
		{
			name:                 "mokapi.com with alias *",
			serverName:           "mokapi.com",
			aliases:              []string{"mokapi.com=*"},
			request:              "http://mokapi.io",
			shouldHaveStatusCode: http.StatusOK,
		},
		{
			name:                 "* with alias *",
			serverName:           "mokapi.com",
			aliases:              []string{"*=*"},
			request:              "http://mokapi.io",
			shouldHaveStatusCode: http.StatusOK,
		},
		{
			name:                 "* with alias mokapi.io",
			serverName:           "mokapi.com",
			aliases:              []string{"*=mokapi.io"},
			request:              "http://mokapi.io",
			shouldHaveStatusCode: http.StatusOK,
		},
		{
			name:                 "alias not matches",
			serverName:           "mokapi.com",
			aliases:              []string{"mokapi.com=mokapi.io"},
			request:              "http://mokapi.net",
			shouldHaveStatusCode: http.StatusNotFound,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			aliases := NewServerAliases()
			err := aliases.Parse(tc.aliases)
			require.NoError(t, err)
			server := NewHttpServer("0", aliases)

			err = server.AddOrUpdate(&HttpService{
				Url:     mustParseUrl("http://" + tc.serverName),
				Handler: &testHandler{},
				Name:    "foo",
			})
			require.NoError(t, err)
			r := httptest.NewRequest(http.MethodGet, tc.request, nil)
			w := httptest.NewRecorder()
			server.ServeHTTP(w, r)
			res := w.Result()
			require.Equal(t, tc.shouldHaveStatusCode, res.StatusCode)
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
