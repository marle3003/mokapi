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
	"time"
)

type ConfigWatcher struct {
	providers map[string]dynamic.Provider
	listener  []dynamic.ConfigListener
	configs   map[string]*entry
	cfg       *static.Config
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
		cfg:       cfg,
	}

	w.providers["file"] = file.New(cfg.Providers.File)
	h := http.New(cfg.Providers.Http)
	w.providers["http"] = h
	w.providers["https"] = h
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
	e, exists := w.getConfig(u)

	if !exists {
		c, err = p.Read(u)
		if err != nil {
			w.m.Unlock()
			return nil, err
		}
		e = &entry{config: c}
		w.configs[u.String()] = e
		c.Listeners.Add("ConfigWatcher", w.configChanged)
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
	ch := make(chan dynamic.ConfigEvent)
	for _, p := range w.providers {
		err := p.Start(ch, pool)
		if err != nil {
			return err
		}
	}

	pool.Go(func(ctx context.Context) {
		for i, cfg := range w.cfg.Configs {
			u, _ := url.Parse(fmt.Sprintf("cli://configs/%v.json", i))
			c := &dynamic.Config{
				Info: dynamic.ConfigInfo{
					Provider: "cli",
					Url:      u,
					Checksum: nil,
					Time:     time.Now(),
				},
				Raw: []byte(cfg),
			}
			if err := w.addOrUpdate(dynamic.ConfigEvent{Name: u.String(), Config: c, Event: dynamic.Create}); err != nil {
				log.Error(err)
			}
		}

		for {
			select {
			case <-ctx.Done():
				close(ch)
				return
			case evt := <-ch:
				if evt.Event == dynamic.Delete {
					w.m.Lock()
					e, ok := w.configs[evt.Name]
					w.m.Unlock()
					if !ok {
						continue
					}
					evt.Config = e.config
					go e.config.Listeners.Invoke(evt)
				} else {
					if err := w.addOrUpdate(evt); err != nil {
						log.Error(err)
					}
				}
			}
		}
	})
	return nil
}

func (w *ConfigWatcher) AddListener(f func(e dynamic.ConfigEvent)) {
	w.listener = append(w.listener, f)
}

func (w *ConfigWatcher) addOrUpdate(evt dynamic.ConfigEvent) error {
	w.m.Lock()

	c := evt.Config
	e, ok := w.getConfig(c.Info.Url)
	if !ok && c.Info.Inner() != nil {
		current := c.Info.Inner()
		for !ok {
			if current == nil {
				break
			}
			e, ok = w.getConfig(current.Url)
			current = current.Inner()
		}
		if ok {
			key := e.config.Info.Url.String()
			delete(w.configs, key)
			dynamic.Wrap(c.Info, e.config)
			w.configs[key] = e
		}
	}

	if !ok {
		e = &entry{config: c}
		w.configs[c.Info.Url.String()] = e
		c.Listeners.Add("ConfigWatcher", w.configChanged)
	} else if bytes.Equal(e.config.Info.Checksum, c.Info.Checksum) {
		log.Debugf("Checksum not changed. Skip reloading %v (%v)", e.config.Info.Url.String(), evt.Event)
		w.m.Unlock()
		return nil
	} else {
		e.config.Raw = c.Raw
		e.config.Info.Update(c.Info.Checksum)
		log.Infof("reloading %v", e.config.Info.Url.String())
	}

	w.m.Unlock()
	go e.config.Listeners.Invoke(evt)

	return nil
}

func (w *ConfigWatcher) configChanged(evt dynamic.ConfigEvent) {
	if evt.Event == dynamic.Delete {
		w.remove(evt)
		return
	}

	w.m.Lock()
	e := w.configs[evt.Config.Info.Url.String()]
	e.m.Lock()
	w.m.Unlock()

	c := e.config
	err := dynamic.Parse(c, w)
	if err != nil {
		log.Errorf("parse error %v: %v", c.Info.Path(), err)
		e.m.Unlock()
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
		go l(evt)
	}
}

func (w *ConfigWatcher) remove(evt dynamic.ConfigEvent) {
	w.m.Lock()
	e := w.configs[evt.Config.Info.Url.String()]
	e.m.Lock()
	delete(w.configs, evt.Config.Info.Url.String())
	w.m.Unlock()

	log.Debugf("removing %v", evt.Config.Info.Url.String())
	for _, l := range w.listener {
		go l(evt)
	}
	return
}

func (w *ConfigWatcher) getConfig(u *url.URL) (*entry, bool) {
	if e, ok := w.configs[u.String()]; ok {
		return e, true
	}

	for _, cfg := range w.configs {
		if cfg.config.Info.Match(u) {
			return cfg, true
		}
	}
	return nil, false
}
