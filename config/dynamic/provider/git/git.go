package git

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/provider/file"
	"mokapi/config/static"
	"mokapi/safe"
	"net/url"
	"os"
	"time"
)

type Provider struct {
	url          string
	path         string
	pullInterval string

	hash plumbing.Hash
}

func New(config static.GitProvider) *Provider {
	path, err := ioutil.TempDir("", "mokapi_git_*")
	if err != nil {
		log.Errorf("unable to create temp dir for git provider: %v", err)
	}
	return &Provider{
		url:          config.Url,
		path:         path,
		pullInterval: config.PullInterval,
		hash:         plumbing.Hash{},
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
	repo, err := git.PlainClone(p.path, false, &git.CloneOptions{
		URL: p.url,
	})
	if err != nil {
		return fmt.Errorf("unable to clone git %q: %v", p.url, err)
	}

	f := file.New(static.FileProvider{Directory: p.path})
	f.SkipPrefix = append(f.SkipPrefix, ".git")
	err = f.Start(ch, pool)
	if err != nil {
		log.Errorf("unable to start file provider for git: %v", err)
	}

	pool.Go(func(ctx context.Context) {
		defer func() {
			ticker.Stop()
			err := os.RemoveAll(p.path)
			if err != nil {
				log.Debugf("unable to remove temp dir %q: %v", p.path, err.Error())
			}
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				wt, err := repo.Worktree()
				if err != nil {
					log.Errorf("unable to get git worktree: %v", err.Error())
				}
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
