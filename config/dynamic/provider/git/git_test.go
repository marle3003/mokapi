package git

import (
	"context"
	"io"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/provider/git/gittest"
	"mokapi/config/static"
	"mokapi/safe"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// quietPeriod is how long collectEvents waits for events beyond the expected ones.
const quietPeriod = 500 * time.Millisecond

var gitFiles = map[string]struct{}{
	"LICENSE":     {},
	"README.md":   {},
	"models.yml":  {},
	"openapi.yml": {},
}

// createExampleRepo creates the repository example.git with the files of gitFiles.
func createExampleRepo(t *testing.T, s *gittest.Server) *gittest.Repo {
	t.Helper()
	files := make(map[string]string)
	for name := range gitFiles {
		files[name] = "content of " + name
	}
	return s.CreateRepo("example.git", files)
}

// collectEvents waits until want events have arrived, then collects further events for
// quietPeriod. The result has more than want events if the provider sends too many.
func collectEvents(t *testing.T, ch chan dynamic.ConfigEvent, want int) []dynamic.ConfigEvent {
	t.Helper()
	count := 0
	events := waitForEvents(t, ch, func(dynamic.ConfigEvent) bool {
		count++
		return count == want
	})
	quiet := time.After(quietPeriod)
	for {
		select {
		case e := <-ch:
			events = append(events, e)
		case <-quiet:
			return events
		}
	}
}

// innerPath is the local path of the cloned file.
func innerPath(e dynamic.ConfigEvent) string {
	inner := e.Config.Info.Inner().Url
	if len(inner.Opaque) > 0 {
		return inner.Opaque
	}
	return inner.Path
}

// nextEvent waits for one event.
func nextEvent(t *testing.T, ch chan dynamic.ConfigEvent) dynamic.ConfigEvent {
	t.Helper()
	return waitForEvents(t, ch, func(dynamic.ConfigEvent) bool { return true })[0]
}

// startWithConfig starts the provider with the config as it is, unlike startProvider, which sets
// a temp dir and a pull interval. The pool is stopped at the end of the test, the channel is not
// read.
func startWithConfig(t *testing.T, cfg static.GitProvider) chan dynamic.ConfigEvent {
	t.Helper()
	logrus.SetOutput(io.Discard)
	g := New(cfg)
	p := safe.NewPool(context.Background())
	t.Cleanup(p.Stop)
	ch := make(chan dynamic.ConfigEvent)
	require.NoError(t, g.Start(ch, p))
	return ch
}

func TestGit(t *testing.T) {
	testcases := []struct {
		name  string
		cfg   func(repoUrl string) static.GitProvider
		count int
		test  func(t *testing.T, repoUrl string, events []dynamic.ConfigEvent)
	}{
		{
			name:  "clone",
			cfg:   func(u string) static.GitProvider { return static.GitProvider{Urls: []string{u}} },
			count: 4,
			test: func(t *testing.T, repoUrl string, events []dynamic.ConfigEvent) {
				for _, e := range events {
					name := filepath.Base(e.Config.Info.Inner().Url.String())
					path := innerPath(e)

					require.True(t, strings.HasPrefix(path, os.TempDir()), "file is in system default temp path: %v", path)
					require.Equal(t, "git", e.Config.Info.Provider)
					require.Equal(t, repoUrl+"?file=/"+name, e.Config.Info.Path())
					require.Contains(t, gitFiles, name)
					require.NotNil(t, e.Config.Info.Checksum)
					require.Equal(t, []string{"main"}, e.Config.Info.Tags)
				}
			},
		},
		{
			name:  "clone a branch",
			cfg:   func(u string) static.GitProvider { return static.GitProvider{Urls: []string{u + "?ref=main"}} },
			count: 4,
			test:  requireGitFiles,
		},
		{
			name: "files only models.yml",
			cfg: func(u string) static.GitProvider {
				return static.GitProvider{Repositories: []static.GitRepo{{Url: u + "?ref=main", Files: []string{"models.yml"}}}}
			},
			count: 1,
			test:  requireGitFiles,
		},
		{
			name: "include *.yml",
			cfg: func(u string) static.GitProvider {
				return static.GitProvider{Repositories: []static.GitRepo{{Url: u + "?ref=main", Include: []string{"*.yml"}}}}
			},
			count: 2,
			test:  requireGitFiles,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s := gittest.NewServer(t)
			createExampleRepo(t, s)
			repoUrl := s.Url("example.git")

			ch := startWithConfig(t, tc.cfg(repoUrl))

			events := collectEvents(t, ch, tc.count)
			require.Len(t, events, tc.count)
			tc.test(t, repoUrl, events)
		})
	}
}

func requireGitFiles(t *testing.T, _ string, events []dynamic.ConfigEvent) {
	for _, e := range events {
		require.Contains(t, gitFiles, filepath.Base(e.Config.Info.Inner().Url.String()))
	}
}

func TestGit_MultipleUrls(t *testing.T) {
	s := gittest.NewServer(t)
	createExampleRepo(t, s)
	repoUrl := s.Url("example.git")

	ch := startWithConfig(t, static.GitProvider{Urls: []string{repoUrl, repoUrl + "?ref=main"}})

	events := collectEvents(t, ch, 8)
	require.Len(t, events, 8)
	urls := map[string]struct{}{}
	for _, e := range events {
		urls[e.Config.Info.Url.String()] = struct{}{}
	}
	require.Len(t, urls, 8)
	require.Contains(t, urls, repoUrl+"?file=%2FLICENSE")
	require.Contains(t, urls, repoUrl+"?file=%2FLICENSE&ref=main")
	require.Equal(t, 2, s.CloneCount("example.git"), "one clone per branch key")
}

func TestCustomTempDir(t *testing.T) {
	s := gittest.NewServer(t)
	createExampleRepo(t, s)
	cfg := static.GitProvider{
		Repositories: []static.GitRepo{{Url: s.Url("example.git") + "?ref=main", Files: []string{"models.yml"}}},
		TempDir:      t.TempDir(),
	}

	ch := startWithConfig(t, cfg)

	events := collectEvents(t, ch, 1)
	require.Len(t, events, 1)
	path := innerPath(events[0])
	require.True(t, strings.HasPrefix(path, cfg.TempDir), "file %v is in custom temp path %v", path, cfg.TempDir)
}

func TestGitFileUpdate(t *testing.T) {
	s := gittest.NewServer(t)
	repo := s.CreateRepo("update", map[string]string{"foo.txt": "foo"})

	ch := startWithConfig(t, static.GitProvider{Urls: []string{s.Url("update")}, PullInterval: "1s"})

	e := nextEvent(t, ch)
	initChecksum := e.Config.Info.Checksum
	initTime := e.Config.Info.Time

	repo.Commit(map[string]string{"foo.txt": "bar"})

	// git deletes the file first
	events := waitForEvents(t, ch, updated("bar"))
	require.Len(t, events, 2)
	require.Equal(t, dynamic.Delete, events[0].Event)
	e = events[1]
	require.Equal(t, []byte("bar"), e.Config.Raw)
	require.NotEqual(t, initChecksum, e.Config.Info.Checksum)
	require.Greater(t, e.Config.Info.Time, initTime)
}

func TestGit_EventUrlKeepsSpellingOfConfiguredUrl(t *testing.T) {
	testcases := []struct {
		name string
		cfg  func(repoUrl string) static.GitProvider
		url  func(repoUrl string) string
	}{
		{
			name: "https url",
			cfg:  func(u string) static.GitProvider { return static.GitProvider{Urls: []string{u}} },
			url:  func(u string) string { return u + "?file=%2Fpet.yaml" },
		},
		{
			name: "git+https url",
			cfg:  func(u string) static.GitProvider { return static.GitProvider{Urls: []string{"git+" + u}} },
			url:  func(u string) string { return "git+" + u + "?file=%2Fpet.yaml" },
		},
		{
			name: "git+https repository",
			cfg: func(u string) static.GitProvider {
				return static.GitProvider{Repositories: []static.GitRepo{{Url: "git+" + u}}}
			},
			url: func(u string) string { return "git+" + u + "?file=%2Fpet.yaml" },
		},
		{
			name: "git+https url with branch",
			cfg:  func(u string) static.GitProvider { return static.GitProvider{Urls: []string{"git+" + u + "?ref=main"}} },
			url:  func(u string) string { return "git+" + u + "?file=%2Fpet.yaml&ref=main" },
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s := gittest.NewServer(t)
			s.CreateRepo("schemas", map[string]string{"pet.yaml": "name: Odie"})
			repoUrl := s.Url("schemas")

			cfg := tc.cfg(repoUrl)
			cfg.TempDir = t.TempDir()
			ch := startWithConfig(t, cfg)

			e := nextEvent(t, ch)
			require.Equal(t, tc.url(repoUrl), e.Config.Info.Url.String())
			require.Equal(t, []byte("name: Odie"), e.Config.Raw)
		})
	}
}

func TestGit_InvalidUrlIsSkipped(t *testing.T) {
	// no event is expected, Start must neither fail nor panic
	startWithConfig(t, static.GitProvider{
		Urls:         []string{"://invalid"},
		Repositories: []static.GitRepo{{Url: "git+://invalid"}},
		TempDir:      t.TempDir(),
	})
}

func TestGitSimpleUrl(t *testing.T) {
	s := gittest.NewServer(t)
	s.CreateRepo("simple", map[string]string{"foo.txt": "bar"})

	ch := startWithConfig(t, static.GitProvider{Urls: []string{s.Url("simple")}})

	e := nextEvent(t, ch)
	require.Equal(t, s.Url("simple")+"?file=%2Ffoo.txt", e.Config.Info.Url.String())
}

// file:// URLs are a supported config form, this is the only test that clones such a repository.
func TestGit_FileUrl(t *testing.T) {
	dir := t.TempDir()
	repo, err := git.PlainInit(dir, false)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "foo.txt"), []byte("bar"), 0o600))
	w, err := repo.Worktree()
	require.NoError(t, err)
	_, err = w.Add("foo.txt")
	require.NoError(t, err)
	_, err = w.Commit("added foo.txt", &git.CommitOptions{Author: &object.Signature{Name: "test", When: time.Now()}})
	require.NoError(t, err)
	repoUrl := "file://" + filepath.ToSlash(dir)

	ch := startWithConfig(t, static.GitProvider{Urls: []string{repoUrl}, TempDir: t.TempDir()})

	events := collectEvents(t, ch, 1)
	require.Len(t, events, 1)
	require.Equal(t, repoUrl+"?file=%2Ffoo.txt", events[0].Config.Info.Url.String())
	require.Equal(t, []byte("bar"), events[0].Config.Raw)
}

// Not implemented correctly by go-git https://github.com/go-git/go-git/issues/90
func testGit_SparseUrl(t *testing.T) {
	s := gittest.NewServer(t)
	s.CreateRepo("sparse", map[string]string{"foo/foo.txt": "bar", "bar/bar.txt": "bar"})

	ch := startWithConfig(t, static.GitProvider{Urls: []string{s.Url("sparse") + "//foo"}})

	events := collectEvents(t, ch, 1)
	require.Len(t, events, 1)
	require.Equal(t, "foo.txt", filepath.Base(events[0].Config.Info.Url.String()))
}
