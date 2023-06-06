package http

import (
	"context"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/common"
	"mokapi/config/static"
	"mokapi/safe"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

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
