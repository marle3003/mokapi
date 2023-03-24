package file

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"io"
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
	watched    map[string]struct{}
	isInit     bool

	watcher *fsnotify.Watcher
}

func New(cfg static.FileProvider) *Provider {
	p := &Provider{
		cfg:        cfg,
		SkipPrefix: []string{"_"},
		watched:    make(map[string]struct{}),
		isInit:     true,
	}

	if cfg.SkipPrefix != nil {
		p.SkipPrefix = cfg.SkipPrefix
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
	file := u.Path
	if len(u.Opaque) > 0 {
		file = u.Opaque
	}

	err := p.watcher.Add(file)
	if err != nil {
		log.Warnf("unable to add watcher on file %v: %v", file, err)
	}
	return p.readFile(file)
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
			for _, i := range strings.Split(path, string(os.PathListSeparator)) {
				if err := p.walk(i, ch); err != nil {
					log.Errorf("file provider: %v", err)
				}
			}
			p.isInit = false
		})
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
					fileInfo, err := os.Stat(evt.Name)
					if err != nil {
						// skip
						continue
					}
					if !fileInfo.IsDir() {
						events = append(events, evt)
					} else if _, ok := p.watched[evt.Name]; !p.isInit && !ok {
						pool.Go(func(ctx context.Context) {
							err := p.walk(evt.Name, ch)
							if err != nil {
								log.Errorf("unable to process dir %v: %v", evt.Name, err)
							}
						})
					}
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
						p.watchPath(dir)
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
				events = make([]fsnotify.Event, 0)
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
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read file %v: %v", path, err)
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	u, _ := url.Parse(fmt.Sprintf("file:%v", abs))

	h := sha256.New()
	if _, err := io.Copy(h, bytes.NewReader(data)); err != nil {
		return nil, err
	}

	return &common.Config{
		Url:          u,
		Raw:          data,
		Checksum:     h.Sum(nil),
		ProviderName: "file",
	}, nil
}

func (p *Provider) walk(root string, ch chan<- *common.Config) error {
	walkDir := func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.Mode().IsDir() {
			if p.skip(path) {
				return filepath.SkipDir
			}
			p.watchPath(path)
		} else if !p.skip(path) {
			if c, err := p.readFile(path); err != nil {
				log.Error(err)
			} else if len(c.Raw) > 0 {
				p.watchPath(path)
				ch <- c
			}
		}

		return nil
	}

	return filepath.Walk(root, walkDir)
}

func (p *Provider) watchPath(path string) {
	if _, ok := p.watched[path]; ok {
		return
	}
	p.watched[path] = struct{}{}
	p.watcher.Add(path)
}
