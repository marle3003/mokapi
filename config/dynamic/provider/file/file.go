package file

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"mokapi/config/dynamic"
	"mokapi/config/static"
	"mokapi/safe"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
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
	// aliases maps the resolved path of a watched symlinked directory to the paths it was reached by.
	aliases  map[string][]string
	isInit   atomic.Bool
	ignores  IgnoreFiles
	ignoresM sync.RWMutex

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
		aliases:    make(map[string][]string),
		ignores:    make(IgnoreFiles),
		fs:         fs,
	}
	p.isInit.Store(true)

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
				if ctx.Err() != nil {
					return
				}
				if err := p.walk(ctx, file); err != nil {
					log.Errorf("file provider: %v", err)
				}
			}
			p.isInit.Store(false)
		})
	} else {
		p.isInit.Store(false)
	}
	return p.watch(pool)
}

func (p *Provider) Watch(dir string, pool *safe.Pool) {
	pool.Go(func(ctx context.Context) {
		if err := p.walk(ctx, static.FileConfig{Path: dir}); err != nil {
			log.Errorf("file provider: %v", err)
		}
	})
}

func (p *Provider) watch(pool *safe.Pool) error {
	pool.Go(func(ctx context.Context) {
		var (
			events   []fsnotify.Event
			debounce *time.Timer
			fire     <-chan time.Time
		)
		defer func() {
			if debounce != nil {
				debounce.Stop()
			}
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

				events = append(events, evt)
				if debounce != nil {
					debounce.Stop()
				}
				debounce = time.NewTimer(time.Second)
				fire = debounce.C
			case <-fire:
				batch := events
				events, debounce, fire = nil, nil, nil
				// processed inside the loop so that pool.Stop waits for it and batches never overlap
				p.processEvents(ctx, batch)
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
		inc := slices.Concat(p.cfg.Include, info.Include)
		if len(inc) > 0 {
			if !include(inc, path) {
				return true
			}
		}
		ex := slices.Concat(p.cfg.Exclude, info.Exclude)
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

	p.ignoresM.RLock()
	defer p.ignoresM.RUnlock()
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

// walk stops without an error when ctx is done, the remaining files are dropped on shutdown.
func (p *Provider) walk(ctx context.Context, fileInfo static.FileConfig) error {
	p.readMokapiIgnore(fileInfo.Path)
	walkDir := func(path string, fi fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}

		relPath, _ := filepath.Rel(fileInfo.Path, path)
		if fi.IsDir() {
			if p.skip(relPath, true, fileInfo) && path != fileInfo.Path {
				log.Debugf("skip dir: %v", path)
				return filepath.SkipDir
			}
			p.readMokapiIgnore(path)
			p.watchPath(path)
		} else if !p.skip(relPath, false, fileInfo) {
			if c, err := p.readFile(path); err != nil {
				log.Error(err)
			} else if len(c.Raw) > 0 {
				p.watchPath(path)
				if !safe.Send(ctx, p.ch, dynamic.ConfigEvent{Event: dynamic.Create, Config: c, Name: path}) {
					return ctx.Err()
				}
			}
		} else if !isMokapiIgnoreFile(path) {
			log.Debugf("skip file: %v", path)
		}

		return nil
	}

	err := p.fs.Walk(fileInfo.Path, walkDir)
	if ctx.Err() != nil {
		return nil
	}
	return err
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
		p.ignoresM.Lock()
		p.ignores[key] = i
		p.ignoresM.Unlock()
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
	p.watchResolvedPath(path)
}

// watchResolvedPath additionally watches the resolved path of a symlinked directory, because
// a watch reports its events under one path only. When a directory is reachable via several
// links, the watcher might report it under a different one than the link we loaded the file from.
func (p *Provider) watchResolvedPath(path string) {
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil || resolved == path {
		return
	}
	if !slices.Contains(p.aliases[resolved], path) {
		p.aliases[resolved] = append(p.aliases[resolved], path)
	}
	if err := p.watcher.Add(resolved); err != nil {
		log.Debugf("unable to watch resolved path %v of %v: %v", resolved, path, err)
	}
}

// unresolve returns the paths a provider knows the file of an event by. A watch reports its events
// under one path only (inotify: the path added first), so the event is expanded to the same file under
// every link and the resolved path of the group it belongs to, as far as they are watched.
func (p *Provider) unresolve(path string) []string {
	p.m.Lock()
	defer p.m.Unlock()

	names := []string{path}
	for resolved, links := range p.aliases {
		members := append([]string{resolved}, links...)
		for _, member := range members {
			if !isSub(member, path) {
				continue
			}
			rel, _ := filepath.Rel(member, path)
			if rel == "." && isSymlinkTo(member, resolved) {
				// the symlink itself, e.g. removed: it does not concern the other links
				break
			}
			for _, m := range members {
				name := filepath.Join(m, rel)
				if p.isWatchedLocked(name) && !slices.Contains(names, name) {
					names = append(names, name)
				}
			}
			break
		}
	}
	if len(names) > 1 && !p.isWatchedLocked(path) {
		return names[1:]
	}
	return names
}

// isSymlinkTo reports whether path is a symlink that led to resolved, as opposed to a directory that is
// only reachable through the symlink of one of its parents. The path might be removed already, so the
// answer is derived from its parent.
func isSymlinkTo(path, resolved string) bool {
	parent, err := filepath.EvalSymlinks(filepath.Dir(path))
	return err == nil && filepath.Join(parent, filepath.Base(path)) != resolved
}

func (p *Provider) isWatchedLocked(path string) bool {
	_, isWatched := p.watched[path]
	_, isDirWatched := p.watched[filepath.Dir(path)]
	return isWatched || isDirWatched
}

// unwatchAlias forgets a removed link, the links below it and stops watching a resolved path when no other
// link uses it. inotify shares one watch between all paths to a directory, so removing the watch of a link
// also drops it for the remaining paths: they are added again.
func (p *Provider) unwatchAlias(link string) {
	p.m.Lock()
	defer p.m.Unlock()

	for resolved, links := range p.aliases {
		remaining := slices.DeleteFunc(slices.Clone(links), func(l string) bool { return isSub(link, l) })
		if len(remaining) == len(links) {
			continue
		}
		for _, l := range links {
			if isSub(link, l) {
				_ = p.watcher.Remove(l)
			}
		}
		maps.DeleteFunc(p.watched, func(path string, _ watch) bool { return isSub(link, path) })

		if len(remaining) == 0 {
			delete(p.aliases, resolved)
		} else {
			p.aliases[resolved] = remaining
		}
		if _, isWatched := p.watched[resolved]; len(remaining) > 0 || isWatched {
			for _, path := range slices.Concat(remaining, []string{resolved}) {
				_ = p.watcher.Add(path)
			}
		} else {
			_ = p.watcher.Remove(resolved)
		}
	}
}

func (p *Provider) unwatch(path string) (watch, bool) {
	p.m.Lock()
	defer p.m.Unlock()

	w, ok := p.watched[path]
	delete(p.watched, path)
	return w, ok
}

func (p *Provider) processEvents(ctx context.Context, events []fsnotify.Event) {
	done := map[string]bool{}
	var walkList []string

	var expanded []fsnotify.Event
	for _, evt := range events {
		for _, name := range p.unresolve(evt.Name) {
			expanded = append(expanded, fsnotify.Event{Name: name, Op: evt.Op})
		}
	}

	for _, evt := range expanded {
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
			p.unwatchAlias(evt.Name)
			if w, ok := p.unwatch(evt.Name); ok && !w.isDir {
				e := dynamic.ConfigEvent{Event: dynamic.Delete, Name: evt.Name}
				if !safe.Send(ctx, p.ch, e) {
					return
				}
			}
			continue
		}

		fileInfo, err := p.fs.Stat(evt.Name)
		if err != nil {
			continue
		}

		if fileInfo.IsDir() && !p.isInit.Load() {
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
				if !safe.Send(ctx, p.ch, e) {
					return
				}
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
		err = p.walk(ctx, cfg)
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
