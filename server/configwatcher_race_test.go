package server_test

import (
	"context"
	"io"
	"mokapi/config/dynamic"
	"mokapi/config/static"
	"mokapi/safe"
	"mokapi/server"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// Reproduction (before fix): the watcher derives its lookup key from the config's own URL and
// thereby mutates it (OmitHost), while other goroutines read it via ConfigInfo.Key().
func TestConfigWatcher_ConfigKey(t *testing.T) {
	logrus.SetOutput(io.Discard)

	startWatcher := func(t *testing.T) (chan dynamic.ConfigEvent, func()) {
		w := server.NewConfigWatcher(&static.Config{})
		var ch chan dynamic.ConfigEvent
		w.AddProvider("file", &testprovider{start: func(configs chan dynamic.ConfigEvent, pool *safe.Pool) error {
			ch = configs
			return nil
		}})
		pool := safe.NewPool(context.Background())
		require.NoError(t, w.Start(pool))
		return ch, pool.Stop
	}

	t.Run("adding a config does not change its URL or key", func(t *testing.T) {
		ch, stop := startWatcher(t)
		defer stop()

		c := &dynamic.Config{Info: dynamic.ConfigInfo{Url: mustParse("file:///foo/bar.yml")}, Raw: []byte("foo")}
		keyBefore := c.Info.Key()
		var added atomic.Bool
		c.Listeners.Add("test", func(dynamic.ConfigEvent) { added.Store(true) })

		ch <- dynamic.ConfigEvent{Config: c}
		require.Eventually(t, added.Load, 2*time.Second, 5*time.Millisecond)

		require.Equal(t, "file:///foo/bar.yml", c.Info.Url.String())
		require.Equal(t, keyBefore, c.Info.Key())
	})

	t.Run("reading the key while the watcher adds the config", func(t *testing.T) {
		ch, stop := startWatcher(t)
		defer stop()

		c := &dynamic.Config{Info: dynamic.ConfigInfo{Url: mustParse("file:///foo/bar.yml")}, Raw: []byte("foo")}

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				_ = c.Info.Key()
				time.Sleep(time.Millisecond)
			}
		}()
		ch <- dynamic.ConfigEvent{Config: c}
		wg.Wait()
	})
}
