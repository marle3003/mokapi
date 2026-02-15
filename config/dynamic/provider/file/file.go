package file

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
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

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
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
	config.Info.Url = u
	config.SourceType = dynamic.SourceReference

	p.watchPath(file)
	return config, nil
}

func (p *Provider) Start(ch chan dynamic.ConfigEvent, pool *safe.Pool) error {
	p.ch = ch
	var files []static.FileConfig
	if len(p.cfg.Directories) > 0 {
		for _, dir := range p.cfg.Directories {
			for _, d := range strings.Split(dir.Path, string(os.PathListSeparator)) {
				files = append(files, static.FileConfig{Path: d, Include: dir.Include, Exclude: dir.Exclude})
			}
		}

	} else if len(p.cfg.Filenames) > 0 {
		for _, file := range p.cfg.Filenames {
			for _, f := range strings.Split(file, string(os.PathListSeparator)) {
				files = append(files, static.FileConfig{Path: f})
			}
		}
	}
	if len(files) > 0 {
		pool.Go(func(ctx context.Context) {
			for _, file := range files {
				if err := p.walk(file); err != nil {
					log.Errorf("file provider: %v", err)
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
		if err := p.walk(static.FileConfig{Path: dir}); err != nil {
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
			_ = p.watcher.Close()
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

func (p *Provider) skip(path string, isDir bool, info static.FileConfig) bool {
	if p.isWatchPath(path) {
		return false
	}

	if !isDir {
		inc := p.cfg.Include
		if len(info.Include) > 0 {
			inc = append(inc, info.Include...)
		}
		if len(inc) > 0 {
			if !include(inc, path) {
				return true
			}
		}
		ex := p.cfg.Exclude
		if len(info.Exclude) > 0 {
			ex = append(ex, info.Exclude...)
		}
		if len(ex) > 0 {
			if include(ex, path) {
				return true
			}
		}
	}

	if isMokapiIgnoreFile(path) {
		return true
	}

	name := filepath.Base(path)
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

	u, err := ParseUrl(path)
	if err != nil {
		return nil, err
	}
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

func (p *Provider) walk(fileInfo static.FileConfig) error {
	p.readMokapiIgnore(fileInfo.Path)
	walkDir := func(path string, fi fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			if p.skip(path, true, fileInfo) && path != fileInfo.Path {
				log.Debugf("skip dir: %v", path)
				return filepath.SkipDir
			}
			p.readMokapiIgnore(path)
			p.watchPath(path)
		} else if !p.skip(path, false, fileInfo) {
			if c, err := p.readFile(path); err != nil {
				log.Error(err)
			} else if len(c.Raw) > 0 {
				p.watchPath(path)
				p.ch <- dynamic.ConfigEvent{Event: dynamic.Create, Config: c, Name: path}
			}
		} else if !isMokapiIgnoreFile(path) {
			log.Debugf("skip file: %v", path)
		}

		return nil
	}

	return p.fs.Walk(fileInfo.Path, walkDir)
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
	var walkList []string

	for _, evt := range events {
		key := fmt.Sprintf("%v:%v", evt.Op, evt.Name)
		if _, ok := done[key]; ok {
			continue
		}
		if evt.Op == fsnotify.Write {
			// skip write event if we already have a create event.
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
		} else if evt.Has(fsnotify.Chmod) {
			e.Event = dynamic.Update
		}

		dir, _ := filepath.Split(evt.Name)
		isDir := dir == evt.Name

		if !p.skipEvent(e, isDir) {
			if isDir {
				p.watchPath(dir)
			} else {
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

		cfg, err := p.getFileConfig(dir)
		if err != nil {
			log.Debugf("skip event: unable to get file config for %v: %v", dir, err)
		}

		cfg.Path = dir
		err = p.walk(cfg)
		if err != nil {
			log.Errorf("unable to process dir %v: %v", dir, err)
		}
	}
}

func (p *Provider) skipEvent(evt dynamic.ConfigEvent, isDir bool) bool {
	if p.isWatchPath(evt.Name) {
		return false
	}
	cfg, err := p.getFileConfig(evt.Name)
	if err != nil {
		log.Debugf("skip event: unable to get file config for %v: %v", evt.Name, err)
	}

	return p.skip(evt.Name, isDir, cfg)
}

func (p *Provider) isWatchPath(path string) bool {
	p.m.Lock()
	defer p.m.Unlock()

	_, ok := p.watched[path]
	return ok
}

func (p *Provider) getFileConfig(path string) (static.FileConfig, error) {
	for _, cfg := range p.cfg.Directories {
		if isSub(cfg.Path, path) {
			return cfg, nil
		}
	}
	return static.FileConfig{}, errors.New("directory config not found")
}

func include(s []string, v string) bool {
	for _, i := range s {
		if Match(i, v) {
			return true
		}
	}
	return false
}

func isMokapiIgnoreFile(path string) bool {
	name := filepath.Base(path)
	return name == mokapiIgnoreFile
}

func isSub(parent, sub string) bool {
	up := ".." + string(os.PathSeparator)
	rel, err := filepath.Rel(parent, sub)
	if err != nil {
		return false
	}
	if !strings.HasPrefix(rel, up) && rel != ".." {
		return true
	}
	return false
}
