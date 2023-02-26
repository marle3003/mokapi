package dynamic

import (
	"bytes"
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
	listener  []common.ConfigListener
	configs   map[string]*common.Config
	m         sync.Mutex
}

func NewConfigWatcher(cfg *static.Config) *ConfigWatcher {
	w := &ConfigWatcher{
		providers: make(map[string]common.Provider),
		configs:   make(map[string]*common.Config),
	}

	w.providers["file"] = file.New(cfg.Providers.File)
	http := http.New(cfg.Providers.Http)
	w.providers["http"] = http
	w.providers["https"] = http
	w.providers["git"] = git.New(cfg.Providers.Git)

	return w
}

func (w *ConfigWatcher) Read(u *url.URL, opts ...common.ConfigOptions) (*common.Config, error) {
	p, ok := w.providers[u.Scheme]
	if !ok {
		return nil, fmt.Errorf("unsupported scheme: %v", u.String())
	}
	w.m.Lock()

	var err error
	var parse bool
	c, exists := w.configs[u.String()]
	if !exists {
		c, err = p.Read(u)
		if err != nil {
			w.m.Unlock()
			return nil, err
		}
		w.configs[u.String()] = c
		w.m.Unlock()
		c.AddListener("ConfigWatcher", func(cfg *common.Config) {
			w.configChanged(cfg)
		})
		parse = true
	} else {
		w.m.Unlock()
		parse = c.Data == nil
	}

	c.Options(opts...)
	if parse {
		err = c.Parse(w)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (w *ConfigWatcher) Start(pool *safe.Pool) error {
	ch := make(chan *common.Config)
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

func (w *ConfigWatcher) ReadServices(services static.Services) {
	for name, p := range services {
		var err error
		var cfg *common.Config
		if len(p.Config.File) > 0 {
			u, _ := url.Parse(fmt.Sprintf("file:%v", p.Config.File))
			cfg, err = w.Read(u)

		}
		if len(p.Config.Url) > 0 {
			u, _ := url.Parse(fmt.Sprintf(p.Config.Url))
			cfg, err = w.Read(u)
		}
		if err != nil {
			log.Errorf("unable to read config for %v: %v", name, err)
			continue
		}
		if cfg != nil {
			cfg.Key = name
			w.configChanged(cfg)
		}
	}
}

func (w *ConfigWatcher) addOrUpdate(c *common.Config) error {
	w.m.Lock()

	cfg, ok := w.configs[c.Url.String()]
	if !ok {
		w.configs[c.Url.String()] = c
		cfg = c
		cfg.AddListener("ConfigWatcher", func(cfg *common.Config) {
			w.configChanged(cfg)
		})
	} else if bytes.Equal(cfg.Checksum, c.Checksum) {
		w.m.Unlock()
		return nil
	} else {
		cfg.Raw = c.Raw
		cfg.Checksum = c.Checksum
	}

	w.m.Unlock()
	go cfg.Changed()

	return nil
}

func (w *ConfigWatcher) configChanged(cfg *common.Config) {
	err := cfg.Parse(w)
	if err != nil {
		log.Errorf("parse error %v: %v", cfg.Url, err)
		return
	}

	if err = cfg.Validate(); err != nil {
		log.Infof("skipping file %v: %v", cfg.Url, err)
		return
	}

	log.Debugf("processing %v", cfg.Url)

	for _, l := range w.listener {
		l(cfg)
	}
}
