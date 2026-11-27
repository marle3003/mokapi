package gittest_test

import (
	"context"
	"mokapi/config/dynamic/provider/git/gittest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	t.Run("shallow clone, commit and pull", func(t *testing.T) {
		s := gittest.NewServer(t)
		repo := s.CreateRepo("schemas", map[string]string{"pet.yaml": "name: Odie", "dir/a.txt": "a"})

		dir := t.TempDir()
		clone, err := git.PlainClone(dir, false, &git.CloneOptions{URL: s.Url("schemas"), Depth: 1, SingleBranch: true})
		require.NoError(t, err)
		require.Equal(t, 1, s.CloneCount("schemas"))
		requireFile(t, dir, "pet.yaml", "name: Odie")
		requireFile(t, dir, "dir/a.txt", "a")

		wt, err := clone.Worktree()
		require.NoError(t, err)
		err = wt.Pull(&git.PullOptions{SingleBranch: true, Depth: 1})
		require.ErrorIs(t, err, git.NoErrAlreadyUpToDate)
		require.Equal(t, 1, s.CloneCount("schemas"), "pull without news downloads nothing")

		repo.Commit(map[string]string{"pet.yaml": "name: Garfield"})
		require.NoError(t, wt.Pull(&git.PullOptions{SingleBranch: true, Depth: 1}))
		requireFile(t, dir, "pet.yaml", "name: Garfield")

		repo.Remove("dir/a.txt")
		require.NoError(t, wt.Pull(&git.PullOptions{SingleBranch: true, Depth: 1}))
		require.NoFileExists(t, filepath.Join(dir, "dir", "a.txt"))
	})

	t.Run("branch", func(t *testing.T) {
		s := gittest.NewServer(t)
		repo := s.CreateRepo("schemas", map[string]string{"pet.yaml": "name: Odie"})
		repo.CreateBranch("develop")
		repo.Commit(map[string]string{"pet.yaml": "name: Garfield"})

		dir := t.TempDir()
		_, err := git.PlainClone(dir, false, &git.CloneOptions{
			URL: s.Url("schemas"), Depth: 1, SingleBranch: true,
			ReferenceName: plumbing.NewBranchReferenceName("develop"),
		})
		require.NoError(t, err)
		requireFile(t, dir, "pet.yaml", "name: Garfield")

		dir = t.TempDir()
		_, err = git.PlainClone(dir, false, &git.CloneOptions{URL: s.Url("schemas"), Depth: 1, SingleBranch: true})
		require.NoError(t, err)
		requireFile(t, dir, "pet.yaml", "name: Odie")

		_, err = git.PlainClone(t.TempDir(), false, &git.CloneOptions{
			URL: s.Url("schemas"), Depth: 1, SingleBranch: true,
			ReferenceName: plumbing.NewBranchReferenceName("unknown"),
		})
		require.Error(t, err)
	})

	t.Run("unknown repository", func(t *testing.T) {
		s := gittest.NewServer(t)
		_, err := git.PlainClone(t.TempDir(), false, &git.CloneOptions{URL: s.Url("unknown"), Depth: 1})
		require.Error(t, err)
		require.Equal(t, 0, s.CloneCount("unknown"))
	})
}

func TestServer_HoldClones(t *testing.T) {
	clone := func(ctx context.Context, url string) error {
		_, err := git.PlainCloneContext(ctx, t.TempDir(), false, &git.CloneOptions{URL: url, Depth: 1, SingleBranch: true})
		return err
	}

	t.Run("clone waits for release and then succeeds", func(t *testing.T) {
		s := gittest.NewServer(t)
		s.CreateRepo("schemas", map[string]string{"pet.yaml": "name: Odie"})
		arrived, release := s.HoldClones("schemas")

		done := make(chan error, 1)
		go func() { done <- clone(context.Background(), s.Url("schemas")) }()

		select {
		case <-arrived:
		case <-time.After(5 * time.Second):
			t.Fatal("clone request did not arrive")
		}
		select {
		case err := <-done:
			t.Fatalf("clone finished while held: %v", err)
		case <-time.After(200 * time.Millisecond):
		}

		release()
		release() // idempotent
		select {
		case err := <-done:
			require.NoError(t, err)
		case <-time.After(5 * time.Second):
			t.Fatal("clone did not finish after release")
		}
	})

	t.Run("cancelled context ends a held clone", func(t *testing.T) {
		s := gittest.NewServer(t)
		s.CreateRepo("schemas", map[string]string{"pet.yaml": "name: Odie"})
		arrived, _ := s.HoldClones("schemas")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		done := make(chan error, 1)
		go func() { done <- clone(ctx, s.Url("schemas")) }()

		select {
		case <-arrived:
		case <-time.After(5 * time.Second):
			t.Fatal("clone request did not arrive")
		}
		cancel()
		select {
		case err := <-done:
			require.ErrorContains(t, err, "context canceled")
		case <-time.After(5 * time.Second):
			t.Fatal("held clone was not cancelled")
		}
	})

	t.Run("each hold releases only the clones it caught", func(t *testing.T) {
		s := gittest.NewServer(t)
		s.CreateRepo("schemas", map[string]string{"pet.yaml": "name: Odie"})
		arrivedA, releaseFirst := s.HoldClones("schemas")

		doneA := make(chan error, 1)
		go func() { doneA <- clone(context.Background(), s.Url("schemas")) }()
		select {
		case <-arrivedA:
		case <-time.After(5 * time.Second):
			t.Fatal("clone A request did not arrive")
		}

		arrivedB, releaseSecond := s.HoldClones("schemas")
		doneB := make(chan error, 1)
		go func() { doneB <- clone(context.Background(), s.Url("schemas")) }()
		select {
		case <-arrivedB:
		case <-time.After(5 * time.Second):
			t.Fatal("clone B request did not arrive")
		}

		releaseFirst()
		select {
		case err := <-doneA:
			require.NoError(t, err)
		case <-time.After(5 * time.Second):
			t.Fatal("clone A did not finish after its release")
		}
		select {
		case err := <-doneB:
			t.Fatalf("clone B finished while held: %v", err)
		case <-time.After(200 * time.Millisecond):
		}

		releaseSecond()
		select {
		case err := <-doneB:
			require.NoError(t, err)
		case <-time.After(5 * time.Second):
			t.Fatal("clone B did not finish after its release")
		}
	})

	t.Run("other repository is not held", func(t *testing.T) {
		s := gittest.NewServer(t)
		s.CreateRepo("schemas", map[string]string{"pet.yaml": "name: Odie"})
		s.CreateRepo("other", map[string]string{"pet.yaml": "name: Garfield"})
		s.HoldClones("schemas")

		require.NoError(t, clone(context.Background(), s.Url("other")))
	})

	t.Run("unreleased hold does not block the test cleanup", func(t *testing.T) {
		requireCleanupDoesNotHang(t, clone, func(s *gittest.Server) {})
	})

	t.Run("replaced hold does not block the test cleanup", func(t *testing.T) {
		requireCleanupDoesNotHang(t, clone, func(s *gittest.Server) { s.HoldClones("schemas") })
	})
}

// requireCleanupDoesNotHang holds a clone of "schemas", calls afterArrived and ends the inner
// test without releasing the first hold. If the cleanup of the inner test hangs, a watchdog
// releases the hold itself, so a bug fails the test instead of hanging the test binary.
func requireCleanupDoesNotHang(t *testing.T, clone func(ctx context.Context, url string) error, afterArrived func(s *gittest.Server)) {
	cloneDone := make(chan struct{})
	var cleanupHung atomic.Bool

	t.Run("inner", func(inner *testing.T) {
		cleanupFinished := make(chan struct{})
		inner.Cleanup(func() { close(cleanupFinished) }) // runs after the server cleanups

		s := gittest.NewServer(inner)
		s.CreateRepo("schemas", map[string]string{"pet.yaml": "name: Odie"})
		arrived, release := s.HoldClones("schemas")
		go func() {
			select {
			case <-cleanupFinished:
			case <-time.After(3 * time.Second):
				cleanupHung.Store(true)
				release()
			}
		}()

		go func() {
			defer close(cloneDone)
			_ = clone(context.Background(), s.Url("schemas"))
		}()
		select {
		case <-arrived:
		case <-time.After(5 * time.Second):
			inner.Fatal("clone request did not arrive")
		}
		afterArrived(s)
	})

	require.False(t, cleanupHung.Load(), "cleanup hangs on a held request")
	select {
	case <-cloneDone:
	case <-time.After(5 * time.Second):
		t.Fatal("clone goroutine did not end")
	}
}

func requireFile(t *testing.T, dir, name, content string) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(name)))
	require.NoError(t, err)
	require.Equal(t, content, string(b))
}
