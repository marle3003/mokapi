package git

import (
	"context"
	"io"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/provider/git/gittest"
	"mokapi/config/static"
	"mokapi/safe"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// requireDone fails the test if f does not return in time, e.g. because a clone or a send on
// the event channel blocks although the pool is stopping.
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

func requireEmptyDir(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Empty(t, entries, "temp dir of the clone is removed")
}

func TestProvider_StopDuringClone(t *testing.T) {
	logrus.SetOutput(io.Discard)
	s := gittest.NewServer(t)
	s.CreateRepo("schemas", map[string]string{"pet.yaml": "name: Odie"})
	arrived, _ := s.HoldClones("schemas")

	tempDir := t.TempDir()
	g := New(static.GitProvider{Urls: []string{s.Url("schemas")}, TempDir: tempDir})
	pool := safe.NewPool(context.Background())
	require.NoError(t, g.Start(make(chan dynamic.ConfigEvent), pool))

	select {
	case <-arrived:
	case <-time.After(5 * time.Second):
		t.Fatal("clone did not start")
	}

	// the clone is held by the server, it only ends because the clone is bound to the pool's context
	requireDone(t, pool.Stop)
	requireEmptyDir(t, tempDir)
}

func TestProvider_StopWhileForwarding(t *testing.T) {
	logrus.SetOutput(io.Discard)
	s := gittest.NewServer(t)
	s.CreateRepo("schemas", map[string]string{"a.yaml": "name: Odie", "b.yaml": "name: Garfield"})

	tempDir := t.TempDir()
	g := New(static.GitProvider{Urls: []string{s.Url("schemas")}, TempDir: tempDir})
	pool := safe.NewPool(context.Background())
	ch := make(chan dynamic.ConfigEvent)
	require.NoError(t, g.Start(ch, pool))

	// forward is now delivering the second file to a consumer that stops reading
	nextEvent(t, ch)

	requireDone(t, pool.Stop)
	requireEmptyDir(t, tempDir)
}
