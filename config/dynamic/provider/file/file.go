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
	"math"
	"mokapi/config/dynamic"
	"mokapi/config/static"
	"mokapi/safe"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
)

const mokapiIgnoreFile = ".mokapiignore"

var Bom = []byte{0xEF, 0xBB, 0xBF}

type watch struct {
	isDir bool
}

type Provider struct {
	cfg        static.FileProvider
	SkipPrefix []string
	watched    map[string]watch
	isInit     bool
	ignores    IgnoreFiles

	watcher *fsnotify.Watcher
	fs      FSReader
	ch      chan<- dynamic.ConfigEvent

	m sync.Mutex
}

func New(cfg static.FileProvider) *Provider {
	return NewWithWalker(cfg, &Reader{})
}

func NewWithWalker(cfg static.FileProvider, fs FSReader) *Provider {
	p := &Provider{
		cfg:        cfg,
		SkipPrefix: []string{"_"},
		watched:    make(map[string]watch),
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

func (p *Provider) Start(ch chan dynamic.ConfigEvent, pool *safe.Pool) error {
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
		mu     sync.Mutex
		t      *time.Timer
		events []fsnotify.Event
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
				events = append(events, evt)
				if t == nil {
					t = time.AfterFunc(math.MaxInt64, func() {
						mu.Lock()
						e := events
						events = nil
						t = nil
						mu.Unlock()
						p.processEvents(e)
					})
				}
				t.Reset(time.Second)
				mu.Unlock()
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
				p.ch <- dynamic.ConfigEvent{Event: dynamic.Create, Config: c, Name: path}
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

	// add watcher to file does not work, see watcher.Add
	fileInfo, err := p.fs.Stat(path)
	if err != nil {
		return
	}
	if !fileInfo.IsDir() {
		p.watched[path] = watch{isDir: false}

		path = filepath.Dir(path)
		p.watched[path] = watch{isDir: true}
	} else {
		p.watched[path] = watch{isDir: true}
	}

	_ = p.watcher.Add(path)
}

func (p *Provider) processEvents(events []fsnotify.Event) {
	done := map[string]bool{}
	walkList := []string{}

	for _, evt := range events {
		key := fmt.Sprintf("%v:%v", evt.Op, evt.Name)
		if _, ok := done[key]; ok {
			continue
		}
		if evt.Op == fsnotify.Write {
			// skip write event if we have already a create event.
			key = fmt.Sprintf("%v:%v", fsnotify.Create, evt.Name)
			if _, ok := done[key]; ok {
				continue
			}
		}
		done[key] = true

		if evt.Has(fsnotify.Remove) || evt.Has(fsnotify.Rename) {
			p.m.Lock()
			p.m.Unlock()

			if w, ok := p.watched[evt.Name]; ok {
				if !w.isDir {
					e := dynamic.ConfigEvent{Event: dynamic.Delete, Name: evt.Name}
					p.ch <- e
				}

				delete(p.watched, evt.Name)
			}
			continue
		}

		fileInfo, err := p.fs.Stat(evt.Name)
		if err != nil {
			continue
		}

		if fileInfo.IsDir() && !p.isInit {
			walkList = append(walkList, evt.Name)
			continue
		}

		e := dynamic.ConfigEvent{Name: evt.Name}
		if evt.Has(fsnotify.Create) {
			e.Event = dynamic.Create
		} else if evt.Has(fsnotify.Write) {
			e.Event = dynamic.Update
		}

		dir, _ := filepath.Split(evt.Name)
		if dir == evt.Name && !p.skip(dir, true) {
			p.watchPath(dir)
		} else {
			if !p.skip(evt.Name, false) {
				e.Config, err = p.readFile(evt.Name)
				if err != nil {
					log.Errorf("unable to read file %v", evt.Name)
				}
				p.ch <- e
			}
		}
	}

	slices.SortFunc(walkList, func(a, b string) int {
		return len(a) - len(b)
	})
	var doneWalk []string
Walk:
	for _, dir := range walkList {
		for _, d := range doneWalk {
			if strings.HasPrefix(dir, d) {
				continue Walk
			}
		}
		doneWalk = append(doneWalk, dir)

		err := p.walk(dir)
		if err != nil {
			log.Errorf("unable to process dir %v: %v", dir, err)
		}
	}
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
