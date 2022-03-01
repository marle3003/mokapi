package dynamic

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/provider/file"
	"mokapi/config/dynamic/provider/git"
	"mokapi/config/dynamic/provider/http"
	"mokapi/config/static"
	"mokapi/safe"
	"net/url"
	"sync"
)

type ConfigWatcher struct {
	providers map[string]common.Provider
	listener  []func(c *common.File)
	files     map[string]*common.File
	m         sync.Mutex
}

func NewConfigWatcher(cfg *static.Config) *ConfigWatcher {
	w := &ConfigWatcher{
		providers: make(map[string]common.Provider),
		files:     make(map[string]*common.File),
	}
	if len(cfg.Providers.File.Filename) > 0 || len(cfg.Providers.File.Directory) > 0 {
		w.providers["file"] = file.New(cfg.Providers.File)
	}

	if len(cfg.Providers.Http.Url) > 0 {
		w.providers["http"] = http.New(cfg.Providers.Http)
	}

	if len(cfg.Providers.Git.Url) > 0 {
		w.providers["git"] = git.New(cfg.Providers.Git)
	}

	if len(w.providers) == 0 {
		log.Infof("no providers configured")
	}

	return w
}

func (w *ConfigWatcher) Read(u *url.URL, opts ...common.FileOptions) (*common.File, error) {
	p, ok := w.providers[u.Scheme]
	if !ok {
		return nil, fmt.Errorf("unsupported scheme: %v", u.String())
	}

	readFile := !w.hasFile(u)
	f := w.getFile(u, opts...)

	if readFile {
		c, err := p.Read(u)
		if err != nil {
			return nil, err
		}
		err = f.Parse(c, w)
		if err != nil {
			return nil, err
		}
	}

	return f, nil
}

func (w *ConfigWatcher) Start(pool *safe.Pool) error {
	ch := make(chan *common.Config, 100)
	for _, p := range w.providers {
		err := p.Start(ch, pool)
		if err != nil {
			return err
		}
	}

	pool.Go(func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case c := <-ch:
				f := w.getFile(c.Url)
				err := f.Parse(c, w)
				if err != nil {
					log.Error(err)
					continue
				}
				if len(f.Listeners) == 0 {
					f.Listeners = append(f.Listeners, w.listener...)
				}

				f.Changed()
			}
		}
	})
	return nil
}

func (w *ConfigWatcher) AddListener(f func(*common.File)) {
	w.listener = append(w.listener, f)
}

func (w *ConfigWatcher) getFile(u *url.URL, opts ...common.FileOptions) *common.File {
	w.m.Lock()
	defer w.m.Unlock()

	f, ok := w.files[u.String()]
	if !ok {
		f = common.NewFile(u, opts...)
		w.files[u.String()] = f
	} else {
		f.Options(opts...)
	}
	return f
}

func (w *ConfigWatcher) hasFile(u *url.URL) bool {
	w.m.Lock()
	defer w.m.Unlock()

	_, ok := w.files[u.String()]
	return ok
}
