package file

import (
	"context"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/provider/file/filetest"
	"mokapi/config/static"
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
