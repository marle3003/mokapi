package runtime

import (
	"context"
	"mokapi/config/static"
	"mokapi/runtime/search"
	"mokapi/safe"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSearchIndex_AddWhileStarting(t *testing.T) {
	s := newSearchIndex(static.Search{Enabled: true, InMemory: true})
	pool := safe.NewPool(context.Background())
	defer pool.Stop()

	const n = 50
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Add("id"+string(rune('a'+i)), map[string]any{"name": "item"})
		}()
	}
	go s.start(pool)
	wg.Wait()
	// Search waits until start() has flushed the queue (it closes ready after its flush wait)

	result, err := s.Search(search.Request{Limit: 100})
	require.NoError(t, err)
	require.Equal(t, uint64(n), result.Total)
}

func TestSearchIndex_AddAfterFailedStart(t *testing.T) {
	// an index path that is a file makes creating the index fail
	file := filepath.Join(t.TempDir(), "file")
	require.NoError(t, os.WriteFile(file, nil, 0600))
	s := newSearchIndex(static.Search{Enabled: true, IndexPath: file})

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				s.Add("id", map[string]any{"name": "item"})
				s.Delete("id")
			}
		}()
	}
	pool := safe.NewPool(context.Background())
	t.Cleanup(pool.Stop)
	s.start(pool)
	wg.Wait()

	_, err := s.Search(search.Request{})
	require.ErrorIs(t, err, &search.ErrNotEnabled{})
}

func TestSearchIndex_SearchWhenDisabled(t *testing.T) {
	s := newSearchIndex(static.Search{Enabled: false})

	done := make(chan error, 1)
	go func() {
		_, err := s.Search(search.Request{})
		done <- err
	}()

	select {
	case err := <-done:
		require.EqualError(t, err, "search is not enabled")
	case <-time.After(time.Second):
		t.Fatal("Search blocks although search is disabled")
	}
}
