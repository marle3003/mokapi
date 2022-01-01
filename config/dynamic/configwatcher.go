package dynamic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Masterminds/sprig"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/provider/file"
	"mokapi/config/dynamic/provider/git"
	"mokapi/config/dynamic/provider/http"
	"mokapi/config/dynamic/script"
	"mokapi/config/static"
	"mokapi/safe"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
)

type ConfigWatcher struct {
	providers map[string]common.Provider
	listener  []func(c *common.File)
	files     map[string]*common.File
	fileMutex map[string]*sync.Mutex
	m         sync.RWMutex
}

func NewConfigWatcher(cfg *static.Config) *ConfigWatcher {
	w := &ConfigWatcher{
		providers: make(map[string]common.Provider),
		files:     make(map[string]*common.File),
		fileMutex: make(map[string]*sync.Mutex),
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

	return w
}

func (w *ConfigWatcher) Read(u *url.URL, opts ...common.FileOptions) (*common.File, error) {
	r, ok := w.providers[u.Scheme]
	if !ok {
		return nil, fmt.Errorf("unsupported scheme: %v", u.String())
	}
	f := w.getFile(u)
	if f.Data == nil {
		c, err := r.Read(u)
		if err != nil {
			return nil, err
		}
		f, err = w.loadConfig(c)
		if err != nil {
			return nil, err
		}
	}

	for _, opt := range opts {
		opt(f)
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
				f, err := w.loadConfig(c)
				if err != nil {
					log.Error(err)
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

func (w *ConfigWatcher) loadConfig(c *common.Config) (*common.File, error) {
	_, name := filepath.Split(c.Url.String())

	if filepath.Ext(name) == ".tmpl" {
		var err error
		c.Data, err = renderTemplate(c.Data)
		if err != nil {
			return nil, fmt.Errorf("unable to render template %v: %v", c.Url, err)
		}
		name = name[0 : len(name)-len(filepath.Ext(name))]
	}

	f := w.getFile(c.Url)
	m := w.getMutex(c.Url)
	m.Lock()
	defer m.Unlock()

	switch filepath.Ext(name) {
	case ".yml", ".yaml":
		err := yaml.Unmarshal(c.Data, f)
		if err != nil {
			f.Data = string(c.Data)
		}
	case ".json":
		err := json.Unmarshal(c.Data, f)
		if err != nil {
			f.Data = string(c.Data)
		}
	case ".lua":
		if f.Data == nil {
			f.Data = script.New(name, c.Data)
		} else {
			script := f.Data.(*script.Script)
			script.Code = string(c.Data)
		}
	default:
		f.Data = string(c.Data)
	}

	if p, ok := f.Data.(common.Parser); ok {
		err := p.Parse(f, w)
		if err != nil {
			return nil, err
		}
	}

	return f, nil
}

func (w *ConfigWatcher) getFile(u *url.URL) *common.File {
	w.m.Lock()
	defer w.m.Unlock()

	f, ok := w.files[u.String()]
	if !ok {
		f = &common.File{Url: u}
		w.files[u.String()] = f
	}
	return f
}

func (w *ConfigWatcher) getMutex(u *url.URL) *sync.Mutex {
	w.m.Lock()
	defer w.m.Unlock()

	_, ok := w.fileMutex[u.String()]
	if !ok {
		w.fileMutex[u.String()] = &sync.Mutex{}
	}
	return w.fileMutex[u.String()]
}

func renderTemplate(b []byte) ([]byte, error) {
	content := string(b)

	funcMap := sprig.TxtFuncMap()
	funcMap["extractUsername"] = extractUsername
	tmpl := template.New("").Funcs(funcMap)

	tmpl, err := tmpl.Parse(content)
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, false)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func extractUsername(s string) string {
	slice := strings.Split(s, "\\")
	return slice[len(slice)-1]
}
