package http

import (
	"context"
	"io"
	"mokapi/config/dynamic"
	"mokapi/config/static"
	"mokapi/safe"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func requireDone(t *testing.T, f func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		f()
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("operation still blocked after the context was canceled")
	}
}

func TestCheckFiles_StopsWhenContextIsDone(t *testing.T) {
	logrus.SetOutput(io.Discard)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("foo: bar"))
	}))
	defer server.Close()

	p := New(static.HttpProvider{Urls: []string{server.URL}})
	p.files[server.URL] = 0
	ctx, cancel := context.WithCancel(context.Background())

	// nobody reads the channel, the changed file can only leave checkFiles through the context
	cancel()

	requireDone(t, func() { p.checkFiles(ctx, make(chan dynamic.ConfigEvent)) })
}

func TestProvider_StopDuringPoll(t *testing.T) {
	logrus.SetOutput(io.Discard)
	polled := make(chan struct{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("foo: bar"))
		select {
		case polled <- struct{}{}:
		default:
		}
	}))
	defer server.Close()

	pool := safe.NewPool(context.Background())
	p := New(static.HttpProvider{Urls: []string{server.URL}})
	require.NoError(t, p.Start(make(chan dynamic.ConfigEvent), pool))
	<-polled

	requireDone(t, pool.Stop)
}
