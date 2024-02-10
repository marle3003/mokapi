package server

import (
	"bytes"
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/provider/file"
	"mokapi/config/dynamic/provider/git"
	"mokapi/config/dynamic/provider/http"
	"mokapi/config/dynamic/provider/npm"
	"mokapi/config/static"
	"mokapi/safe"
	"net/url"
	"sync"
)

type ConfigWatcher struct {
	providers map[string]dynamic.Provider
	listener  []dynamic.ConfigListener
	configs   map[string]*entry
	m         sync.Mutex
}

type entry struct {
	config *dynamic.Config
	m      sync.Mutex
}

func NewConfigWatcher(cfg *static.Config) *ConfigWatcher {
	w := &ConfigWatcher{
		providers: make(map[string]dynamic.Provider),
		configs:   make(map[string]*entry),
	}

	w.providers["file"] = file.New(cfg.Providers.File)
	http := http.New(cfg.Providers.Http)
	w.providers["http"] = http
	w.providers["https"] = http
	w.providers["git"] = git.New(cfg.Providers.Git)
	w.providers["npm"] = npm.New(cfg.Providers.Npm)

	return w
}

func (w *ConfigWatcher) Read(u *url.URL, v any) (*dynamic.Config, error) {
	p, ok := w.providers[u.Scheme]
	if !ok {
		return nil, fmt.Errorf("unsupported scheme: %v", u.String())
	}
	w.m.Lock()

	var err error
	var parse bool
	var c *dynamic.Config
	e, exists := w.configs[u.String()]
	if !exists {
		c, err = p.Read(u)
		if err != nil {
			w.m.Unlock()
			return nil, err
		}
		e = &entry{config: c}
		w.configs[u.String()] = e
		c.Listeners.Add("ConfigWatcher", func(cfg *dynamic.Config) {
			w.configChanged(cfg)
		})
		parse = true
	} else {
		c = e.config
		parse = c.Data == nil
	}

	if parse {
		e.m.Lock()
		defer e.m.Unlock()
		w.m.Unlock()

		if c.Data == nil {
			if v != nil {
				c.Data = v
			}
			// Currently, read does not validate config. Add Validate would break compatibility
			err = dynamic.Parse(e.config, w)
			if err != nil {
				return nil, err
			}
		}
	} else {
		w.m.Unlock()
	}

	return c, nil
}

func (w *ConfigWatcher) Start(pool *safe.Pool) error {
	ch := make(chan *dynamic.Config)
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

func (w *ConfigWatcher) AddListener(f func(config *dynamic.Config)) {
	w.listener = append(w.listener, f)
}

func (w *ConfigWatcher) addOrUpdate(c *dynamic.Config) error {
	w.m.Lock()

	e, ok := w.configs[c.Info.Url.String()]
	if !ok {
		e = &entry{config: c}
		w.configs[c.Info.Url.String()] = e
		c.Listeners.Add("ConfigWatcher", func(cfg *dynamic.Config) {
			w.configChanged(cfg)
		})
	} else if bytes.Equal(e.config.Info.Checksum, c.Info.Checksum) {
		w.m.Unlock()
		return nil
	} else {
		e.config.Raw = c.Raw
		e.config.Info.Update(c.Info.Checksum)
	}

	w.m.Unlock()
	go e.config.Listeners.Invoke(e.config)

	return nil
}

func (w *ConfigWatcher) configChanged(c *dynamic.Config) {
	w.m.Lock()
	e := w.configs[c.Info.Url.String()]
	e.m.Lock()
	w.m.Unlock()

	err := dynamic.Parse(c, w)

	if err != nil {
		log.Errorf("parse error %v: %v", c.Info.Path(), err)
		return
	}

	if c.Data == nil {
		e.m.Unlock()
		return
	}

	if err = dynamic.Validate(c); err != nil {
		log.Infof("skipping file %v: %v", c.Info.Path(), err)
		return
	}

	e.m.Unlock()

	log.Debugf("processing %v", c.Info.Path())

	for _, l := range w.listener {
		l(e.config)
	}
}
