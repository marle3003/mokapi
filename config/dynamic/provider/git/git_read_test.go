package git

import (
	"context"
	"errors"
	"io"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/provider/file"
	"mokapi/config/dynamic/provider/git/gittest"
	"mokapi/config/static"
	"mokapi/safe"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func mustParse(t *testing.T, s string) *url.URL {
	t.Helper()
	u, err := url.Parse(s)
	require.NoError(t, err)
	return u
}

func TestCanonicalUrl(t *testing.T) {
	testcases := []struct {
		name     string
		url      string
		expected string
	}{
		{"file without slash", "git+https://host/org/repo?file=pet.yaml", "git+https://host/org/repo?file=%2Fpet.yaml"},
		{"file with slash", "git+https://host/org/repo?file=/pet.yaml", "git+https://host/org/repo?file=%2Fpet.yaml"},
		{"file encoded", "git+https://host/org/repo?file=%2Fpet.yaml", "git+https://host/org/repo?file=%2Fpet.yaml"},
		{"query is sorted", "git+https://host/org/repo?ref=main&file=pet.yaml", "git+https://host/org/repo?file=%2Fpet.yaml&ref=main"},
		{"already canonical", "git+https://host/org/repo?file=%2Fpet.yaml&ref=main", "git+https://host/org/repo?file=%2Fpet.yaml&ref=main"},
		{"without file", "git+https://host/org/repo?ref=main", "git+https://host/org/repo?ref=main"},
		{"https is unchanged", "https://host/org/repo?ref=main&file=pet.yaml", "https://host/org/repo?ref=main&file=pet.yaml"},
		{"file scheme is unchanged", "file:/tmp/a.yaml", "file:/tmp/a.yaml"},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, CanonicalUrl(mustParse(t, tc.url)).String())
		})
	}
}

func TestParseReference(t *testing.T) {
	testcases := []struct {
		name      string
		url       string
		cloneUrl  string
		branch    string
		file      string
		canonical string
		err       string
	}{
		{
			name: "default branch", url: "git+https://host/org/repo?file=pet.yaml",
			cloneUrl: "https://host/org/repo", file: "/pet.yaml",
			canonical: "git+https://host/org/repo?file=%2Fpet.yaml",
		},
		{
			name: "branch and nested file", url: "git+https://host/org/repo.git?ref=develop&file=/a/b.yaml",
			cloneUrl: "https://host/org/repo.git", branch: "develop", file: "/a/b.yaml",
			canonical: "git+https://host/org/repo.git?file=%2Fa%2Fb.yaml&ref=develop",
		},
		{
			name: "branch with slash", url: "git+https://host/repo?ref=feature%2Fx&file=a.yaml",
			cloneUrl: "https://host/repo", branch: "feature/x", file: "/a.yaml",
			canonical: "git+https://host/repo?file=%2Fa.yaml&ref=feature%2Fx",
		},
		{
			name: "file is cleaned", url: "git+https://host/repo?file=./a//b/./c.yaml",
			cloneUrl: "https://host/repo", file: "/a/b/c.yaml",
			canonical: "git+https://host/repo?file=%2F.%2Fa%2F%2Fb%2F.%2Fc.yaml",
		},
		{name: "file is required", url: "git+https://host/repo?ref=main", err: "git URL requires query parameter 'file'"},
		{name: "file is not empty", url: "git+https://host/repo?file=", err: "git URL requires query parameter 'file'"},
		{name: "parent directory", url: "git+https://host/repo?file=../a.yaml", err: "must not leave the repository"},
		{name: "parent directory in the middle", url: "git+https://host/repo?file=a/../../b.yaml", err: "must not leave the repository"},
		{name: "parent directory that stays inside", url: "git+https://host/repo?file=a/../b.yaml", err: "must not leave the repository"},
		{name: "git directory", url: "git+https://host/repo?file=.git/config", err: "must not point into .git"},
		{name: "git directory with slash", url: "git+https://host/repo?file=/.git/config", err: "must not point into .git"},
		{name: "git directory by dots", url: "git+https://host/repo?file=a/./../.git/config", err: "must not leave the repository"},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ref, err := parseReference(mustParse(t, tc.url))
			if tc.err != "" {
				require.ErrorContains(t, err, tc.err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.cloneUrl, ref.cloneUrl.String())
			require.Equal(t, tc.branch, ref.branch)
			require.Equal(t, tc.file, ref.file)
			require.Equal(t, tc.canonical, ref.canonical.String())
		})
	}
}

func TestRepositoryKey(t *testing.T) {
	testcases := []struct {
		name     string
		url      string
		ref      string
		expected string
	}{
		{"plain", "https://host/org/repo", "", "host/org/repo?ref="},
		{"git suffix", "https://host/org/repo.git", "", "host/org/repo?ref="},
		{"trailing slash", "https://host/org/repo/", "", "host/org/repo?ref="},
		{"branch", "https://host/org/repo", "main", "host/org/repo?ref=main"},
		{"port", "https://127.0.0.1:8443/repo", "", "127.0.0.1:8443/repo?ref="},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, repositoryKey(mustParse(t, tc.url), tc.ref))
		})
	}
}

func TestProvider_LookupOfConfiguredRepository(t *testing.T) {
	logrus.SetOutput(io.Discard)
	g := New(static.GitProvider{
		TempDir: t.TempDir(),
		Urls:    []string{"git+https://host/org/repo.git?ref=develop", "https://host/org/other"},
	})

	require.Contains(t, g.repos, "host/org/repo?ref=develop")
	require.Contains(t, g.repos, "host/org/other?ref=")
	require.NotContains(t, g.repos, "host/org/repo?ref=")

	// a branch that is not configured is cloned from the URL of the configured repository
	got := g.configuredRepository(mustParse(t, "https://host/org/repo/"))
	require.NotNil(t, got)
	require.Equal(t, "https://host/org/repo.git", got.repoUrl)
	require.Nil(t, g.configuredRepository(mustParse(t, "https://host/org/unknown")))
}

func TestRepository_LocalFile(t *testing.T) {
	outside := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(outside, "secret.yaml"), []byte("secret"), 0o600))

	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "dir"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(root, ".git"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "dir", "a.yaml"), []byte("a"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root, ".git", "config"), []byte("c"), 0o600))
	require.NoError(t, os.Symlink(filepath.Join(root, "dir", "a.yaml"), filepath.Join(root, "inside-link.yaml")))
	require.NoError(t, os.Symlink(filepath.Join(outside, "secret.yaml"), filepath.Join(root, "outside-link.yaml")))
	require.NoError(t, os.Symlink(outside, filepath.Join(root, "outside-dir")))
	require.NoError(t, os.Symlink(filepath.Join(root, ".git", "config"), filepath.Join(root, "git-link")))
	r := &repository{localPath: root, repoUrl: "https://host/repo"}

	testcases := []struct {
		name string
		file string
		err  string
	}{
		{name: "file", file: "/dir/a.yaml"},
		{name: "symlink inside the clone", file: "/inside-link.yaml"},
		{name: "symlink to a file outside", file: "/outside-link.yaml", err: "outside of the repository"},
		{name: "symlink to a directory outside", file: "/outside-dir/secret.yaml", err: "outside of the repository"},
		{name: "symlink into .git", file: "/git-link", err: "outside of the repository"},
		{name: "missing file", file: "/dir/missing.yaml", err: "file '/dir/missing.yaml' not found in git repository https://host/repo"},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			local, err := r.localFile(tc.file)
			if tc.err != "" {
				require.ErrorContains(t, err, tc.err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, filepath.Join(root, filepath.FromSlash(tc.file)), local)
		})
	}
}

// startProvider starts the provider with a pull interval of one second. Without drain the test reads
// the events of the provider from the returned channel.
func startProvider(t *testing.T, cfg static.GitProvider, drain bool) (*Provider, chan dynamic.ConfigEvent, *safe.Pool) {
	t.Helper()
	logrus.SetOutput(io.Discard)

	if cfg.TempDir == "" {
		cfg.TempDir = t.TempDir()
	}
	if cfg.PullInterval == "" {
		cfg.PullInterval = "1s"
	}
	g := New(cfg)
	pool := safe.NewPool(context.Background())
	t.Cleanup(pool.Stop)
	ch := make(chan dynamic.ConfigEvent)
	require.NoError(t, g.Start(ch, pool))

	if drain {
		done := make(chan struct{})
		t.Cleanup(func() { close(done) })
		go func() {
			for {
				select {
				case <-ch:
				case <-done:
					return
				}
			}
		}()
	}
	return g, ch, pool
}

func gitUrl(s *gittest.Server, repo, query string) *url.URL {
	u, _ := url.Parse("git+" + s.Url(repo) + "?" + query)
	return u
}

const petYaml = "name: Odie"

func TestProvider_ReadConfiguredRepository(t *testing.T) {
	t.Run("wrapped config of the file in the clone", func(t *testing.T) {
		s := gittest.NewServer(t)
		s.CreateRepo("schemas", map[string]string{"pet.yaml": petYaml, "dir/a.yaml": "a"})
		g, _, _ := startProvider(t, static.GitProvider{Urls: []string{s.Url("schemas")}}, true)

		c, err := g.Read(gitUrl(s, "schemas", "file=pet.yaml"))
		require.NoError(t, err)
		require.Equal(t, []byte(petYaml), c.Raw)
		require.Equal(t, "git", c.Info.Provider)
		require.Equal(t, "git+"+s.Url("schemas")+"?file=%2Fpet.yaml", c.Info.Url.String())
		require.Equal(t, []string{"main"}, c.Info.Tags)
		require.Equal(t, dynamic.SourceReference, c.SourceType)
		require.NotEmpty(t, c.Info.Checksum)

		inner := c.Info.Inner()
		require.Equal(t, "file", inner.Url.Scheme)
		require.Equal(t, "pet.yaml", filepath.Base(inner.Url.Path))
		require.True(t, filepath.IsAbs(inner.Url.Path))
		require.Equal(t, 1, s.CloneCount("schemas"), "the configured clone is reused")
	})

	t.Run("files and include of the repository do not restrict", func(t *testing.T) {
		s := gittest.NewServer(t)
		s.CreateRepo("schemas", map[string]string{"pet.yaml": petYaml, "README.md": "#", "dir/b.yaml": "b"})
		g, _, _ := startProvider(t, static.GitProvider{Repositories: []static.GitRepo{
			{Url: s.Url("schemas"), Files: []string{"README.md"}, Include: []string{"*.md"}},
		}}, true)

		c, err := g.Read(gitUrl(s, "schemas", "file=pet.yaml"))
		require.NoError(t, err)
		require.Equal(t, []byte(petYaml), c.Raw)
		c, err = g.Read(gitUrl(s, "schemas", "file=dir/b.yaml"))
		require.NoError(t, err)
		require.Equal(t, []byte("b"), c.Raw)
	})

	t.Run("git+https spelling in the configuration", func(t *testing.T) {
		s := gittest.NewServer(t)
		s.CreateRepo("schemas", map[string]string{"pet.yaml": petYaml})
		g, _, _ := startProvider(t, static.GitProvider{Urls: []string{"git+" + s.Url("schemas") + "?ref=main"}}, true)

		c, err := g.Read(gitUrl(s, "schemas", "file=pet.yaml&ref=main"))
		require.NoError(t, err)
		require.Equal(t, []byte(petYaml), c.Raw)
		require.Equal(t, 1, s.CloneCount("schemas"))
	})

	t.Run("failed clone is the error of every read", func(t *testing.T) {
		s := gittest.NewServer(t)
		g, _, _ := startProvider(t, static.GitProvider{Urls: []string{s.Url("unknown")}}, true)

		_, err := g.Read(gitUrl(s, "unknown", "file=pet.yaml"))
		require.ErrorContains(t, err, "repository not found")
		_, err2 := g.Read(gitUrl(s, "unknown", "file=pet.yaml"))
		require.Equal(t, err, err2)
	})

	t.Run("errors of the URL", func(t *testing.T) {
		s := gittest.NewServer(t)
		s.CreateRepo("schemas", map[string]string{"pet.yaml": petYaml})
		g, _, _ := startProvider(t, static.GitProvider{Urls: []string{s.Url("schemas")}}, true)

		_, err := g.Read(gitUrl(s, "schemas", ""))
		require.EqualError(t, err, "git URL requires query parameter 'file'")
		_, err = g.Read(gitUrl(s, "schemas", "file=../x.yaml"))
		require.ErrorContains(t, err, "must not leave the repository")
		_, err = g.Read(gitUrl(s, "schemas", "file=missing.yaml"))
		require.ErrorContains(t, err, "not found in git repository")
	})
}

func TestProvider_ReadWaitsForClone(t *testing.T) {
	logrus.SetOutput(io.Discard)

	// newWaiting returns a provider with a repository whose clone is not done
	newWaiting := func(t *testing.T) (*Provider, *repository, *safe.Pool) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "pet.yaml"), []byte(petYaml), 0o600))
		repo, err := git.PlainInit(dir, false)
		require.NoError(t, err)

		g := New(static.GitProvider{TempDir: t.TempDir()})
		r := &repository{
			key: "host/repo?ref=", localPath: dir, repoUrl: "https://host/repo", repo: repo,
			ready: make(chan struct{}), files: file.New(static.FileProvider{}),
		}
		pool := safe.NewPool(context.Background())
		t.Cleanup(pool.Stop)
		require.NoError(t, g.Start(make(chan dynamic.ConfigEvent), pool))
		// added after Start, otherwise Start clones it
		g.m.Lock()
		g.repos[r.key] = r
		g.m.Unlock()
		return g, r, pool
	}

	read := func(g *Provider) chan error {
		done := make(chan error, 1)
		go func() {
			c, err := g.Read(mustParse(t, "git+https://host/repo?file=pet.yaml"))
			if err == nil && string(c.Raw) != petYaml {
				err = errors.New("unexpected content")
			}
			done <- err
		}()
		return done
	}

	t.Run("read returns when the clone is done", func(t *testing.T) {
		g, r, _ := newWaiting(t)
		done := read(g)

		select {
		case err := <-done:
			t.Fatalf("read must wait for the clone, got %v", err)
		case <-time.After(200 * time.Millisecond):
		}

		close(r.ready)
		require.NoError(t, <-done)
	})

	t.Run("read fails with the error of the clone", func(t *testing.T) {
		g, r, _ := newWaiting(t)
		done := read(g)

		r.err = errors.New("clone failed")
		close(r.ready)
		require.EqualError(t, <-done, "clone failed")
	})

	t.Run("read stops when the provider stops", func(t *testing.T) {
		g, _, pool := newWaiting(t)
		done := read(g)

		pool.Stop()
		require.ErrorIs(t, <-done, context.Canceled)
	})
}

func TestProvider_ReadWithoutStart(t *testing.T) {
	g := New(static.GitProvider{})
	_, err := g.Read(mustParse(t, "git+https://host/repo?file=pet.yaml"))
	require.EqualError(t, err, "git provider not started")
}

func tempDirEntries(t *testing.T, dir string) int {
	t.Helper()
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	return len(entries)
}

func TestProvider_ReadOnDemandRepository(t *testing.T) {
	t.Run("clone of the default branch", func(t *testing.T) {
		s := gittest.NewServer(t)
		s.CreateRepo("schemas", map[string]string{"pet.yaml": petYaml})
		tempDir := t.TempDir()
		g, _, _ := startProvider(t, static.GitProvider{TempDir: tempDir}, true)

		c, err := g.Read(gitUrl(s, "schemas", "file=pet.yaml"))
		require.NoError(t, err)
		require.Equal(t, []byte(petYaml), c.Raw)
		require.Equal(t, []string{"main"}, c.Info.Tags)
		require.Equal(t, 1, s.CloneCount("schemas"))
		require.Equal(t, 1, tempDirEntries(t, tempDir))
		require.True(t, g.repos["127.0.0.1:"+portOf(t, s)+"/schemas?ref="].onDemand)
	})

	t.Run("one clone per repository and branch", func(t *testing.T) {
		s := gittest.NewServer(t)
		repo := s.CreateRepo("schemas", map[string]string{"pet.yaml": petYaml, "other.yaml": "other"})
		repo.CreateBranch("develop")
		repo.Commit(map[string]string{"pet.yaml": "name: Garfield"})
		g, _, _ := startProvider(t, static.GitProvider{}, true)

		var wg sync.WaitGroup
		for _, query := range []string{"file=pet.yaml", "file=other.yaml", "file=/pet.yaml&ref=main", "ref=main&file=pet.yaml"} {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := g.Read(gitUrl(s, "schemas", query))
				require.NoError(t, err)
			}()
		}
		wg.Wait()
		require.Equal(t, 2, s.CloneCount("schemas"), "default branch and main are different keys")

		c, err := g.Read(gitUrl(s, "schemas", "file=pet.yaml&ref=develop"))
		require.NoError(t, err)
		require.Equal(t, []byte("name: Garfield"), c.Raw)
		require.Equal(t, []string{"develop"}, c.Info.Tags)
		require.Equal(t, 3, s.CloneCount("schemas"))
	})

	t.Run("configured clone of another branch is not reused", func(t *testing.T) {
		s := gittest.NewServer(t)
		repo := s.CreateRepo("schemas", map[string]string{"pet.yaml": petYaml})
		repo.CreateBranch("develop")
		repo.Commit(map[string]string{"pet.yaml": "name: Garfield"})
		g, _, _ := startProvider(t, static.GitProvider{Urls: []string{s.Url("schemas") + "?ref=develop"}}, true)

		c, err := g.Read(gitUrl(s, "schemas", "file=pet.yaml&ref=develop"))
		require.NoError(t, err)
		require.Equal(t, []byte("name: Garfield"), c.Raw)
		require.Equal(t, 1, s.CloneCount("schemas"))

		c, err = g.Read(gitUrl(s, "schemas", "file=pet.yaml"))
		require.NoError(t, err)
		require.Equal(t, []byte(petYaml), c.Raw)
		require.Equal(t, 2, s.CloneCount("schemas"))
	})

	t.Run("pull interval of the configured repository or the global one", func(t *testing.T) {
		s := gittest.NewServer(t)
		for _, name := range []string{"configured", "other"} {
			repo := s.CreateRepo(name, map[string]string{"pet.yaml": petYaml})
			repo.CreateBranch("develop")
			repo.Checkout("main")
		}
		g, _, _ := startProvider(t, static.GitProvider{
			PullInterval: "11s",
			Repositories: []static.GitRepo{{Url: s.Url("configured"), PullInterval: "7s"}},
		}, true)

		for _, query := range []string{"file=pet.yaml&ref=develop", "file=pet.yaml"} {
			_, err := g.Read(gitUrl(s, "configured", query))
			require.NoError(t, err)
		}
		_, err := g.Read(gitUrl(s, "other", "file=pet.yaml"))
		require.NoError(t, err)

		g.m.Lock()
		defer g.m.Unlock()
		require.Equal(t, "7s", g.repos["127.0.0.1:"+portOf(t, s)+"/configured?ref=develop"].config.PullInterval)
		require.Equal(t, "11s", g.repos["127.0.0.1:"+portOf(t, s)+"/other?ref="].config.PullInterval)
	})

	t.Run("no clone after the provider stopped", func(t *testing.T) {
		s := gittest.NewServer(t)
		s.CreateRepo("schemas", map[string]string{"pet.yaml": petYaml})
		g, _, pool := startProvider(t, static.GitProvider{}, true)
		pool.Stop()

		_, err := g.Read(gitUrl(s, "schemas", "file=pet.yaml"))
		require.ErrorIs(t, err, context.Canceled)
		require.Equal(t, 0, s.CloneCount("schemas"))
		require.Empty(t, g.repos)
	})

	t.Run("failed clone is removed and tried again", func(t *testing.T) {
		s := gittest.NewServer(t)
		tempDir := t.TempDir()
		g, _, _ := startProvider(t, static.GitProvider{TempDir: tempDir}, true)

		_, err := g.Read(gitUrl(s, "late", "file=pet.yaml"))
		require.ErrorContains(t, err, "repository not found")
		require.Empty(t, g.repos)
		require.Equal(t, 0, tempDirEntries(t, tempDir))

		s.CreateRepo("late", map[string]string{"pet.yaml": petYaml})
		c, err := g.Read(gitUrl(s, "late", "file=pet.yaml"))
		require.NoError(t, err)
		require.Equal(t, []byte(petYaml), c.Raw)
	})

	t.Run("unknown branch", func(t *testing.T) {
		s := gittest.NewServer(t)
		s.CreateRepo("schemas", map[string]string{"pet.yaml": petYaml})
		g, _, _ := startProvider(t, static.GitProvider{}, true)

		_, err := g.Read(gitUrl(s, "schemas", "file=pet.yaml&ref=nothing"))
		require.ErrorContains(t, err, "refs/heads/nothing")
		require.Empty(t, g.repos)
	})

	t.Run("pull with the global interval", func(t *testing.T) {
		s := gittest.NewServer(t)
		repo := s.CreateRepo("schemas", map[string]string{"pet.yaml": petYaml})
		g, _, _ := startProvider(t, static.GitProvider{PullInterval: "1s"}, true)
		_, err := g.Read(gitUrl(s, "schemas", "file=pet.yaml"))
		require.NoError(t, err)

		repo.Commit(map[string]string{"pet.yaml": "name: Garfield"})

		require.Eventually(t, func() bool {
			c, err := g.Read(gitUrl(s, "schemas", "file=pet.yaml"))
			return err == nil && string(c.Raw) == "name: Garfield"
		}, 10*time.Second, 100*time.Millisecond)
	})

	t.Run("pull of a branch", func(t *testing.T) {
		s := gittest.NewServer(t)
		repo := s.CreateRepo("schemas", map[string]string{"pet.yaml": petYaml})
		repo.CreateBranch("develop")
		repo.Commit(map[string]string{"pet.yaml": "name: Garfield"})
		g, _, _ := startProvider(t, static.GitProvider{PullInterval: "1s"}, true)
		_, err := g.Read(gitUrl(s, "schemas", "file=pet.yaml&ref=develop"))
		require.NoError(t, err)

		repo.Commit(map[string]string{"pet.yaml": "name: Heathcliff"})

		require.Eventually(t, func() bool {
			c, err := g.Read(gitUrl(s, "schemas", "file=pet.yaml&ref=develop"))
			return err == nil && string(c.Raw) == "name: Heathcliff"
		}, 10*time.Second, 100*time.Millisecond)
	})

	t.Run("clones are removed when the provider stops", func(t *testing.T) {
		s := gittest.NewServer(t)
		s.CreateRepo("schemas", map[string]string{"pet.yaml": petYaml})
		tempDir := t.TempDir()
		g, _, pool := startProvider(t, static.GitProvider{TempDir: tempDir}, true)
		_, err := g.Read(gitUrl(s, "schemas", "file=pet.yaml"))
		require.NoError(t, err)
		require.Equal(t, 1, tempDirEntries(t, tempDir))

		pool.Stop()

		require.Equal(t, 0, tempDirEntries(t, tempDir))
	})
}

func portOf(t *testing.T, s *gittest.Server) string {
	t.Helper()
	return mustParse(t, s.Url("x")).Port()
}

// waitForEvents collects the events of ch until done is true for an event.
func waitForEvents(t *testing.T, ch chan dynamic.ConfigEvent, done func(e dynamic.ConfigEvent) bool) []dynamic.ConfigEvent {
	t.Helper()
	var events []dynamic.ConfigEvent
	timeout := time.After(15 * time.Second)
	for {
		select {
		case e := <-ch:
			events = append(events, e)
			if done(e) {
				return events
			}
		case <-timeout:
			t.Fatalf("timeout, events: %v", len(events))
		}
	}
}

// updated is true for the event of a changed file with the content.
func updated(content string) func(e dynamic.ConfigEvent) bool {
	return func(e dynamic.ConfigEvent) bool {
		return e.Event != dynamic.Delete && string(e.Config.Raw) == content
	}
}

func eventUrls(events []dynamic.ConfigEvent) []string {
	var urls []string
	for _, e := range events {
		if e.Event == dynamic.Delete {
			urls = append(urls, "delete "+e.Name)
		} else {
			urls = append(urls, e.Config.Info.Url.String())
		}
	}
	return urls
}

// skipInitialEvents reads the events of the configured repository that come from loading its files.
func skipInitialEvents(t *testing.T, ch chan dynamic.ConfigEvent, files int) {
	t.Helper()
	for i := 0; i < files; i++ {
		select {
		case <-ch:
		case <-time.After(10 * time.Second):
			t.Fatal("timeout waiting for the initial events")
		}
	}
}

func TestProvider_EventsOfReferencedFiles(t *testing.T) {
	t.Run("one event per URL with its own config", func(t *testing.T) {
		s := gittest.NewServer(t)
		repo := s.CreateRepo("schemas", map[string]string{"pet.yaml": petYaml})
		g, ch, _ := startProvider(t, static.GitProvider{Urls: []string{s.Url("schemas")}}, false)
		skipInitialEvents(t, ch, 1)
		_, err := g.Read(gitUrl(s, "schemas", "file=pet.yaml"))
		require.NoError(t, err)

		repo.Commit(map[string]string{"pet.yaml": "name: Garfield"})

		var changed []dynamic.ConfigEvent
		waitForEvents(t, ch, func(e dynamic.ConfigEvent) bool {
			if updated("name: Garfield")(e) {
				changed = append(changed, e)
			}
			return len(changed) == 2
		})
		require.Len(t, changed, 2)
		require.ElementsMatch(t, []string{
			s.Url("schemas") + "?file=%2Fpet.yaml",
			"git+" + s.Url("schemas") + "?file=%2Fpet.yaml",
		}, eventUrls(changed))
		require.NotSame(t, changed[0].Config, changed[1].Config)
	})

	t.Run("no second event for the URL of the provider", func(t *testing.T) {
		s := gittest.NewServer(t)
		repo := s.CreateRepo("schemas", map[string]string{"pet.yaml": petYaml})
		g, ch, _ := startProvider(t, static.GitProvider{Urls: []string{"git+" + s.Url("schemas")}}, false)
		skipInitialEvents(t, ch, 1)
		_, err := g.Read(gitUrl(s, "schemas", "file=/pet.yaml"))
		require.NoError(t, err)

		repo.Commit(map[string]string{"pet.yaml": "name: Garfield"})

		events := waitForEvents(t, ch, updated("name: Garfield"))
		select {
		case e := <-ch:
			t.Fatalf("unexpected second event %v", eventUrls([]dynamic.ConfigEvent{e}))
		case <-time.After(1500 * time.Millisecond):
		}
		for _, e := range events {
			if e.Event != dynamic.Delete {
				require.Equal(t, "git+"+s.Url("schemas")+"?file=%2Fpet.yaml", e.Config.Info.Url.String())
			}
		}
	})

	t.Run("referenced file that the repository skips", func(t *testing.T) {
		s := gittest.NewServer(t)
		repo := s.CreateRepo("schemas", map[string]string{"pet.yaml": petYaml, "README.md": "#"})
		g, ch, _ := startProvider(t, static.GitProvider{Repositories: []static.GitRepo{
			{Url: s.Url("schemas"), Files: []string{"README.md"}},
		}}, false)
		skipInitialEvents(t, ch, 1)
		_, err := g.Read(gitUrl(s, "schemas", "file=pet.yaml"))
		require.NoError(t, err)

		repo.Commit(map[string]string{"pet.yaml": "name: Garfield"})

		events := waitForEvents(t, ch, updated("name: Garfield"))
		require.Equal(t, "git+"+s.Url("schemas")+"?file=%2Fpet.yaml", events[len(events)-1].Config.Info.Url.String())
		own := s.Url("schemas") + "?file=%2Fpet.yaml"
		require.NotContains(t, eventUrls(events), own, "file is not part of the configured repository")
		require.NotContains(t, eventUrls(events), "delete "+own)
	})

	t.Run("delete of a referenced file is named by the URL", func(t *testing.T) {
		s := gittest.NewServer(t)
		repo := s.CreateRepo("schemas", map[string]string{"pet.yaml": petYaml, "README.md": "#"})
		g, ch, _ := startProvider(t, static.GitProvider{Urls: []string{s.Url("schemas")}}, false)
		skipInitialEvents(t, ch, 2)
		_, err := g.Read(gitUrl(s, "schemas", "file=pet.yaml"))
		require.NoError(t, err)

		repo.Remove("pet.yaml")

		var names []string
		waitForEvents(t, ch, func(e dynamic.ConfigEvent) bool {
			if e.Event == dynamic.Delete {
				names = append(names, e.Name)
			}
			return len(names) == 2
		})
		// the file of the repository keeps the path as name, the config watcher ignores it
		require.Len(t, names, 2)
		require.Contains(t, names, "git+"+s.Url("schemas")+"?file=%2Fpet.yaml")
		require.NotContains(t, names, s.Url("schemas")+"?file=%2Fpet.yaml")
	})

	t.Run("on-demand clone sends referenced files only", func(t *testing.T) {
		s := gittest.NewServer(t)
		repo := s.CreateRepo("schemas", map[string]string{"pet.yaml": petYaml, "other.yaml": "other"})
		g, ch, _ := startProvider(t, static.GitProvider{}, false)
		_, err := g.Read(gitUrl(s, "schemas", "file=pet.yaml"))
		require.NoError(t, err)

		repo.Commit(map[string]string{"other.yaml": "changed", "pet.yaml": "name: Garfield"})

		events := waitForEvents(t, ch, updated("name: Garfield"))
		for _, u := range eventUrls(events) {
			require.Contains(t, u, "pet.yaml")
		}
		require.Equal(t, "git+"+s.Url("schemas")+"?file=%2Fpet.yaml", events[len(events)-1].Config.Info.Url.String())
		select {
		case e := <-ch:
			require.Contains(t, eventUrls([]dynamic.ConfigEvent{e})[0], "pet.yaml")
		case <-time.After(1500 * time.Millisecond):
		}
	})
}

func TestProvider_SameRepositoryConfiguredSeveralTimes(t *testing.T) {
	t.Run("one clone with the files of all entries", func(t *testing.T) {
		s := gittest.NewServer(t)
		s.CreateRepo("schemas", map[string]string{"a.yaml": "a", "b.yaml": "b", "c.yaml": "c"})
		g, ch, _ := startProvider(t, static.GitProvider{
			Urls: []string{s.Url("schemas")},
			Repositories: []static.GitRepo{
				{Url: s.Url("schemas"), Files: []string{"a.yaml"}},
				{Url: "git+" + s.Url("schemas") + "/", Include: []string{"b.*"}},
			},
		}, false)

		// the entry of urls has no filter, so it allows all files
		urls := map[string]bool{}
		timeout := time.After(10 * time.Second)
		for len(urls) < 3 {
			select {
			case e := <-ch:
				urls[e.Config.Info.Url.String()] = true
			case <-timeout:
				t.Fatalf("timeout, got %v", urls)
			}
		}
		require.Equal(t, 1, s.CloneCount("schemas"))
		require.Len(t, g.repos, 1)
	})

	t.Run("a file is loaded if any entry allows it", func(t *testing.T) {
		s := gittest.NewServer(t)
		s.CreateRepo("schemas", map[string]string{"a.yaml": "a", "b.yaml": "b", "c.yaml": "c"})
		_, ch, _ := startProvider(t, static.GitProvider{Repositories: []static.GitRepo{
			{Url: s.Url("schemas"), Files: []string{"a.yaml"}},
			{Url: "git+" + s.Url("schemas") + "/", Include: []string{"b.*"}},
		}}, false)

		var urls []string
		timeout := time.After(3 * time.Second)
	Collect:
		for {
			select {
			case e := <-ch:
				urls = append(urls, e.Config.Info.Url.String())
			case <-timeout:
				break Collect
			}
		}
		require.ElementsMatch(t, []string{
			s.Url("schemas") + "?file=%2Fa.yaml",
			s.Url("schemas") + "?file=%2Fb.yaml",
		}, urls, "c.yaml is allowed by no entry, the URLs have the spelling of the first entry")
		require.Equal(t, 1, s.CloneCount("schemas"))
	})

	t.Run("auth and pull interval of the first entry that sets them", func(t *testing.T) {
		logrus.SetOutput(io.Discard)
		hook := logrustest.NewGlobal()
		t.Cleanup(hook.Reset)

		g := New(static.GitProvider{
			PullInterval: "11s",
			Repositories: []static.GitRepo{
				{Url: "https://host/org/repo"},
				{Url: "git+https://host/org/repo.git", PullInterval: "5s"},
				{Url: "https://host/org/repo", PullInterval: "9s"},
				{Url: "https://host/org/other"},
			},
		})

		require.Len(t, g.repos, 2)
		require.Equal(t, "5s", g.repos["host/org/repo?ref="].config.PullInterval)
		require.Equal(t, "11s", g.repos["host/org/other?ref="].config.PullInterval)
		var warnings []string
		for _, e := range hook.AllEntries() {
			if e.Level == logrus.WarnLevel {
				warnings = append(warnings, e.Message)
			}
		}
		require.Len(t, warnings, 1)
		require.Contains(t, warnings[0], "different pull intervals (5s, 9s)")
	})

	t.Run("different branches are different repositories", func(t *testing.T) {
		g := New(static.GitProvider{Urls: []string{"https://host/org/repo", "https://host/org/repo?ref=develop"}})
		require.Len(t, g.repos, 2)
	})
}
