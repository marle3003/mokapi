package gittest_test

import (
	"mokapi/config/dynamic/provider/git/gittest"
	"os"
	"path/filepath"
	"testing"

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

func requireFile(t *testing.T, dir, name, content string) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(name)))
	require.NoError(t, err)
	require.Equal(t, content, string(b))
}
