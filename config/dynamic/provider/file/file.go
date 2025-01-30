package file

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"io/fs"
	"math"
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
	} else {
		p.isInit = false
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
	var (
		waitFor      = 1 * time.Second
		mu           sync.Mutex
		timers       = make(map[string]*time.Timer)
		processEvent = func(evt fsnotify.Event) {
			defer func() {
				mu.Lock()
				delete(timers, evt.Name)
				mu.Unlock()
			}()

			fileInfo, err := p.fs.Stat(evt.Name)
			if err != nil {
				return
			}

			if evt.Has(fsnotify.Remove) || evt.Has(fsnotify.Rename) {
				p.m.Lock()
				defer p.m.Unlock()

				if !fileInfo.IsDir() {
					delete(p.watched, evt.Name)
					return
				}
				for watched := range p.watched {
					if strings.HasPrefix(watched, evt.Name) {
						delete(p.watched, watched)
						if watched == evt.Name {
							err = p.watcher.Remove(watched)
							if err != nil && !errors.Is(err, fsnotify.ErrNonExistentWatch) {
								log.Errorf("remove watcher '%v' failed: %v", evt.Name, err)
							}
						}
					}
				}
				return
			}

			if fileInfo.IsDir() && !p.isInit {
				pool.Go(func(ctx context.Context) {
					err := p.walk(evt.Name)
					if err != nil {
						log.Errorf("unable to process dir %v: %v", evt.Name, err)
					}
				})
				return
			}

			dir, file := filepath.Split(evt.Name)
			if dir == evt.Name && !p.skip(dir, true) {
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
				if !p.skip(evt.Name, false) {
					c, err := p.readFile(evt.Name)
					if err != nil {
						log.Errorf("unable to read file %v", evt.Name)
					}
					p.ch <- c
				}
			}
		}
	)

	pool.Go(func(ctx context.Context) {
		defer func() {
			p.watcher.Close()
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case err, ok := <-p.watcher.Errors:
				if !ok {
					return
				}
				log.Errorf("file watcher error: %s", err)
			case evt, ok := <-p.watcher.Events:
				if !ok {
					return
				}

				mu.Lock()
				t, ok := timers[evt.Name]
				mu.Unlock()

				if !ok {
					t = time.AfterFunc(math.MaxInt64, func() { processEvent(evt) })
					t.Stop()

					mu.Lock()
					timers[evt.Name] = t
					mu.Unlock()
				}

				t.Reset(waitFor)
			}
		}
	})
	return nil
}

func (p *Provider) skip(path string, isDir bool) bool {
	if p.isWatchPath(path) {
		return false
	}

	if !isDir && len(p.cfg.Include) > 0 {
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
			if p.skip(path, true) && path != root {
				log.Debugf("skip dir: %v", path)
				return filepath.SkipDir
			}
			p.readMokapiIgnore(path)
			p.watchPath(path)
		} else if !p.skip(path, false) {
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
