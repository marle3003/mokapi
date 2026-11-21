package runtime

import (
	"mokapi/config/dynamic"
	"path/filepath"
	"slices"
	"sort"
	"sync"
	"sync/atomic"
)

// configSnapshot holds the files of one service (key: file URL) and the immutable config
// merged from them. Config and files are published together as one pair, so a reader that takes
// one Snapshot always sees the config together with the files it was merged from. The zero value
// is ready to use.
//
// Writers (add, remove) must be serialised by the caller, in practice by the store lock of the
// protocol. The lock m only guards the file map among writers; readers never take it, and merge
// runs outside of m, so a slow merge never blocks readers.
type configSnapshot[T any] struct {
	current atomic.Pointer[snapshot[T]]
	m       sync.Mutex                 // guards configs
	configs map[string]*dynamic.Config // key: Info.Url.String()
}

// snapshot is the immutable pair that is published atomically.
type snapshot[T any] struct {
	config *T
	files  []*dynamic.Config
}

// Config returns the current merged config, or nil after the last file was removed.
// The result must not be modified.
func (s *configSnapshot[T]) Config() *T {
	cfg, _ := s.Snapshot()
	return cfg
}

// Snapshot returns the current merged config and the files it was merged from, taken from one
// consistent state. The config is nil after the last file was removed. The files are a copy;
// the config must not be modified.
func (s *configSnapshot[T]) Snapshot() (*T, []*dynamic.Config) {
	cur := s.current.Load()
	if cur == nil {
		return nil, nil
	}
	return cur.config, slices.Clone(cur.files)
}

// publish stores cfg as is, without merging files.
func (s *configSnapshot[T]) publish(cfg *T) {
	s.current.Store(&snapshot[T]{config: cfg})
}

// add adds or replaces the file and publishes the result of merge, which receives all files
// sorted by base name. It returns the published config.
func (s *configSnapshot[T]) add(f *dynamic.Config, merge func(files []*dynamic.Config) *T) *T {
	s.m.Lock()
	if s.configs == nil {
		s.configs = map[string]*dynamic.Config{}
	}
	s.configs[f.Info.Url.String()] = f
	files := s.sortedFiles()
	s.m.Unlock()

	return s.publishMerged(files, merge)
}

// remove removes the file and publishes the result of merge, or nil if no file is left.
// It returns the published config.
func (s *configSnapshot[T]) remove(f *dynamic.Config, merge func(files []*dynamic.Config) *T) *T {
	s.m.Lock()
	delete(s.configs, f.Info.Url.String())
	files := s.sortedFiles()
	s.m.Unlock()

	return s.publishMerged(files, merge)
}

// sortedFiles must be called with m held.
func (s *configSnapshot[T]) sortedFiles() []*dynamic.Config {
	files := make([]*dynamic.Config, 0, len(s.configs))
	for _, f := range s.configs {
		files = append(files, f)
	}
	sort.Slice(files, func(i, j int) bool {
		return filepath.Base(files[i].Info.Url.String()) < filepath.Base(files[j].Info.Url.String())
	})
	return files
}

func (s *configSnapshot[T]) publishMerged(files []*dynamic.Config, merge func(files []*dynamic.Config) *T) *T {
	var cfg *T
	if len(files) > 0 {
		cfg = merge(files)
	}
	s.current.Store(&snapshot[T]{config: cfg, files: files})
	return cfg
}
