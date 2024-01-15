package http

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
	"io"
	"mokapi/config/dynamic"
	"mokapi/config/static"
	"mokapi/config/tls"
	"mokapi/safe"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestProvider_Start(t *testing.T) {
	testcases := []struct {
		name string
		init func() (static.HttpProvider, *httptest.Server)
		test func(t *testing.T, url string, ch chan *dynamic.Config, hook *test.Hook, err error)
	}{
		{
			name: "invalid url",
			init: func() (static.HttpProvider, *httptest.Server) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

				cfg := static.HttpProvider{
					Url: ":80",
				}

				return cfg, server
			},
			test: func(t *testing.T, url string, ch chan *dynamic.Config, hook *test.Hook, err error) {
				require.NoError(t, err)
				require.Equal(t, "invalid url: parse \":80\": missing protocol scheme", hook.LastEntry().Message)
			},
		},
		{
			name: "invalid poll interval",
			init: func() (static.HttpProvider, *httptest.Server) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

				cfg := static.HttpProvider{
					Url:          server.URL,
					PollInterval: ":8",
				}

				return cfg, server
			},
			test: func(t *testing.T, url string, ch chan *dynamic.Config, hook *test.Hook, err error) {
				require.EqualError(t, err, "unable to parse interval \":8\": time: invalid duration \":8\"")
			},
		},
		{
			name: "not status OK",
			init: func() (static.HttpProvider, *httptest.Server) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(400)
				}))

				cfg := static.HttpProvider{
					Url: server.URL,
				}

				return cfg, server
			},
			test: func(t *testing.T, url string, ch chan *dynamic.Config, hook *test.Hook, err error) {
				require.NoError(t, err)
				time.Sleep(1 * time.Second)
				require.Equal(t, fmt.Sprintf("request to %v failed: received non-ok response code: 400", url), hook.LastEntry().Message)
			},
		},
		{
			name: "update file",
			init: func() (static.HttpProvider, *httptest.Server) {
				content := [][]byte{[]byte("foobar"), []byte("success")}
				i := 0

				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write(content[i])
					i++
				}))

				cfg := static.HttpProvider{
					Url: server.URL,
				}

				return cfg, server
			},
			test: func(t *testing.T, url string, ch chan *dynamic.Config, hook *test.Hook, err error) {
				require.NoError(t, err)
				timeout := time.After(time.Second)
				select {
				case c := <-ch:
					require.Equal(t, url, c.Info.Url.String())
					require.Equal(t, "foobar", string(c.Raw))
				case <-timeout:
					t.Fatal("timeout while waiting for http event")
					return
				}
				timeout = time.After(6 * time.Second)
				select {
				case c := <-ch:
					require.Equal(t, url, c.Info.Url.String())
					require.Equal(t, "success", string(c.Raw))
				case <-timeout:
					t.Fatal("timeout while waiting for http event")
					return
				}
			},
		},
		{
			name: "poll timeout reached",
			init: func() (static.HttpProvider, *httptest.Server) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					time.Sleep(6 * time.Second)
				}))

				cfg := static.HttpProvider{
					Url:          server.URL,
					PollInterval: "30s",
				}

				return cfg, server
			},
			test: func(t *testing.T, url string, ch chan *dynamic.Config, hook *test.Hook, err error) {
				require.NoError(t, err)
				time.Sleep(6 * time.Second)
				require.Equal(t, fmt.Sprintf("request to %v failed: request has timed out", url), hook.LastEntry().Message)
			},
		},
		{
			name: "change poll timeout",
			init: func() (static.HttpProvider, *httptest.Server) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					time.Sleep(2 * time.Second)
				}))

				cfg := static.HttpProvider{
					Url:          server.URL,
					PollTimeout:  "2s",
					PollInterval: "30s",
				}

				return cfg, server
			},
			test: func(t *testing.T, url string, ch chan *dynamic.Config, hook *test.Hook, err error) {
				require.NoError(t, err)
				time.Sleep(3 * time.Second)
				require.Equal(t, fmt.Sprintf("request to %v failed: request has timed out", url), hook.LastEntry().Message)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			logrus.SetOutput(io.Discard)
			hook := test.NewGlobal()

			cfg, server := tc.init()
			defer server.Close()

			pool := safe.NewPool(context.Background())
			defer pool.Stop()
			p := New(cfg)
			ch := make(chan *dynamic.Config)
			err := p.Start(ch, pool)

			tc.test(t, cfg.Url, ch, hook, err)
		})
	}
}

func TestNew(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "invalid root CA",
			test: func(t *testing.T) {
				logrus.SetOutput(io.Discard)
				hook := test.NewGlobal()

				New(static.HttpProvider{Ca: "foo"})

				require.Equal(t, "failed to use CA certification for http provider: failed to parse CA cert: x509: malformed certificate", hook.LastEntry().Message)
			},
		},
		{
			name: "invalid proxy URL",
			test: func(t *testing.T) {
				logrus.SetOutput(io.Discard)
				hook := test.NewGlobal()

				New(static.HttpProvider{Proxy: ":8000"})

				require.Equal(t, "invalid proxy url :8000: parse \":8000\": missing protocol scheme", hook.LastEntry().Message)
			},
		},
		{
			name: "invalid poll timeout",
			test: func(t *testing.T) {
				logrus.SetOutput(io.Discard)
				hook := test.NewGlobal()

				New(static.HttpProvider{PollTimeout: ":3"})

				require.Equal(t, "invalid poll timeout argument ':3', using default: time: invalid duration \":3\"", hook.LastEntry().Message)
				require.Equal(t, logrus.WarnLevel, hook.LastEntry().Level)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t)
		})
	}
}

func TestProxy(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "GET", r.Method, "CONNECT is only used with https")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("foo"))
	}))
	defer server.Close()

	pool := safe.NewPool(context.Background())
	defer pool.Stop()

	ch := make(chan *dynamic.Config)
	p := New(static.HttpProvider{
		Url:   "http://foo.bar",
		Proxy: server.URL,
	})
	err := p.Start(ch, pool)
	require.NoError(t, err)

	timeout := time.After(500 * time.Millisecond)
	var configs []*dynamic.Config
Loop:
	for {
		select {
		case c := <-ch:
			configs = append(configs, c)
		case <-timeout:
			break Loop
		}
	}

	require.Equal(t, []byte("foo"), configs[0].Raw)
}

func TestTlsWithCA(t *testing.T) {
	t.Parallel()

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("foo"))
	}))
	defer server.Close()

	config := static.HttpProvider{
		Url: server.URL,
		Ca:  tls.FileOrContent(server.TLS.Certificates[0].Certificate[0]),
	}

	pool := safe.NewPool(context.Background())
	defer pool.Stop()

	ch := make(chan *dynamic.Config)
	p := New(config)
	err := p.Start(ch, pool)
	require.NoError(t, err)

	timeout := time.After(500 * time.Millisecond)
	var configs []*dynamic.Config
Loop:
	for {
		select {
		case c := <-ch:
			configs = append(configs, c)
		case <-timeout:
			break Loop
		}
	}

	require.Equal(t, []byte("foo"), configs[0].Raw)
}

func TestTlsWithSkipCertVerification(t *testing.T) {
	t.Parallel()

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("foo"))
	}))
	defer server.Close()

	config := static.HttpProvider{
		Url:           server.URL,
		TlsSkipVerify: true,
	}

	pool := safe.NewPool(context.Background())
	defer pool.Stop()

	ch := make(chan *dynamic.Config)
	p := New(config)
	err := p.Start(ch, pool)
	require.NoError(t, err)

	timeout := time.After(500 * time.Millisecond)
	var configs []*dynamic.Config
Loop:
	for {
		select {
		case c := <-ch:
			configs = append(configs, c)
		case <-timeout:
			break Loop
		}
	}

	require.Equal(t, []byte("foo"), configs[0].Raw)
}

func TestTlsWithCertError(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("foo"))
	}))
	defer server.Close()

	config := static.HttpProvider{
		Url: server.URL,
	}

	pool := safe.NewPool(context.Background())
	defer pool.Stop()

	hook := test.NewGlobal()

	ch := make(chan *dynamic.Config)
	p := New(config)
	err := p.Start(ch, pool)
	require.NoError(t, err)

	timeout := time.After(500 * time.Millisecond)
	var configs []*dynamic.Config
Loop:
	for {
		select {
		case c := <-ch:
			configs = append(configs, c)
		case <-timeout:
			break Loop
		}
	}

	require.Len(t, configs, 0)
	require.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
	require.Equal(t,
		fmt.Sprintf(`request to https://%v failed: Get "https://%v": tls: failed to verify certificate: x509: certificate signed by unknown authority`,
			server.Listener.Addr(),
			server.Listener.Addr()),
		hook.LastEntry().Message)
}
