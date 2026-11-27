package git

import (
	"context"
	"errors"
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/provider/file"
	"mokapi/config/static"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

const refScheme = "git+https"

// reference is a parsed git+https://<host>/<repo>?ref=<branch>&file=<path> URL.
type reference struct {
	cloneUrl *url.URL
	branch   string
	// file is the cleaned path of the file in the repository, starting with a slash
	file string
	// canonical is the URL a config of the file is stored by
	canonical *url.URL
}

// CanonicalUrl returns the URL a git+https config is stored by: the query is sorted and file starts
// with a slash, so every spelling of the same reference is one config. Other URLs are returned as
// they are.
func CanonicalUrl(u *url.URL) *url.URL {
	if u.Scheme != refScheme {
		return u
	}

	c := *u
	q := u.Query()
	if file := q.Get("file"); file != "" {
		q.Set("file", "/"+strings.TrimPrefix(file, "/"))
	}
	c.RawQuery = q.Encode()
	return &c
}

func parseReference(u *url.URL) (*reference, error) {
	q := u.Query()
	file, err := cleanFilePath(q.Get("file"))
	if err != nil {
		return nil, err
	}

	return &reference{
		cloneUrl:  &url.URL{Scheme: strings.TrimPrefix(u.Scheme, "git+"), User: u.User, Host: u.Host, Path: u.Path},
		branch:    q.Get("ref"),
		file:      file,
		canonical: CanonicalUrl(u),
	}, nil
}

// cleanFilePath checks that the file is inside the repository and returns it as a path starting
// with a slash. The check for '..' has to come first, path.Clean would silently resolve it.
func cleanFilePath(file string) (string, error) {
	if file == "" {
		return "", errors.New("git URL requires query parameter 'file'")
	}
	for _, segment := range strings.Split(filepath.ToSlash(file), "/") {
		if segment == ".." {
			return "", fmt.Errorf("git URL parameter 'file' must not leave the repository: %v", file)
		}
	}
	p := path.Clean("/" + filepath.ToSlash(file))
	if p == "/.git" || strings.HasPrefix(p, "/.git/") {
		return "", fmt.Errorf("git URL parameter 'file' must not point into .git: %v", file)
	}
	return p, nil
}

// Read reads a file of a Git repository. The repository is the clone of a configured repository or
// a shallow clone made for this file. The config is a reference of the file in the clone, so
// it is parsed again when the file changes.
func (p *Provider) Read(u *url.URL) (*dynamic.Config, error) {
	ref, err := parseReference(u)
	if err != nil {
		return nil, err
	}

	r, err := p.repository(ref)
	if err != nil {
		return nil, err
	}

	local, err := r.localFile(ref.file)
	if err != nil {
		return nil, err
	}
	fileUrl, err := file.ParseUrl(local)
	if err != nil {
		return nil, err
	}
	c, err := r.files.Read(fileUrl)
	if err != nil {
		return nil, err
	}

	wrapConfig(c, r, ref.canonical)
	r.addReference(local, ref.canonical)
	return c, nil
}

// repository returns the clone of the referenced repository and branch, once it is ready.
func (p *Provider) repository(ref *reference) (*repository, error) {
	p.m.Lock()
	if p.pool == nil {
		p.m.Unlock()
		return nil, errors.New("git provider not started")
	}
	key := repositoryKey(ref.cloneUrl, ref.branch)
	r, ok := p.repos[key]
	if !ok {
		var err error
		if r, err = p.cloneOnDemand(ref); err != nil {
			p.m.Unlock()
			return nil, err
		}
	}
	ctx := p.ctx
	p.m.Unlock()

	select {
	case <-r.ready:
		return r, r.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// cloneOnDemand registers the repository and clones it in the background. The caller holds p.m.
// A repository that is configured (with any branch) gives the clone URL, so its auth applies, and
// the pull interval.
func (p *Provider) cloneOnDemand(ref *reference) (*repository, error) {
	if err := p.ctx.Err(); err != nil {
		return nil, err
	}

	config := static.GitRepo{PullInterval: p.pullInterval}
	u := *ref.cloneUrl
	if configured := p.configuredRepository(ref.cloneUrl); configured != nil {
		if parsed, err := url.Parse(configured.repoUrl); err == nil {
			u = *parsed
		}
		config.PullInterval = configured.config.PullInterval
	}
	if ref.branch != "" {
		u.RawQuery = url.Values{"ref": {ref.branch}}.Encode()
	}
	config.Url = u.String()

	r, err := newRepository(config)
	if err != nil {
		return nil, err
	}
	r.onDemand = true
	p.repos[r.key] = r

	p.pool.Go(func(ctx context.Context) {
		if err := p.initRepository(ctx, r, p.ch, p.pool); err != nil {
			log.Errorf("clone git repository '%v' failed: %v", r.repoUrl, err)
		}
	})
	return r, nil
}

// configuredRepository returns a configured repository with the same repository id, whatever
// its branch, or nil. The caller holds p.m.
func (p *Provider) configuredRepository(cloneUrl *url.URL) *repository {
	id := repositoryId(cloneUrl)
	var found *repository
	for _, r := range p.repos {
		// the smallest key makes the choice stable
		if !r.onDemand && r.id == id && (found == nil || r.key < found.key) {
			found = r
		}
	}
	return found
}

// forget removes a repository that failed to clone, so that the next Read tries again.
func (p *Provider) forget(r *repository) {
	p.m.Lock()
	defer p.m.Unlock()

	if p.repos[r.key] == r && r.onDemand {
		delete(p.repos, r.key)
	}
}

// localFile returns the path of the file in the clone. The file must not leave the clone by a symlink.
func (r *repository) localFile(file string) (string, error) {
	local := filepath.Join(r.localPath, filepath.FromSlash(file))
	resolved, err := filepath.EvalSymlinks(local)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("file '%v' not found in git repository %v", file, r.repoUrl)
		}
		return "", err
	}
	root, err := filepath.EvalSymlinks(r.localPath)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(root, resolved)
	if err != nil || isOutside(rel) || rel == ".git" || strings.HasPrefix(rel, ".git"+string(filepath.Separator)) {
		return "", fmt.Errorf("git file '%v' resolves to a path outside of the repository", file)
	}
	return local, nil
}

func isOutside(rel string) bool {
	return rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
