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

type Provider struct {
	url          string
	repoUrl      string
	localPath    string
	pullInterval string
	ref          string
	directories  []string

	hash plumbing.Hash
}

func New(config static.GitProvider) *Provider {
	path, err := os.MkdirTemp("", "mokapi_git_*")
	if err != nil {
		log.Errorf("unable to create temp dir for git provider: %v", err)
	}

	u, err := url.Parse(config.Url)
	if err != nil {
		log.Errorf("unable to parse git url %v: %v", config.Url, err)
	}

	split := strings.Split(u.Path, "//")
	var directories []string
	if len(split) > 1 {
		directories = append(directories, split[1])
		u.Path = split[0]
	}

	var ref string
	q := u.Query()
	if len(q) > 0 {
		ref = q.Get("ref")
		q.Del("ref")
		u.RawQuery = q.Encode()
	}

	return &Provider{
		url:          config.Url,
		repoUrl:      u.String(),
		localPath:    path,
		pullInterval: config.PullInterval,
		ref:          ref,
		hash:         plumbing.Hash{},
		directories:  directories,
	}
}

func (p *Provider) Read(u *url.URL) (*common.Config, error) {
	return nil, fmt.Errorf("not supported")
}

func (p *Provider) Start(ch chan *common.Config, pool *safe.Pool) error {
	if len(p.url) == 0 {
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
	repo, err := git.PlainClone(p.localPath, false, &git.CloneOptions{
		URL: p.repoUrl,
	})
	if err != nil {
		return fmt.Errorf("unable to clone git %q: %v", p.url, err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("unable to get git worktree: %v", err.Error())
	}

	h, err := repo.Head()
	if err != nil {
		return fmt.Errorf("unable to get git head: %v", err.Error())
	}

	var ref plumbing.ReferenceName
	pullOptions := &git.PullOptions{SingleBranch: true}
	if len(p.ref) > 0 {
		ref = plumbing.NewBranchReferenceName(p.ref)

		if h.Name() != ref {
			pullOptions.ReferenceName = ref
			err = repo.Fetch(&git.FetchOptions{RefSpecs: []config.RefSpec{"refs/*:refs/*", "HEAD:refs/heads/HEAD"}})
			if err != nil {
				return fmt.Errorf("git fetch error %v: %v", p.url, err.Error())
			}
			err = wt.Checkout(&git.CheckoutOptions{Branch: ref})
			if err != nil && err != git.NoErrAlreadyUpToDate {
				return fmt.Errorf("git checkout error %v: %v", p.url, err.Error())
			}
		}
	}

	chFile := make(chan *common.Config)
	p.startFileProvider(chFile, pool)
	repoUrl, _ := url.Parse(p.url)

	pool.Go(func(ctx context.Context) {
		defer func() {
			ticker.Stop()
			err := os.RemoveAll(p.localPath)
			if err != nil {
				log.Debugf("unable to remove temp dir %q: %v", p.localPath, err.Error())
			}
		}()

		for {
			select {
			case c := <-chFile:
				path := c.Info.Url.Path
				if len(c.Info.Url.Opaque) > 0 {
					path = c.Info.Url.Opaque
				}
				path = strings.TrimPrefix(path, p.localPath)
				c.Info.Parent = &common.ConfigInfo{
					Provider: "git",
					Url:      addFilePath(repoUrl, path),
					Parent:   nil,
				}
				ch <- c
			case <-ctx.Done():
				return
			case <-ticker.C:
				err = repo.Fetch(&git.FetchOptions{RefSpecs: []config.RefSpec{"refs/*:refs/*", "HEAD:refs/heads/HEAD"}})
				if err != nil {
					if err != git.NoErrAlreadyUpToDate {
						log.Errorf("git fetch error %v: %v", p.url, err.Error())
					}
					continue
				}

				err = wt.Pull(pullOptions)
				if err != nil {
					if err != git.NoErrAlreadyUpToDate {
						log.Errorf("unable to pull: %v", err.Error())
					}
					continue
				}

				ref, err := repo.Head()
				if err != nil {
					log.Errorf("unable to get head: %v", err.Error())
					continue
				}

				hash := ref.Hash()
				if hash != p.hash {
					log.Infof("updated git repository from remote")
					p.hash = hash
				}

			}
		}
	})

	return nil
}

func (p *Provider) startFileProvider(ch chan *common.Config, pool *safe.Pool) {
	f := file.New(static.FileProvider{Directory: p.localPath})
	f.SkipPrefix = append(f.SkipPrefix, ".git")
	err := f.Start(ch, pool)
	if err != nil {
		log.Errorf("unable to start file provider for git: %v", err)
	}
}

func addFilePath(u *url.URL, path string) *url.URL {
	newUrl := *u
	path = filepath.ToSlash(path)
	q := newUrl.Query()
	q.Add("file", path)
	newUrl.RawQuery = q.Encode()
	return &newUrl
}
