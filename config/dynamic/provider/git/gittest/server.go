// Package gittest provides a local Git server for tests. It runs `git http-backend` behind an
// HTTPS test server, so clients can do what they do against GitHub: shallow clones, fetches and
// pulls. go-git's in-process server is not used because it can't serve shallow clones.
package gittest

import (
	"net/http"
	"net/http/cgi"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

const defaultBranch = "main"

// Server serves bare repositories created with CreateRepo.
type Server struct {
	t    *testing.T
	root string
	srv  *httptest.Server

	m           sync.Mutex
	uploadPacks map[string]int
}

// Repo is the working copy of a repository on the Server. Changes are pushed to the Server
// immediately.
type Repo struct {
	t    *testing.T
	name string
	dir  string
}

// NewServer starts the server and makes http.DefaultTransport trust its certificate until the
// test ends. Tests that use it must not run in parallel.
func NewServer(t *testing.T) *Server {
	t.Helper()

	gitPath, err := exec.LookPath("git")
	if err != nil {
		t.Fatalf("git http-backend not found: %v", err)
	}
	execPath, err := exec.Command(gitPath, "--exec-path").Output()
	if err != nil {
		t.Fatalf("git http-backend not found: %v", err)
	}
	if _, err = os.Stat(filepath.Join(strings.TrimSpace(string(execPath)), "git-http-backend")); err != nil {
		t.Fatalf("git http-backend not found: %v", err)
	}

	s := &Server{
		t:           t,
		root:        t.TempDir(),
		uploadPacks: make(map[string]int),
	}
	backend := &cgi.Handler{
		Path: gitPath,
		Args: []string{"http-backend"},
		Env: []string{
			"GIT_PROJECT_ROOT=" + s.root,
			"GIT_HTTP_EXPORT_ALL=1",
			"GIT_CONFIG_GLOBAL=" + os.DevNull,
			"GIT_CONFIG_NOSYSTEM=1",
		},
	}
	s.srv = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.countUploadPack(r)
		backend.ServeHTTP(w, r)
	}))
	t.Cleanup(s.srv.Close)

	s.trustCertificate()
	return s
}

func (s *Server) trustCertificate() {
	transport := http.DefaultTransport.(*http.Transport)
	original := transport.TLSClientConfig
	transport.TLSClientConfig = s.srv.Client().Transport.(*http.Transport).TLSClientConfig.Clone()
	s.t.Cleanup(func() { transport.TLSClientConfig = original })
}

// countUploadPack counts requests that transfer objects (clone and fetch). Refs advertisement
// requests (GET info/refs) are not counted.
func (s *Server) countUploadPack(r *http.Request) {
	repo, ok := strings.CutSuffix(strings.TrimPrefix(r.URL.Path, "/"), "/git-upload-pack")
	if r.Method != http.MethodPost || !ok {
		return
	}
	s.m.Lock()
	defer s.m.Unlock()
	s.uploadPacks[repo]++
}

// Url is the HTTPS URL of the repository, e.g. https://127.0.0.1:12345/schemas
func (s *Server) Url(repo string) string {
	return s.srv.URL + "/" + repo
}

// CloneCount is the number of times objects were downloaded from the repository. It does not
// change while the repository is unchanged, because a pull without news needs no download.
func (s *Server) CloneCount(repo string) int {
	s.m.Lock()
	defer s.m.Unlock()
	return s.uploadPacks[repo]
}

// CreateRepo creates the repository with one commit on branch main.
func (s *Server) CreateRepo(name string, files map[string]string) *Repo {
	s.t.Helper()

	bare := filepath.Join(s.root, name)
	r := &Repo{t: s.t, name: name, dir: s.t.TempDir()}
	run(s.t, "", "init", "--bare", "-b", defaultBranch, bare)
	run(s.t, r.dir, "init", "-b", defaultBranch)
	run(s.t, r.dir, "remote", "add", "origin", bare)
	r.Commit(files)
	return r
}

// Commit writes the files and pushes a commit to the current branch. Files with the same name
// are overwritten.
func (r *Repo) Commit(files map[string]string) {
	r.t.Helper()

	for name, content := range files {
		path := filepath.Join(r.dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			r.t.Fatalf("create directory for %v: %v", name, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			r.t.Fatalf("write %v: %v", name, err)
		}
	}
	r.commitAndPush()
}

// Remove deletes the file and pushes a commit to the current branch.
func (r *Repo) Remove(file string) {
	r.t.Helper()

	if err := os.Remove(filepath.Join(r.dir, filepath.FromSlash(file))); err != nil {
		r.t.Fatalf("remove %v: %v", file, err)
	}
	r.commitAndPush()
}

// CreateBranch creates the branch from the current one, switches to it and pushes it. Later
// commits go to this branch until Checkout is called.
func (r *Repo) CreateBranch(name string) {
	r.t.Helper()

	run(r.t, r.dir, "checkout", "-b", name)
	run(r.t, r.dir, "push", "origin", name)
}

// Checkout switches the branch that later commits go to.
func (r *Repo) Checkout(name string) {
	r.t.Helper()
	run(r.t, r.dir, "checkout", name)
}

func (r *Repo) commitAndPush() {
	r.t.Helper()

	run(r.t, r.dir, "add", "--all")
	run(r.t, r.dir, "-c", "user.name=gittest", "-c", "user.email=gittest@mokapi.io",
		"commit", "--allow-empty", "-m", "test commit")
	run(r.t, r.dir, "push", "origin", "HEAD")
}

func run(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", strings.Join(args, " "), err, out)
	}
}
