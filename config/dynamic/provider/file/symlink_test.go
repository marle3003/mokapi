package file

import (
	"context"
	"mokapi/config/dynamic"
	"mokapi/config/static"
	"mokapi/safe"
	"mokapi/try"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/stretchr/testify/require"
)

// mustSymlink skips the test where the OS does not allow creating symlinks (e.g. Windows without privileges).
func mustSymlink(t *testing.T, target, link string) {
	t.Helper()
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks not supported: %v", err)
	}
}

func startProvider(t *testing.T, dir string) chan dynamic.ConfigEvent {
	t.Helper()
	_, ch := startProviderWithHandle(t, dir)
	return ch
}

func startProviderWithHandle(t *testing.T, dir string) (*Provider, chan dynamic.ConfigEvent) {
	t.Helper()
	p := New(static.FileProvider{Directories: []static.FileConfig{{Path: dir}}})
	pool := safe.NewPool(context.Background())
	t.Cleanup(pool.Stop)
	ch := make(chan dynamic.ConfigEvent)
	require.NoError(t, p.Start(ch, pool))
	return p, ch
}

func TestSymlink_FileInWatchedDirectory(t *testing.T) {
	root, external := t.TempDir(), t.TempDir()
	require.NoError(t, createFile(filepath.Join(external, "foo.js"), "foo"))
	mustSymlink(t, filepath.Join(external, "foo.js"), filepath.Join(root, "link.js"))

	ch := startProvider(t, root)

	c := waitFileUpdate(t, ch)
	require.Equal(t, "foo", string(c.Raw))
}

func TestSymlink_DirectoryInWatchedDirectory(t *testing.T) {
	root, external := t.TempDir(), t.TempDir()
	require.NoError(t, createFile(filepath.Join(external, "foo.js"), "foo"))
	mustSymlink(t, external, filepath.Join(root, "linked"))

	ch := startProvider(t, root)

	c := waitFileUpdate(t, ch)
	require.Equal(t, "foo", string(c.Raw))
}

func TestSymlink_DirectoryIsRootPath(t *testing.T) {
	external := t.TempDir()
	require.NoError(t, createFile(filepath.Join(external, "foo.js"), "foo"))
	link := filepath.Join(t.TempDir(), "root-link")
	mustSymlink(t, external, link)

	ch := startProvider(t, link)

	c := waitFileUpdate(t, ch)
	require.Equal(t, "foo", string(c.Raw))
}

func TestSymlink_DirectoryCreatedAtRuntime(t *testing.T) {
	root, external := t.TempDir(), t.TempDir()
	require.NoError(t, createFile(filepath.Join(external, "foo.js"), "foo"))

	p, ch := startProviderWithHandle(t, root)
	try.WaitFor(t, func() bool { return !p.isInit.Load() && p.isWatchPath(root) })
	mustSymlink(t, external, filepath.Join(root, "linked"))

	c := waitFileUpdate(t, ch)
	require.Equal(t, "foo", string(c.Raw))
}

// collectEvents returns the events received until none arrives for quiet.
func collectEvents(ch chan dynamic.ConfigEvent, quiet time.Duration) []dynamic.ConfigEvent {
	var events []dynamic.ConfigEvent
	for {
		select {
		case e := <-ch:
			events = append(events, e)
		case <-time.After(quiet):
			return events
		}
	}
}

func eventNames(events []dynamic.ConfigEvent) []string {
	var names []string
	for _, e := range events {
		names = append(names, e.Name)
	}
	return names
}

func TestSymlink_CycleToParentDirectory(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, createFile(filepath.Join(root, "sub", "foo.js"), "foo"))
	mustSymlink(t, filepath.Join(root, "sub"), filepath.Join(root, "sub", "loop"))

	ch := startProvider(t, root)

	events := collectEvents(ch, time.Second)
	require.Equal(t, []string{filepath.Join(root, "sub", "foo.js")}, eventNames(events))
}

func TestSymlink_CycleToAncestorOfRoot(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "root")
	require.NoError(t, createFile(filepath.Join(root, "foo.js"), "foo"))
	mustSymlink(t, base, filepath.Join(root, "up"))

	ch := startProvider(t, root)

	// "up" leads to base, which contains root again: that part is cut as a cycle
	events := collectEvents(ch, time.Second)
	require.Equal(t, []string{filepath.Join(root, "foo.js")}, eventNames(events))
}

func TestSymlink_BrokenLinkIsSkipped(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, createFile(filepath.Join(root, "foo.js"), "foo"))
	mustSymlink(t, filepath.Join(root, "missing"), filepath.Join(root, "broken"))

	ch := startProvider(t, root)

	events := collectEvents(ch, time.Second)
	require.Equal(t, []string{filepath.Join(root, "foo.js")}, eventNames(events))
}

func TestSymlink_UpdateInLinkedDirectory(t *testing.T) {
	root, external := t.TempDir(), t.TempDir()
	require.NoError(t, createFile(filepath.Join(external, "foo.js"), "foo"))
	mustSymlink(t, external, filepath.Join(root, "linked"))

	ch := startProvider(t, root)
	waitFileUpdate(t, ch)

	require.NoError(t, createFile(filepath.Join(external, "foo.js"), "bar"))

	select {
	case e := <-ch:
		require.Equal(t, dynamic.Update, e.Event)
		require.Equal(t, filepath.Join(root, "linked", "foo.js"), e.Name)
		require.Equal(t, "bar", string(e.Config.Raw))
	case <-time.After(5 * time.Second):
		t.Fatal("timeout while waiting for update event")
	}
}

func TestSymlink_SameTargetViaTwoLinks(t *testing.T) {
	root, external := t.TempDir(), t.TempDir()
	require.NoError(t, createFile(filepath.Join(external, "foo.js"), "foo"))
	mustSymlink(t, external, filepath.Join(root, "a"))
	mustSymlink(t, external, filepath.Join(root, "b"))

	ch := startProvider(t, root)

	// not a cycle: the target is read once per link, each under its own path
	events := collectEvents(ch, time.Second)
	require.ElementsMatch(t, []string{
		filepath.Join(root, "a", "foo.js"),
		filepath.Join(root, "b", "foo.js"),
	}, eventNames(events))
}

// updatedNames returns each updated path once, a rewritten file may produce a chmod and a write event.
func updatedNames(events []dynamic.ConfigEvent) []string {
	var names []string
	for _, e := range events {
		if e.Event == dynamic.Update && !slices.Contains(names, e.Name) {
			names = append(names, e.Name)
		}
	}
	return names
}

func TestSymlink_UpdateWhenLinkTargetsDirectoryInsideRoot(t *testing.T) {
	root := t.TempDir()
	real := filepath.Join(root, "sub", "foo.js")
	require.NoError(t, createFile(real, "foo"))
	mustSymlink(t, filepath.Join(root, "sub"), filepath.Join(root, "a"))

	ch := startProvider(t, root)
	require.Len(t, collectEvents(ch, time.Second), 2)

	require.NoError(t, createFile(real, "bar"))

	// the real file and the file reached via the link are both loaded, so both are updated
	events := collectEvents(ch, 2*time.Second)
	require.ElementsMatch(t, []string{real, filepath.Join(root, "a", "foo.js")}, updatedNames(events))
}

func TestSymlink_UpdateWhenSameTargetViaTwoLinks(t *testing.T) {
	root, external := t.TempDir(), t.TempDir()
	require.NoError(t, createFile(filepath.Join(external, "foo.js"), "foo"))
	mustSymlink(t, external, filepath.Join(root, "a"))
	mustSymlink(t, external, filepath.Join(root, "b"))

	ch := startProvider(t, root)
	require.Len(t, collectEvents(ch, time.Second), 2)

	require.NoError(t, createFile(filepath.Join(external, "foo.js"), "bar"))

	events := collectEvents(ch, 2*time.Second)
	require.ElementsMatch(t, []string{
		filepath.Join(root, "a", "foo.js"),
		filepath.Join(root, "b", "foo.js"),
	}, updatedNames(events))
}

// The watcher is fed by hand: kqueue (macOS) reports no event when a symlink to a directory is removed.
func TestSymlink_RemovedLinkReleasesResolvedWatch(t *testing.T) {
	root, external := t.TempDir(), t.TempDir()
	require.NoError(t, createFile(filepath.Join(external, "foo.js"), "foo"))
	a, b := filepath.Join(root, "a"), filepath.Join(root, "b")
	mustSymlink(t, external, a)
	mustSymlink(t, external, b)

	p, ch := startProviderWithHandle(t, root)
	require.Len(t, collectEvents(ch, time.Second), 2)
	resolved, err := filepath.EvalSymlinks(external)
	require.NoError(t, err)
	require.Contains(t, p.watcher.WatchList(), resolved)

	p.processEvents(context.Background(), []fsnotify.Event{{Name: a, Op: fsnotify.Remove}})
	require.Equal(t, []string{b}, p.aliases[resolved], "still used by the other link")
	require.Contains(t, p.watcher.WatchList(), resolved)

	p.processEvents(context.Background(), []fsnotify.Event{{Name: b, Op: fsnotify.Remove}})
	require.NotContains(t, p.aliases, resolved)
	require.NotContains(t, p.watcher.WatchList(), resolved)
}
