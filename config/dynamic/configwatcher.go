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
	listener  []func(c *common.Config)
	configs   map[string]*common.Config
	m         sync.Mutex
}

func NewConfigWatcher(cfg *static.Config) *ConfigWatcher {
	w := &ConfigWatcher{
		providers: make(map[string]common.Provider),
		configs:   make(map[string]*common.Config),
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

func (w *ConfigWatcher) Read(u *url.URL, opts ...common.ConfigOptions) (*common.Config, error) {
	p, ok := w.providers[u.Scheme]
	if !ok {
		return nil, fmt.Errorf("unsupported scheme: %v", u.String())
	}
	w.m.Lock()

	var err error
	c, exists := w.configs[u.String()]
	if !exists {
		c, err = p.Read(u)
		if err != nil {
			w.m.Unlock()
			return nil, err
		}
		w.configs[u.String()] = c
		w.m.Unlock()
		c.Options(opts...)
		err = c.Parse(w)
		if err != nil {
			return nil, err
		}
	} else {
		w.m.Unlock()
		c.Options(opts...)
	}

	return c, nil
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
				close(ch)
				return
			case c := <-ch:
				if err := w.addOrUpdate(c); err != nil {
					log.Error(err)
				}
			}
		}
	})
	return nil
}

func (w *ConfigWatcher) AddListener(f func(config *common.Config)) {
	w.listener = append(w.listener, f)
}

func (w *ConfigWatcher) addOrUpdate(c *common.Config) error {
	w.m.Lock()

	cfg, ok := w.configs[c.Url.String()]
	if !ok {
		w.configs[c.Url.String()] = c
		cfg = c
		cfg.Listeners = append(cfg.Listeners, w.listener...)
	} else {
		cfg.Raw = c.Raw
	}
	w.m.Unlock()

	err := cfg.Parse(w)
	if err != nil {
		return err
	}

	cfg.Changed()
	return nil
}
