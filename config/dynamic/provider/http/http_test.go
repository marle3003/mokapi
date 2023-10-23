package http

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/common"
	"mokapi/config/static"
	"mokapi/config/tls"
	"mokapi/safe"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

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

	ch := make(chan *common.Config)
	p := New(static.HttpProvider{
		Url:   "http://foo.bar",
		Proxy: server.URL,
	})
	err := p.Start(ch, pool)
	require.NoError(t, err)

	timeout := time.After(500 * time.Millisecond)
	var configs []*common.Config
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

func TestProvider(t *testing.T) {
	t.Parallel()

	content := [][]byte{[]byte("foobar"), []byte("success")}
	i := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(content[i])
	}))
	defer server.Close()

	pool := safe.NewPool(context.Background())
	defer pool.Stop()
	p := New(static.HttpProvider{
		Url: server.URL,
	})
	ch := make(chan *common.Config)
	err := p.Start(ch, pool)
	require.NoError(t, err)

	timeout := time.After(time.Second)
	select {
	case c := <-ch:
		require.Equal(t, p.config.Url, c.Info.Url.String())
		require.Equal(t, content[0], c.Raw)
	case <-timeout:
		t.Fatal("timeout while waiting for http event")
		return
	}
	i++
	timeout = time.After(6 * time.Second)
	select {
	case c := <-ch:
		require.Equal(t, p.config.Url, c.Info.Url.String())
		require.Equal(t, content[1], c.Raw)
	case <-timeout:
		t.Fatal("timeout while waiting for http event")
		return
	}
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

	ch := make(chan *common.Config)
	p := New(config)
	err := p.Start(ch, pool)
	require.NoError(t, err)

	timeout := time.After(500 * time.Millisecond)
	var configs []*common.Config
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

	ch := make(chan *common.Config)
	p := New(config)
	err := p.Start(ch, pool)
	require.NoError(t, err)

	timeout := time.After(500 * time.Millisecond)
	var configs []*common.Config
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

	ch := make(chan *common.Config)
	p := New(config)
	err := p.Start(ch, pool)
	require.NoError(t, err)

	timeout := time.After(500 * time.Millisecond)
	var configs []*common.Config
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
		fmt.Sprintf(`request to "https://%v" failed: Get "https://%v": tls: failed to verify certificate: x509: certificate signed by unknown authority`,
			server.Listener.Addr(),
			server.Listener.Addr()),
		hook.LastEntry().Message)
}
