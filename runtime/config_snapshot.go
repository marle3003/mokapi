package runtime

import (
	"mokapi/config/dynamic"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
)

// configSnapshot holds the files of one service (key: file URL) and the immutable config
// merged from them. A reader takes one snapshot via Config and uses only that. The zero value
// is ready to use.
//
// Writers (add, remove) must be serialised by the caller, in practice by the store lock of the
// protocol. The lock m only protects the file map for concurrent Configs calls. merge runs
// outside of m, so a slow merge never blocks readers of Configs.
type configSnapshot[T any] struct {
	config  atomic.Pointer[T]
	m       sync.RWMutex               // guards configs
	configs map[string]*dynamic.Config // key: Info.Url.String()
}

// Config returns the current merged config, or nil after the last file was removed.
// The result must not be modified.
func (s *configSnapshot[T]) Config() *T {
	return s.config.Load()
}

// Configs returns a copy of the current files.
func (s *configSnapshot[T]) Configs() []*dynamic.Config {
	s.m.RLock()
	defer s.m.RUnlock()

	var r []*dynamic.Config
	for _, config := range s.configs {
		r = append(r, config)
	}
	return r
}

// publish stores cfg as is, without merging files.
func (s *configSnapshot[T]) publish(cfg *T) {
	s.config.Store(cfg)
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
	s.config.Store(cfg)
	return cfg
}
