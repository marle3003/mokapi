package file

import (
	"context"
	"fmt"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"mokapi/config/dynamic/common"
	"mokapi/config/static"
	"mokapi/safe"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Provider struct {
	cfg        static.FileProvider
	SkipPrefix []string

	watcher *fsnotify.Watcher
}

func New(cfg static.FileProvider) *Provider {
	p := &Provider{
		cfg:        cfg,
		SkipPrefix: []string{"_"},
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Errorf("unable to add file watcher: %v", err)
	} else {
		p.watcher = watcher
	}

	return p
}

func (p *Provider) Read(u *url.URL) (*common.Config, error) {
	err := p.watcher.Add(u.Path)
	if err != nil {
		return nil, err
	}
	return p.readFile(u.Path)
}

func (p *Provider) Start(ch chan *common.Config, pool *safe.Pool) error {
	var path string
	if len(p.cfg.Directory) > 0 {
		path = p.cfg.Directory

	} else if len(p.cfg.Filename) > 0 {
		path = p.cfg.Filename
	}
	if len(path) > 0 {
		pool.Go(func(ctx context.Context) {
			if err := p.walk(path, ch); err != nil {
				log.Errorf("file provider: %v", err)
			}
		})

		err := p.watcher.Add(path)
		if err != nil {
			return fmt.Errorf("error adding path to watcher: %v", err)
		}
	}
	return p.watch(ch, pool)
}

func (p *Provider) watch(ch chan<- *common.Config, pool *safe.Pool) error {
	ticker := time.NewTicker(time.Second)
	var events []fsnotify.Event

	pool.Go(func(ctx context.Context) {
		defer func() {
			p.watcher.Close()
			ticker.Stop()
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case evt := <-p.watcher.Events:
				// temporary files ends with '~' in name
				if len(evt.Name) > 0 && !strings.HasSuffix(evt.Name, "~") {
					events = append(events, evt)
				}
			case <-ticker.C:
				m := make(map[string]struct{})
				for _, evt := range events {
					if _, ok := m[evt.Name]; ok {
						continue
					}
					m[evt.Name] = struct{}{}

					dir, file := filepath.Split(evt.Name)
					if dir == evt.Name && !p.skip(dir) {
						if err := p.watcher.Add(dir); err != nil {
							log.Errorf("unable to watch directory %v", dir)
						}
					} else if p.cfg.Filename != "" {
						if _, configFile := filepath.Split(p.cfg.Filename); file == configFile {
							c, err := p.readFile(evt.Name)
							if err != nil {
								log.Errorf("unable to read file %v", evt.Name)
							}
							ch <- c
						}
					} else if !p.skip(evt.Name) {
						c, err := p.readFile(evt.Name)
						if err != nil {
							log.Errorf("unable to read file %v", evt.Name)
						}
						ch <- c
					}
				}
			}
		}
	})
	return nil
}

func (p *Provider) skip(path string) bool {
	name := filepath.Base(path)
	for _, s := range p.SkipPrefix {
		if strings.HasPrefix(name, s) {
			return true
		}
	}
	return false
}

func (p *Provider) readFile(path string) (*common.Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read file %v: %v", path, err)
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	u, _ := url.Parse(fmt.Sprintf("file:%v", abs))

	return &common.Config{
		Url:          u,
		Data:         data,
		ProviderName: "file",
	}, nil
}

func (p *Provider) walk(path string, ch chan<- *common.Config) error {
	walkDir := func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.Mode().IsDir() {
			if p.skip(path) {
				return filepath.SkipDir
			}
		} else if !p.skip(path) {
			if c, err := p.readFile(path); err != nil {
				log.Error(err)
			} else if len(c.Data) > 0 {
				ch <- c
			}
		}

		return nil
	}

	return filepath.Walk(path, walkDir)
}
