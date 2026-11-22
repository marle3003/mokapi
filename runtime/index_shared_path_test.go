package runtime

import (
	"context"
	"fmt"
	"mokapi/config/static"
	"mokapi/runtime/search"
	"mokapi/safe"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const sharedPathChildEnv = "MOKAPI_TEST_SHARED_INDEX_CHILD"

// Several processes (go test runs packages as separate processes) use the default
// on-disk index at the same time. Without a directory per index they would share
// os.TempDir()/mokapi-bleve-index, and each start() would delete the index the others use.
func TestSearchIndex_ProcessesWithDefaultPath(t *testing.T) {
	if os.Getenv(sharedPathChildEnv) != "" {
		runIndexLikeATestPackage(t)
		return
	}

	var wg sync.WaitGroup
	outputs := make([]string, 3)
	for p := range outputs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestSearchIndex_ProcessesWithDefaultPath$", "-test.count=1")
			cmd.Env = append(os.Environ(), sharedPathChildEnv+"=1")
			out, err := cmd.CombinedOutput()
			if err != nil {
				outputs[p] = string(out)
			}
		}()
	}
	wg.Wait()
	for _, out := range outputs {
		require.Empty(t, out)
	}
}

func runIndexLikeATestPackage(t *testing.T) {
	pool := safe.NewPool(context.Background())
	defer pool.Stop()
	for round := 0; round < 4; round++ {
		s := newSearchIndex(static.Search{Enabled: true})
		for i := 0; i < 200; i++ {
			s.Add(fmt.Sprintf("a-%d", i), map[string]any{"name": "item"})
		}
		s.start(pool)
		for i := 0; i < 30; i++ {
			result, err := s.Search(search.Request{Limit: 500})
			require.NoError(t, err)
			require.Equal(t, uint64(200), result.Total)
		}
	}
}
