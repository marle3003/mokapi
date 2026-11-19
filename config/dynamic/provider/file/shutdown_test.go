package file

import (
	"context"
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/provider/file/filetest"
	"mokapi/config/static"
	"mokapi/safe"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/require"
)

// requireDone fails the test if f does not return in time, e.g. because a send on the
// event channel blocks although nobody reads it anymore.
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

func newProviderWithUnreadChannel(entries ...*filetest.Entry) (*Provider, chan dynamic.ConfigEvent) {
	p := NewWithWalker(static.FileProvider{}, &filetest.MockFS{Entries: entries})
	ch := make(chan dynamic.ConfigEvent)
	p.ch = ch
	return p, ch
}

func TestWalk_StopsWhenContextIsDone(t *testing.T) {
	p, ch := newProviderWithUnreadChannel(
		&filetest.Entry{Name: "a.txt", Data: []byte("a")},
		&filetest.Entry{Name: "b.txt", Data: []byte("b")},
		&filetest.Entry{Name: "c.txt", Data: []byte("c")},
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var err error
	done := make(chan struct{})
	go func() {
		defer close(done)
		err = p.walk(ctx, static.FileConfig{Path: "./"})
	}()

	// the walk is now delivering the second file to a consumer that stops reading
	<-ch
	cancel()

	requireDone(t, func() { <-done })
	require.NoError(t, err)
}

func TestProcessEvents_StopsWhenContextIsDone(t *testing.T) {
	testcases := []struct {
		name  string
		setup func(p *Provider)
		event fsnotify.Event
	}{
		{
			name:  "changed file",
			event: fsnotify.Event{Name: "foo.txt", Op: fsnotify.Write},
		},
		{
			name:  "removed file",
			setup: func(p *Provider) { p.watched["foo.txt"] = watch{isDir: false} },
			event: fsnotify.Event{Name: "foo.txt", Op: fsnotify.Remove},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p, _ := newProviderWithUnreadChannel(&filetest.Entry{Name: "foo.txt", Data: []byte("foo")})
			if tc.setup != nil {
				tc.setup(p)
			}
			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan struct{})
			go func() {
				defer close(done)
				p.processEvents(ctx, []fsnotify.Event{tc.event})
			}()

			// nobody reads the channel, the event can only leave processEvents through the context
			cancel()

			requireDone(t, func() { <-done })
		})
	}
}

// blockingStatFS holds every Stat call until release is closed, to keep processEvents busy.
type blockingStatFS struct {
	*filetest.MockFS
	entered chan struct{}
	release chan struct{}
}

func (f *blockingStatFS) Stat(name string) (os.FileInfo, error) {
	select {
	case f.entered <- struct{}{}:
	default:
	}
	<-f.release
	return f.MockFS.Stat(name)
}

// The event is processed one second after it arrived (debounce), so this test takes about a second.
func TestStart_PoolStopWaitsForRunningProcessEvents(t *testing.T) {
	blocking := &blockingStatFS{
		MockFS:  &filetest.MockFS{Entries: []*filetest.Entry{{Name: "foo.txt", Data: []byte("foo")}}},
		entered: make(chan struct{}, 1),
		release: make(chan struct{}),
	}
	p := NewWithWalker(static.FileProvider{}, blocking)
	events := make(chan fsnotify.Event)
	p.watcher.Events = events

	pool := safe.NewPool(context.Background())
	canceled := make(chan struct{})
	pool.Go(func(ctx context.Context) {
		<-ctx.Done()
		close(canceled)
	})
	require.NoError(t, p.Start(make(chan dynamic.ConfigEvent), pool))

	events <- fsnotify.Event{Name: "foo.txt", Op: fsnotify.Write}
	select {
	case <-blocking.entered:
	case <-time.After(5 * time.Second):
		t.Fatal("event was not processed")
	}

	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		pool.Stop()
	}()
	<-canceled

	select {
	case <-stopped:
		t.Error("pool.Stop returned while processEvents is still running")
	case <-time.After(200 * time.Millisecond):
	}

	close(blocking.release)
	requireDone(t, func() { <-stopped })
}

func TestProcessEvents_RemoveWhileWatchingPaths(t *testing.T) {
	var entries []*filetest.Entry
	var events []fsnotify.Event
	for i := 0; i < 200; i++ {
		name := fmt.Sprintf("dir%d", i)
		entries = append(entries, &filetest.Entry{Name: name, IsDir: true})
		events = append(events, fsnotify.Event{Name: name, Op: fsnotify.Remove})
	}
	p, _ := newProviderWithUnreadChannel(entries...)

	done := make(chan struct{})
	go func() {
		defer close(done)
		for _, e := range entries {
			p.watchPath(e.Name)
		}
	}()
	p.processEvents(context.Background(), events)
	<-done
}

// Start walks the configured directories in one pool goroutine while Watch (npm provider)
// walks another directory in a second one. Both read and write the .mokapiignore files.
func TestWalk_ConcurrentWalksWithMokapiIgnoreFiles(t *testing.T) {
	var entries []*filetest.Entry
	for _, root := range []string{"a", "b"} {
		entries = append(entries, &filetest.Entry{Name: root, IsDir: true})
		for i := 0; i < 200; i++ {
			dir := fmt.Sprintf("%s/dir%d", root, i)
			entries = append(entries,
				&filetest.Entry{Name: dir, IsDir: true},
				&filetest.Entry{Name: dir + "/.mokapiignore", Data: []byte("*.log")},
				&filetest.Entry{Name: dir + "/foo.txt", Data: []byte("foo")},
			)
		}
	}
	p := NewWithWalker(static.FileProvider{}, &filetest.MockFS{Entries: entries})
	ch := make(chan dynamic.ConfigEvent)
	p.ch = ch

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		for {
			select {
			case <-ch:
			case <-ctx.Done():
				return
			}
		}
	}()

	var wg sync.WaitGroup
	for _, root := range []string{"a", "b"} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			require.NoError(t, p.walk(ctx, static.FileConfig{Path: root}))
		}()
	}
	wg.Wait()
}

// skip combines the provider's patterns with those of a directory. The result must be a new
// slice, otherwise concurrent walks write their patterns into the shared backing array.
func TestSkip_DoesNotWriteIntoProviderPatterns(t *testing.T) {
	testcases := []struct {
		name  string
		setup func(cfg *static.FileProvider) []string
		info  static.FileConfig
	}{
		{
			name: "include",
			setup: func(cfg *static.FileProvider) []string {
				backing := make([]string, 1, 10)
				backing[0] = "*.txt"
				cfg.Include = backing
				return backing
			},
			info: static.FileConfig{Include: []string{"*.json"}},
		},
		{
			name: "exclude",
			setup: func(cfg *static.FileProvider) []string {
				backing := make([]string, 1, 10)
				backing[0] = "*.log"
				cfg.Exclude = backing
				return backing
			},
			info: static.FileConfig{Exclude: []string{"*.tmp"}},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			var cfg static.FileProvider
			backing := tc.setup(&cfg)
			p := NewWithWalker(cfg, &filetest.MockFS{})

			p.skip("foo.txt", false, tc.info)

			require.Empty(t, backing[:2][1])
		})
	}
}
