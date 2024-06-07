package file

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"io"
	"io/fs"
	"mokapi/config/dynamic"
	"mokapi/config/static"
	"mokapi/safe"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const mokapiIgnoreFile = ".mokapiignore"

var Bom = []byte{0xEF, 0xBB, 0xBF}

type Provider struct {
	cfg        static.FileProvider
	SkipPrefix []string
	watched    map[string]struct{}
	isInit     bool
	ignores    IgnoreFiles

	watcher *fsnotify.Watcher
	fs      FSReader
	ch      chan<- *dynamic.Config

	m sync.Mutex
}

func New(cfg static.FileProvider) *Provider {
	return NewWithWalker(cfg, &Reader{})
}

func NewWithWalker(cfg static.FileProvider, fs FSReader) *Provider {
	p := &Provider{
		cfg:        cfg,
		SkipPrefix: []string{"_"},
		watched:    make(map[string]struct{}),
		isInit:     true,
		ignores:    make(IgnoreFiles),
		fs:         fs,
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

func (p *Provider) Read(u *url.URL) (*dynamic.Config, error) {
	file := u.Path
	if len(u.Opaque) > 0 {
		file = u.Opaque
	}

	config, err := p.readFile(file)
	if err != nil {
		return nil, err
	}

	p.watchPath(file)
	return config, nil
}

func (p *Provider) Start(ch chan *dynamic.Config, pool *safe.Pool) error {
	p.ch = ch
	var path []string
	if len(p.cfg.Directories) > 0 {
		path = p.cfg.Directories

	} else if len(p.cfg.Filenames) > 0 {
		path = p.cfg.Filenames
	}
	if len(path) > 0 {
		pool.Go(func(ctx context.Context) {
			for _, pathItem := range path {
				for _, i := range strings.Split(pathItem, string(os.PathListSeparator)) {
					if err := p.walk(i); err != nil {
						log.Errorf("file provider: %v", err)
					}
				}
			}
			p.isInit = false
		})
	}
	return p.watch(pool)
}

func (p *Provider) Watch(dir string, pool *safe.Pool) {
	pool.Go(func(ctx context.Context) {
		if err := p.walk(dir); err != nil {
			log.Errorf("file provider: %v", err)
		}
	})
}

func (p *Provider) watch(pool *safe.Pool) error {
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
					fileInfo, err := p.fs.Stat(evt.Name)
					if err != nil {
						// skip
						continue
					}
					if !fileInfo.IsDir() {
						events = append(events, evt)
					} else if _, ok := p.watched[evt.Name]; !p.isInit && !ok {
						pool.Go(func(ctx context.Context) {
							err := p.walk(evt.Name)
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
					} else if len(p.cfg.Filenames) > 0 {
						for _, filename := range p.cfg.Filenames {
							if _, configFile := filepath.Split(filename); file == configFile {
								c, err := p.readFile(evt.Name)
								if err != nil {
									log.Errorf("unable to read file %v", evt.Name)
								}
								p.ch <- c
							}
						}
					} else {
						if !p.skip(evt.Name) {
							c, err := p.readFile(evt.Name)
							if err != nil {
								log.Errorf("unable to read file %v", evt.Name)
							}
							p.ch <- c
						}
					}
				}
				events = make([]fsnotify.Event, 0)
			}
		}
	})
	return nil
}

func (p *Provider) skip(path string) bool {
	if p.isWatchPath(path) {
		return false
	}
	if len(p.cfg.Include) > 0 {
		return !include(p.cfg.Include, path)
	}

	name := filepath.Base(path)
	if name == mokapiIgnoreFile {
		return true
	}
	for _, s := range p.SkipPrefix {
		if strings.HasPrefix(name, s) {
			return true
		}
	}
	return p.ignores.Match(path)
}

func (p *Provider) readFile(path string) (*dynamic.Config, error) {
	data, err := p.fs.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// remove bom sequence if present
	if len(data) >= 4 && bytes.Equal(data[0:3], Bom) {
		data = data[3:]
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

	stats, err := p.fs.Stat(path)
	if err != nil {
		return nil, err
	}

	return &dynamic.Config{
		Info: dynamic.ConfigInfo{
			Url: u, Provider: "file",
			Checksum: h.Sum(nil),
			Time:     stats.ModTime(),
		},
		Raw: data,
	}, nil
}

func (p *Provider) walk(root string) error {
	p.readMokapiIgnore(root)
	walkDir := func(path string, fi fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			if p.skip(path) && path != root {
				log.Debugf("skip dir: %v", path)
				return filepath.SkipDir
			}
			p.readMokapiIgnore(path)
			p.watchPath(path)
		} else if !p.skip(path) {
			if c, err := p.readFile(path); err != nil {
				log.Error(err)
			} else if len(c.Raw) > 0 {
				p.watchPath(path)
				p.ch <- c
			}
		} else {
			log.Debugf("skip file: %v", path)
		}

		return nil
	}

	return p.fs.Walk(root, walkDir)
}

func (p *Provider) readMokapiIgnore(path string) {
	f := filepath.Join(path, mokapiIgnoreFile)
	b, err := p.fs.ReadFile(f)
	if err != nil {
		return
	}
	if i, err := newIgnoreFile(f, b); err != nil {
		log.Errorf("unable to read file %v: %v", f, err)
	} else {
		key := filepath.Clean(path)
		p.ignores[key] = i
	}
}

func (p *Provider) watchPath(path string) {
	p.m.Lock()
	defer p.m.Unlock()

	if _, ok := p.watched[path]; ok {
		return
	}
	p.watched[path] = struct{}{}

	// add watcher to file does not work, see watcher.Add
	fileInfo, err := p.fs.Stat(path)
	if err != nil {
		return
	}
	if !fileInfo.IsDir() {
		path = filepath.Dir(path)

		if _, ok := p.watched[path]; ok {
			return
		}
	}

	p.watcher.Add(path)
}

func (p *Provider) isWatchPath(path string) bool {
	p.m.Lock()
	defer p.m.Unlock()

	_, ok := p.watched[path]
	return ok
}

func include(s []string, v string) bool {
	for _, i := range s {
		if Match(i, v) {
			return true
		}
	}
	return false
}
