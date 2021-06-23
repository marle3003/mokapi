package dynamic

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	log "github.com/sirupsen/logrus"
	"mokapi/config/static"
	"os"
	"time"
)

type gitWatcher struct {
	url          string
	path         string
	pullInterval string

	close chan bool

	hash plumbing.Hash
}

func newGitWatcher(path string, config static.GitProvider) *gitWatcher {
	return &gitWatcher{
		url:          config.Url,
		path:         path,
		pullInterval: config.PullInterval,
		close:        make(chan bool),
		hash:         plumbing.Hash{},
	}
}

func (g *gitWatcher) Start() {
	go func() {
		repo, err := git.PlainClone(g.path, false, &git.CloneOptions{
			URL: g.url,
		})
		if err != nil {
			log.Errorf("unable to clone git %q: %v", g.url, err.Error())
			return
		}

		interval := time.Second * 5
		if len(g.pullInterval) > 0 {
			interval, err = time.ParseDuration(g.pullInterval)
			if err != nil {
				log.Errorf("unable to parse interval %q: %v", g.pullInterval, err.Error())
			}
		}

		ticker := time.NewTicker(interval)

		defer func() {
			ticker.Stop()
			err := os.RemoveAll(g.path)
			if err != nil {
				log.Debugf("unable to remove temp dir %q: %v", g.path, err.Error())
			}
		}()

		for {
			select {
			case <-g.close:
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
				if hash != g.hash {
					log.Infof("updated git repository from remote")
					g.hash = hash
				}

			}
		}
	}()
}
