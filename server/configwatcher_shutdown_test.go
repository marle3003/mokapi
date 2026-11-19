package server_test

import (
	"context"
	"fmt"
	"mokapi/config/static"
	"mokapi/safe"
	"mokapi/server"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Stopping the pool while the file provider still walks a directory must not make the walk
// send on the event channel that the ConfigWatcher closes on shutdown.
func TestConfigWatcher_StopWhileFileProviderIsWalking(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 2000; i++ {
		name := filepath.Join(dir, fmt.Sprintf("spec-%04d.json", i))
		require.NoError(t, os.WriteFile(name, []byte(`{"openapi":"3.0","info":{"title":"foo"}}`), 0o644))
	}

	for i := 0; i < 20; i++ {
		w := server.NewConfigWatcher(&static.Config{
			Providers: static.Providers{File: static.FileProvider{Directories: []static.FileConfig{{Path: dir}}}},
		})
		pool := safe.NewPool(context.Background())
		require.NoError(t, w.Start(pool))

		// let the walk start, the consumer is slower than the walk
		time.Sleep(5 * time.Millisecond)
		stopped := make(chan struct{})
		go func() {
			pool.Stop()
			close(stopped)
		}()
		select {
		case <-stopped:
		case <-time.After(10 * time.Second):
			t.Fatal("pool.Stop() blocked, a provider still waits for the event channel")
		}
	}
}
