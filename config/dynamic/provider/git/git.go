package git

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/provider/file"
	"mokapi/config/static"
	"mokapi/safe"
	"net/url"
	"os"
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
		URL:        p.repoUrl,
		NoCheckout: true,
	})
	if err != nil {
		return fmt.Errorf("unable to clone git %q: %v", p.url, err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		log.Errorf("unable to get git worktree: %v", err.Error())
	}

	var ref plumbing.ReferenceName
	if len(p.ref) > 0 {
		ref = plumbing.NewBranchReferenceName(p.ref)
	}

	err = wt.Checkout(&git.CheckoutOptions{
		SparseCheckoutDirectories: p.directories,
		Branch:                    ref,
	})
	if err != nil {
		log.Errorf("git checkout error %v: %v", p.url, err.Error())
	}

	p.startFileProvider(ch, pool)

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
			case <-ctx.Done():
				return
			case <-ticker.C:
				err = wt.Pull(&git.PullOptions{})
				if err != nil {
					if err != git.NoErrAlreadyUpToDate {
						log.Errorf("unable to pull: %v", err.Error())
						continue
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
