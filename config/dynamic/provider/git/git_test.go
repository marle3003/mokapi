package git

import (
	"context"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/static"
	"mokapi/safe"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var gitFiles = map[string]struct{}{"LICENSE": {}, "README.md": {}, "models.yml": {}, "openapi.yml": {}}

func TestGit(t *testing.T) {
	g := New(static.GitProvider{Url: "https://github.com/marle3003/mokapi-example.git"})
	p := safe.NewPool(context.Background())
	defer p.Stop()
	ch := make(chan *dynamic.Config)
	err := g.Start(ch, p)
	require.NoError(t, err)

	timeout := time.After(1 * time.Second)
	i := 0
Stop:
	for {
		select {
		case <-timeout:
			break Stop
		case c := <-ch:
			i++
			name := filepath.Base(c.Info.Inner().Url.String())
			_, ok := gitFiles[name]
			assert.True(t, ok)
			require.Equal(t, "git", c.Info.Provider)
			require.Equal(t, "https://github.com/marle3003/mokapi-example.git?file=/"+name, c.Info.Path())
		}
	}
	require.Len(t, gitFiles, i)
}

func TestGit_Branch(t *testing.T) {
	g := New(static.GitProvider{Url: "https://github.com/marle3003/mokapi-example.git?ref=main"})
	p := safe.NewPool(context.Background())
	defer func() {
		p.Stop()
	}()
	ch := make(chan *dynamic.Config)
	err := g.Start(ch, p)
	require.NoError(t, err)

	timeout := time.After(1 * time.Second)
	i := 0
Stop:
	for {
		select {
		case <-timeout:
			break Stop
		case c := <-ch:
			i++
			name := filepath.Base(c.Info.Inner().Url.String())
			_, ok := gitFiles[name]
			assert.True(t, ok)
		}
	}
	require.Len(t, gitFiles, i)
}

func TestGit_MultipleUrls(t *testing.T) {
	g := New(static.GitProvider{Urls: []string{
		"https://github.com/marle3003/mokapi-example.git",
		"https://github.com/marle3003/mokapi-example.git?ref=main",
	}})
	p := safe.NewPool(context.Background())
	defer func() {
		p.Stop()
	}()
	ch := make(chan *dynamic.Config)
	err := g.Start(ch, p)
	require.NoError(t, err)

	timeout := time.After(1 * time.Second)
	files := map[string]*dynamic.Config{}
Stop:
	for {
		select {
		case <-timeout:
			break Stop
		case c := <-ch:
			files[c.Info.Url.String()] = c
		}
	}
	require.Len(t, files, 8)
	require.Contains(t, files, "https://github.com/marle3003/mokapi-example.git?file=%2FLICENSE")
	require.Contains(t, files, "https://github.com/marle3003/mokapi-example.git?file=%2FLICENSE&ref=main")
}

// go-git requires git installed for file:// repositories
func testGit_SimpleUrl(t *testing.T) {
	repo := newGitRepo(t, t.Name())
	defer func() {
		err := os.RemoveAll(repo.dir)
		require.NoError(t, err)
	}()

	repo.commit(t, "foo.txt", "bar")

	g := New(static.GitProvider{Url: repo.url.String()})
	p := safe.NewPool(context.Background())
	defer p.Stop()

	ch := make(chan *dynamic.Config)
	err := g.Start(ch, p)
	require.NoError(t, err)

	select {
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout")
	case c := <-ch:
		require.Equal(t, "foo.txt", filepath.Base(c.Info.Url.String()))
	}
}

// Not implemented correctly by go-git https://github.com/go-git/go-git/issues/90
func testGit_SparseUrl(t *testing.T) {
	repo := newGitRepo(t, t.Name())

	repo.commit(t, "foo/foo.txt", "bar")
	repo.commit(t, "bar/bar.txt", "bar")

	g := New(static.GitProvider{Url: repo.url.String() + "//foo"})
	p := safe.NewPool(context.Background())
	defer p.Stop()

	ch := make(chan *dynamic.Config)
	err := g.Start(ch, p)
	require.NoError(t, err)

	files := make(map[string]string)
Stop:
	for {
		select {
		case <-time.After(3 * time.Second):
			break Stop
		case c := <-ch:
			files[filepath.Base(c.Info.Url.String())] = c.Info.Url.String()
		}
	}

	require.Len(t, files, 1)
	require.Contains(t, files, "foo.txt")
}

type gitTestRepo struct {
	url  *url.URL
	dir  string
	repo *git.Repository
}

func newGitRepo(t *testing.T, name string) *gitTestRepo {
	dir, err := os.MkdirTemp("", "mokapi")
	require.NoError(t, err)
	t.Cleanup(func() {
		err := os.RemoveAll(dir)
		require.NoError(t, err)
	})

	repoDir := filepath.Join(dir, name)
	err = os.Mkdir(repoDir, 0700)
	require.NoError(t, err)

	r, err := git.PlainInit(repoDir, false)
	require.NoError(t, err)

	u, err := url.Parse("file://" + filepath.ToSlash(repoDir))
	require.NoError(t, err)

	return &gitTestRepo{
		url:  u,
		dir:  repoDir,
		repo: r,
	}
}

func (g *gitTestRepo) commit(t *testing.T, file, content string) {
	path := filepath.Join(g.dir, file)
	err := os.MkdirAll(filepath.Dir(path), 0700)
	require.NoError(t, err)

	err = os.WriteFile(path, []byte(content), 0600)
	require.NoError(t, err)

	w, err := g.repo.Worktree()
	require.NoError(t, err)

	_, err = w.Add(file)
	require.NoError(t, err)
	ts, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05-07:00")
	require.NoError(t, err)
	_, err = w.Commit("added "+file, &git.CommitOptions{Author: &object.Signature{When: ts}})
	require.NoError(t, err)
}
