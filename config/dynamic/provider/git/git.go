package git

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/common"
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

	repo        *git.Repository
	wt          *git.Worktree
	pullOptions *git.PullOptions
	hash        plumbing.Hash
}

type Provider struct {
	repositories []*repository
	pullInterval string
}

func New(config static.GitProvider) *Provider {
	gitUrls := config.Urls
	if len(config.Url) > 0 {
		gitUrls = append(gitUrls, config.Url)
	}

	var repos []*repository
	for _, gitUrl := range gitUrls {
		path, err := os.MkdirTemp("", "mokapi_git_*")
		if err != nil {
			log.Errorf("unable to create temp dir for git provider: %v", err)
		}

		u, err := url.Parse(gitUrl)
		if err != nil {
			log.Errorf("unable to parse git url %v: %v", config.Url, err)
		}

		var ref string
		q := u.Query()
		if len(q) > 0 {
			ref = q.Get("ref")
			q.Del("ref")
			u.RawQuery = q.Encode()
		}

		repos = append(repos, &repository{
			url:       gitUrl,
			repoUrl:   u.String(),
			localPath: path,
			ref:       ref,
			hash:      plumbing.Hash{},
		})
	}

	return &Provider{
		repositories: repos,
		pullInterval: config.PullInterval,
	}
}

func (p *Provider) Read(_ *url.URL) (*common.Config, error) {
	return nil, fmt.Errorf("not supported")
}

func (p *Provider) Start(ch chan *common.Config, pool *safe.Pool) error {
	if len(p.repositories) == 0 {
		return nil
	}

	var err error
	interval := time.Second * 5
	if len(p.pullInterval) > 0 {
		interval, err = time.ParseDuration(p.pullInterval)
		if err != nil {
			return fmt.Errorf("unable to parse interval %q: %v", p.pullInterval, err)
		}
	}

	ticker := time.NewTicker(interval)

	for _, r := range p.repositories {
		r := r
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

		var ref plumbing.ReferenceName
		r.pullOptions = &git.PullOptions{SingleBranch: true}
		if len(r.ref) > 0 {
			ref = plumbing.NewBranchReferenceName(r.ref)

			if h.Name() != ref {
				r.pullOptions.ReferenceName = ref
				err = r.repo.Fetch(&git.FetchOptions{RefSpecs: []config.RefSpec{"refs/*:refs/*", "HEAD:refs/heads/HEAD"}})
				if err != nil {
					return fmt.Errorf("git fetch error %v: %v", r.url, err.Error())
				}
				err = r.wt.Checkout(&git.CheckoutOptions{Branch: ref})
				if err != nil && err != git.NoErrAlreadyUpToDate {
					return fmt.Errorf("git checkout error %v: %v", r.url, err.Error())
				}
			}
		}

		chFile := make(chan *common.Config)
		p.startFileProvider(r.localPath, chFile, pool)

		pool.Go(func(ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
					return
				case c := <-chFile:

					info := common.ConfigInfo{
						Provider: "git",
						Url:      getUrl(r, c.Info.Url),
					}
					common.Wrap(info, c)
					ch <- c
				}
			}
		})
	}

	pool.Go(func(ctx context.Context) {
		defer func() {
			ticker.Stop()
			p.cleanup()
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				for _, r := range p.repositories {
					err = r.repo.Fetch(&git.FetchOptions{RefSpecs: []config.RefSpec{"refs/*:refs/*", "HEAD:refs/heads/HEAD"}})
					if err != nil {
						if err != git.NoErrAlreadyUpToDate {
							log.Errorf("git fetch error %v: %v", r.url, err.Error())
						}
						continue
					}

					err = r.wt.Pull(r.pullOptions)
					if err != nil {
						if err != git.NoErrAlreadyUpToDate {
							log.Errorf("unable to pull: %v", err.Error())
						}
						continue
					}

					ref, err := r.repo.Head()
					if err != nil {
						log.Errorf("unable to get head: %v", err.Error())
						continue
					}

					hash := ref.Hash()
					if hash != r.hash {
						log.Infof("updated git repository from remote")
						r.hash = hash
					}
				}
			}
		}
	})

	return nil
}

func (p *Provider) startFileProvider(dir string, ch chan *common.Config, pool *safe.Pool) {
	f := file.New(static.FileProvider{Directory: dir})
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
