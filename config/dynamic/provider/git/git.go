package git

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/client"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/provider/file"
	"mokapi/config/static"
	"mokapi/safe"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type repository struct {
	url       string
	repoUrl   string
	localPath string
	ref       string
	auth      *static.GitAuth

	repo        *git.Repository
	wt          *git.Worktree
	pullOptions *git.PullOptions
	hash        plumbing.Hash
	config      static.GitRepo
	init        bool
}

type Provider struct {
	repositories []*repository
	pullInterval string
	transport    *transport
}

func New(config static.GitProvider) *Provider {
	repoConfigs := config.Repositories
	for _, url := range config.Urls {
		if len(url) > 0 {
			repoConfigs = append(repoConfigs, static.GitRepo{Url: url})
		}
	}

	var repos []*repository
	for _, repoConfig := range repoConfigs {
		path, err := os.MkdirTemp(config.TempDir, "mokapi_git_*")
		if err != nil {
			log.Errorf("unable to create temp dir for git provider: %v", err)
		}

		u, err := url.Parse(repoConfig.Url)
		if err != nil {
			log.Errorf("unable to parse git url %v: %v", repoConfig.Url, err)
		}

		var ref string
		q := u.Query()
		if len(q) > 0 {
			ref = q.Get("ref")
			q.Del("ref")
			u.RawQuery = q.Encode()
		}

		repos = append(repos, &repository{
			url:       repoConfig.Url,
			repoUrl:   u.String(),
			localPath: path,
			ref:       ref,
			hash:      plumbing.Hash{},
			auth:      repoConfig.Auth,
			config:    repoConfig,
		})
	}

	t := newTransport()
	client.InstallProtocol("http", http.NewClient(newClient(t)))
	client.InstallProtocol("https", http.NewClient(newClient(t)))

	return &Provider{
		repositories: repos,
		pullInterval: config.PullInterval,
		transport:    t,
	}
}

func (p *Provider) Read(_ *url.URL) (*dynamic.Config, error) {
	return nil, fmt.Errorf("not supported")
}

func (p *Provider) Start(ch chan *dynamic.Config, pool *safe.Pool) error {
	if len(p.repositories) == 0 {
		return nil
	}

	for _, r := range p.repositories {
		r := r
		pool.Go(func(ctx context.Context) {
			err := p.initRepository(r, ch, pool)
			if err != nil {
				log.Errorf("init git repository failed: %v", err)
			}
		})
	}

	return nil
}

func (p *Provider) initRepository(r *repository, ch chan *dynamic.Config, pool *safe.Pool) error {
	if len(r.config.PullInterval) == 0 {
		r.config.PullInterval = p.pullInterval
	}

	err := p.transport.addAuth(r)
	if err != nil {
		return err
	}

	r.repo, err = git.PlainClone(r.localPath, false, &git.CloneOptions{
		URL: r.repoUrl,
	})
	if err != nil {
		return fmt.Errorf("unable to clone git %q: %v", r.repoUrl, err)
	}

	r.wt, err = r.repo.Worktree()
	if err != nil {
		return fmt.Errorf("unable to get git worktree: %v", err.Error())
	}

	h, err := r.repo.Head()
	if err != nil {
		return fmt.Errorf("unable to get git head: %v", err.Error())
	}

	r.pullOptions = &git.PullOptions{SingleBranch: true}
	if len(r.ref) > 0 {
		ref := plumbing.NewBranchReferenceName(r.ref)

		if h.Name() != ref {
			r.pullOptions.ReferenceName = ref
			err = r.repo.Fetch(&git.FetchOptions{RefSpecs: []config.RefSpec{"refs/*:refs/*", "HEAD:refs/heads/HEAD"}})
			if errors.Is(err, git.ErrForceNeeded) {
				err = r.repo.Fetch(&git.FetchOptions{RefSpecs: []config.RefSpec{"+refs/*:refs/*", "HEAD:refs/heads/HEAD"}})
			}
			if err != nil {
				return fmt.Errorf("git fetch error %v: %v", r.url, err.Error())
			}
			err = r.wt.Checkout(&git.CheckoutOptions{Branch: ref})
			if err != nil && err != git.NoErrAlreadyUpToDate {
				return fmt.Errorf("git checkout error %v: %v", r.url, err.Error())
			}
		}
	}

	ref, err := r.repo.Head()
	r.hash = ref.Hash()

	chFile := make(chan *dynamic.Config)
	p.startFileProvider(r.localPath, chFile, pool)

	err = startPullInterval(r, pool)
	if err != nil {
		return err
	}
	pool.Go(func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case c := <-chFile:
				path := c.Info.Url.Path
				if len(c.Info.Url.Opaque) > 0 {
					path = c.Info.Url.Opaque
				}
				relative := path[len(r.localPath)+1:]
				if skip(relative, r) {
					log.Debugf("skip file: %v", getUrl(r, c.Info.Url))
					continue
				}

				wrapConfig(c, r)
				ch <- c
			}
		}
	})
	return nil
}

func (p *Provider) startFileProvider(dir string, ch chan *dynamic.Config, pool *safe.Pool) {
	f := file.New(static.FileProvider{Directories: []string{dir}})
	f.SkipPrefix = append(f.SkipPrefix, ".git")
	err := f.Start(ch, pool)
	if err != nil {
		log.Errorf("unable to start file provider for git: %v", err)
	}
}

func (p *Provider) cleanup() {
	for _, repo := range p.repositories {
		err := os.RemoveAll(repo.localPath)
		if err != nil {
			log.Debugf("unable to remove temp dir %q: %v", repo.localPath, err.Error())
		}
	}
}

func getUrl(r *repository, file *url.URL) *url.URL {
	path := file.Path
	if len(file.Opaque) > 0 {
		path = file.Opaque
	}
	path = strings.TrimPrefix(path, r.localPath)

	u, _ := url.Parse(r.url)
	path = filepath.ToSlash(path)
	q := u.Query()
	q.Add("file", path)
	u.RawQuery = q.Encode()
	return u
}

func skip(path string, repo *repository) bool {
	if len(repo.config.Files) == 0 && len(repo.config.Include) == 0 {
		return false
	}

	path = filepath.ToSlash(path)

	if contains(repo.config.Files, path) {
		return false
	}
	if match(repo.config.Include, path) {
		return false
	}

	return true
}

func startPullInterval(r *repository, pool *safe.Pool) error {
	if len(r.config.PullInterval) == 0 {
		return nil
	}
	interval, err := time.ParseDuration(r.config.PullInterval)
	if err != nil {
		return fmt.Errorf("unable to parse interval %q: %v", r.config.PullInterval, err)
	}

	ticker := time.NewTicker(interval)

	pool.Go(func(ctx context.Context) {
		defer func() {
			ticker.Stop()
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				pull(r)
			}
		}
	})
	return nil
}

func pull(r *repository) {
	if r.repo == nil || r.config.PullInterval != "" {
		return
	}
	err := r.repo.Fetch(&git.FetchOptions{RefSpecs: []config.RefSpec{"refs/*:refs/*", "HEAD:refs/heads/HEAD"}})
	if errors.Is(err, git.ErrForceNeeded) {
		err = r.repo.Fetch(&git.FetchOptions{RefSpecs: []config.RefSpec{"+refs/*:refs/*", "HEAD:refs/heads/HEAD"}})
	}
	if err != nil {
		if !errors.Is(err, git.NoErrAlreadyUpToDate) {
			log.Errorf("git fetch error %v: %v", r.url, err.Error())
		}
		return
	}

	err = r.wt.Pull(r.pullOptions)
	if err != nil {
		if !errors.Is(err, git.NoErrAlreadyUpToDate) {
			log.Errorf("unable to pull: %v", err.Error())
		}
		return
	}

	ref, err := r.repo.Head()
	if err != nil {
		log.Errorf("unable to get head: %v", err.Error())
		return
	}

	hash := ref.Hash()
	if hash != r.hash {
		log.Infof("updated git repository from remote")
		r.hash = hash
	}
}

func contains(s []string, v string) bool {
	for _, i := range s {
		if i == v {
			return true
		}
	}
	return false
}

func match(s []string, v string) bool {
	for _, i := range s {
		if file.Match(i, v) {
			return true
		}
	}
	return false
}

func wrapConfig(c *dynamic.Config, r *repository) {
	u := getUrl(r, c.Info.Url)
	info := dynamic.ConfigInfo{
		Provider: "git",
		Url:      u,
		// to query git log takes too long
		Time: time.Now(),
	}

	dynamic.Wrap(info, c)
}
