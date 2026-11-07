package service_test

import (
	"bytes"
	"fmt"
	"io"
	"mokapi/config/static"
	"mokapi/lib"
	"mokapi/providers/openapi"
	"mokapi/runtime/events"
	"mokapi/server/cert"
	"mokapi/server/service"
	"mokapi/try"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
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
	logrus.SetOutput(io.Discard)

	testcases := []struct {
		name string
		test func(t *testing.T, h *service.HttpServer, port string, sm *events.StoreManager)
	}{
		{
			name: "add service on root",
			test: func(t *testing.T, s *service.HttpServer, port string, sm *events.StoreManager) {
				err := s.AddOrUpdate(&service.HttpService{
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
			test: func(t *testing.T, s *service.HttpServer, port string, sm *events.StoreManager) {
				err := s.AddOrUpdate(&service.HttpService{
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
			test: func(t *testing.T, s *service.HttpServer, port string, sm *events.StoreManager) {
				err := s.AddOrUpdate(&service.HttpService{
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
			test: func(t *testing.T, s *service.HttpServer, port string, sm *events.StoreManager) {
				err := s.AddOrUpdate(&service.HttpService{
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
			test: func(t *testing.T, s *service.HttpServer, port string, sm *events.StoreManager) {
				err := s.AddOrUpdate(&service.HttpService{
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
			test: func(t *testing.T, s *service.HttpServer, port string, sm *events.StoreManager) {
				err := s.AddOrUpdate(&service.HttpService{
					Url:     mustParseUrl(""),
					Handler: nil,
					Name:    "foo",
				})
				require.NoError(t, err)

				err = s.AddOrUpdate(&service.HttpService{
					Url:     mustParseUrl(""),
					Handler: nil,
					Name:    "bar",
				})
				require.NoError(t, err)
			},
		},
		{
			name: "update service",
			test: func(t *testing.T, s *service.HttpServer, port string, sm *events.StoreManager) {
				err := s.AddOrUpdate(&service.HttpService{
					Url:     mustParseUrl(""),
					Handler: nil,
					Name:    "foo",
				})
				require.NoError(t, err)

				err = s.AddOrUpdate(&service.HttpService{
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
			test: func(t *testing.T, s *service.HttpServer, port string, sm *events.StoreManager) {
				err := s.AddOrUpdate(&service.HttpService{
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

				err = s.AddOrUpdate(&service.HttpService{
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
			test: func(t *testing.T, s *service.HttpServer, port string, sm *events.StoreManager) {
				err := s.AddOrUpdate(&service.HttpService{
					Url: mustParseUrl(""),
					Handler: &testHandler{f: func(rw http.ResponseWriter, r *http.Request) *openapi.HttpError {
						return &openapi.HttpError{
							StatusCode: http.StatusBadRequest,
							Header:     map[string][]string{"Foo": {"bar"}},
							Message:    "Test Error",
							Traits:     events.NewTraits().WithNamespace("http").WithName("API").With("path", "/foo"),
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

				evts := sm.GetEvents(events.NewTraits())
				require.Len(t, evts, 1)
				require.Equal(t, "namespace=http, name=API, path=/foo", evts[0].Traits.String())
				require.IsType(t, &openapi.HttpLog{}, evts[0].Data)
				d := evts[0].Data.(*openapi.HttpLog)
				require.Equal(t, "API", d.Api)
				require.Equal(t, "/foo", d.Path)
				require.Equal(t, map[string]string{"Content-Type": "text/plain; charset=utf-8", "Foo": "bar"}, d.Response.Headers)
				require.Equal(t, "Test Error", d.Response.Body)
			},
		},
		{
			name: "add service different base path",
			test: func(t *testing.T, s *service.HttpServer, port string, sm *events.StoreManager) {
				err := s.AddOrUpdate(&service.HttpService{
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

				err = s.AddOrUpdate(&service.HttpService{
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
			test: func(t *testing.T, s *service.HttpServer, port string, sm *events.StoreManager) {
				err := s.AddOrUpdate(&service.HttpService{
					Url:     mustParseUrl("/foo"),
					Handler: &testHandler{},
					Name:    "foo",
				})
				require.NoError(t, err)

				err = s.AddOrUpdate(&service.HttpService{
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
			test: func(t *testing.T, s *service.HttpServer, port string, sm *events.StoreManager) {
				err := s.AddOrUpdate(&service.HttpService{
					Url:     mustParseUrl("/foo"),
					Handler: &testHandler{},
					Name:    "foo",
				})
				require.NoError(t, err)

				err = s.AddOrUpdate(&service.HttpService{
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
		{
			name: "send request body to undefined endpoint, but the body should be read",
			test: func(t *testing.T, s *service.HttpServer, port string, sm *events.StoreManager) {
				err := s.AddOrUpdate(&service.HttpService{
					Url:     mustParseUrl("/foo"),
					Handler: &testHandler{},
					Name:    "foo",
				})
				require.NoError(t, err)

				spy := &spyBody{Reader: bytes.NewBufferString(`{"foo": "abc","bar": 12}`)}

				try.Request(t,
					http.MethodPost,
					fmt.Sprintf("http://localhost:%v/bar", port),
					nil,
					spy,
					try.HasStatusCode(http.StatusNotFound),
					try.HasBody(fmt.Sprintf("There was no service listening at http://localhost:%s/bar\n", port)),
				)

				require.True(t, spy.readCalled, "server needs to read body")
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			sm := &events.StoreManager{}
			sm.SetStore(20, events.NewTraits().WithNamespace("http"))

			port := fmt.Sprintf("%v", try.GetFreePort())
			s := service.NewHttpServer(port, sm)
			require.NoError(t, s.Start())
			defer s.Stop()

			tc.test(t, s, port, sm)
		})

	}
}

func TestHttp(t *testing.T) {
	logrus.SetOutput(io.Discard)
	hook := test.NewGlobal()

	sm := &events.StoreManager{}
	sm.SetStore(20, events.NewTraits().WithNamespace("http"))

	port := fmt.Sprintf("%v", try.GetFreePort())
	s := service.NewHttpServer(port, sm)
	require.NoError(t, s.Start())
	defer s.Stop()

	err := s.AddOrUpdate(&service.HttpService{
		Url:     mustParseUrl("https://localhost"),
		Handler: &testHandler{},
		Name:    "foo",
	})
	require.NoError(t, err)
	require.Len(t, hook.Entries, 2)
	require.Equal(t, fmt.Sprintf("adding new HTTP host 'localhost' on binding :%s", port), hook.Entries[0].Message)
	require.Equal(t, fmt.Sprintf("adding service 'foo' on binding :%s on path /", port), hook.Entries[1].Message)

	try.GetRequest(t, fmt.Sprintf("http://localhost:%v", port), map[string]string{}, try.HasStatusCode(200))

	require.Len(t, hook.Entries, 3)
	require.Equal(t, fmt.Sprintf("processing HTTP request GET http://localhost:%s/", port), hook.Entries[2].Message)
}

func TestHttpTls(t *testing.T) {
	logrus.SetOutput(io.Discard)
	hook := test.NewGlobal()

	sm := &events.StoreManager{}
	sm.SetStore(20, events.NewTraits().WithNamespace("http"))

	certStore, err := cert.NewStore(&static.Config{})
	require.NoError(t, err)

	port := fmt.Sprintf("%v", try.GetFreePort())
	s := service.NewHttpServerTls(port, certStore, sm)
	require.NoError(t, s.Start())
	defer s.Stop()

	err = s.AddOrUpdate(&service.HttpService{
		Url:     mustParseUrl("https://localhost"),
		Handler: &testHandler{},
		Name:    "foo",
	})
	require.NoError(t, err)
	require.Len(t, hook.Entries, 2)
	require.Equal(t, fmt.Sprintf("adding new HTTPS host 'localhost' on binding :%s", port), hook.Entries[0].Message)
	require.Equal(t, fmt.Sprintf("adding service 'foo' on binding :%s on path /", port), hook.Entries[1].Message)

	try.GetRequest(t, fmt.Sprintf("https://localhost:%v", port), map[string]string{}, try.HasStatusCode(200))

	require.Len(t, hook.Entries, 3)
	require.Equal(t, fmt.Sprintf("processing HTTPS request GET https://localhost:%s/", port), hook.Entries[2].Message)
}

// TestHttpServer_ConcurrentAccess mimics parallel config updates while requests arrive: the
// server must be safe for adding and removing services, asking whether it is empty and
// dispatching requests at the same time.
func TestHttpServer_ConcurrentAccess(t *testing.T) {
	logrus.SetOutput(io.Discard)

	sm := &events.StoreManager{}
	sm.SetStore(20, events.NewTraits().WithNamespace("http"))
	s := service.NewHttpServer(fmt.Sprintf("%v", try.GetFreePort()), sm)

	u := mustParseUrl("http://localhost/foo")
	const iterations = 200
	operations := []func(){
		func() {
			err := s.AddOrUpdate(&service.HttpService{Url: u, Handler: &testHandler{}, Name: "foo"})
			assert.NoError(t, err)
		},
		func() { s.RemoveUrl(u) },
		func() { s.Remove("foo") },
		func() { s.CanClose() },
		func() {
			req := httptest.NewRequest(http.MethodGet, "http://localhost/foo", nil)
			s.ServeHTTP(httptest.NewRecorder(), req)
		},
	}

	var wg sync.WaitGroup
	for _, op := range operations {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				op()
			}
		}()
	}
	wg.Wait()
}

// TestHttpServer_HandlerCanChangeServices verifies that a handler (e.g. a script reacting to a
// request) may add or remove services without blocking, because the server must not hold its
// lock while a handler runs.
func TestHttpServer_HandlerCanChangeServices(t *testing.T) {
	logrus.SetOutput(io.Discard)

	sm := &events.StoreManager{}
	sm.SetStore(20, events.NewTraits().WithNamespace("http"))
	s := service.NewHttpServer(fmt.Sprintf("%v", try.GetFreePort()), sm)

	u := mustParseUrl("http://localhost/foo")
	err := s.AddOrUpdate(&service.HttpService{
		Url: u,
		Handler: &testHandler{f: func(rw http.ResponseWriter, r *http.Request) *openapi.HttpError {
			s.RemoveUrl(u)
			rw.WriteHeader(http.StatusOK)
			return nil
		}},
		Name: "foo",
	})
	require.NoError(t, err)

	done := make(chan struct{})
	rec := httptest.NewRecorder()
	go func() {
		defer close(done)
		s.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "http://localhost/foo", nil))
	}()

	select {
	case <-done:
		require.Equal(t, http.StatusOK, rec.Code)
		require.True(t, s.CanClose())
	case <-time.After(5 * time.Second):
		t.Fatal("handler blocked: server holds its lock while calling the handler")
	}
}

func mustParseUrl(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
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

// Start must return only when the port accepts connections. Dials once without retry.
func TestHttpServer_Start_PortIsBoundWhenStartReturns(t *testing.T) {
	logrus.SetOutput(io.Discard)

	for i := 0; i < 200; i++ {
		port, s := startOnFreePort(t)

		conn, err := net.Dial("tcp", "localhost:"+port)
		s.Stop()
		require.NoError(t, err, "iteration %d", i)
		_ = conn.Close()
	}
}

func TestHttpServer_Start_ReturnsErrorWhenPortIsInUse(t *testing.T) {
	ln, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	defer ln.Close()
	port := fmt.Sprintf("%v", ln.Addr().(*net.TCPAddr).Port)

	t.Run("http", func(t *testing.T) {
		s := service.NewHttpServer(port, &events.StoreManager{})
		require.Error(t, s.Start())
		s.Stop() // must not panic after a failed start
	})

	t.Run("https", func(t *testing.T) {
		certStore, err := cert.NewStore(&static.Config{})
		require.NoError(t, err)
		s := service.NewHttpServerTls(port, certStore, &events.StoreManager{})
		require.Error(t, s.Start())
		s.Stop()
	})
}

func TestHttpServer_Stop_RightAfterStart(t *testing.T) {
	for i := 0; i < 50; i++ {
		_, s := startOnFreePort(t)
		s.Stop()
	}
}

// startOnFreePort retries because another process may take the port between GetFreePort and Start.
func startOnFreePort(t *testing.T) (string, *service.HttpServer) {
	var err error
	for attempt := 0; attempt < 5; attempt++ {
		port := fmt.Sprintf("%v", try.GetFreePort())
		s := service.NewHttpServer(port, &events.StoreManager{})
		if err = s.Start(); err == nil {
			return port, s
		}
	}
	require.NoError(t, err)
	return "", nil
}
