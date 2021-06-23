package dynamic

import (
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"mokapi/config/static"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ConfigWatcher struct {
	config  static.Providers
	watcher *fsnotify.Watcher
	fw      *FileWatcher
	configs map[string]Config

	listeners []func(c Config)
	stop      chan bool
}

func NewConfigWatcher(config static.Providers) *ConfigWatcher {
	return &ConfigWatcher{config: config, stop: make(chan bool)}
}

func (cw *ConfigWatcher) AddListener(listener func(c Config)) {
	cw.listeners = append(cw.listeners, listener)
}

func (cw *ConfigWatcher) Close() {
	cw.stop <- true
}

func (cw *ConfigWatcher) Start() error {
	stopFileWatcher := make(chan bool)
	update := make(chan Config)
	cw.fw = NewFileWatcher(update, stopFileWatcher)
	cw.fw.Start()

	if w, err := fsnotify.NewWatcher(); err != nil {
		return errors.Wrapf(err, "unable to start config watcher")
	} else {
		cw.watcher = w
	}

	if len(cw.config.File.Directory) > 0 {
		if err := filepath.Walk(cw.config.File.Directory, cw.walkDir); err != nil {
			log.Error(err)
		}
	} else if len(cw.config.File.Filename) > 0 {
		go cw.fw.add(cw.config.File.Filename)
	} else {
		log.Info("file provider: directory and filename empty")
		return nil
	}

	var gw *gitWatcher
	if len(cw.config.Git.Url) > 0 {
		dir, err := ioutil.TempDir("", "mokapi_git")
		if err != nil {
			return errors.Wrap(err, "unable to create temp dir for git provider")

		} else {
			log.Debugf("git temp directory: %v", dir)
		}

		cw.watcher.Add(dir)
		gw = newGitWatcher(dir, cw.config.Git)
		gw.Start()
	}

	var hw *httpWatcher
	if len(cw.config.Http.Url) > 0 {
		hw = newHttpWatcher(update, cw.config.Http)
		hw.Start()
	}

	ticker := time.NewTicker(time.Second)
	events := make([]fsnotify.Event, 0)

	go func() {
		defer func() {
			log.Debug("closing file watcher")
			ticker.Stop()
			err := cw.watcher.Close()
			if err != nil {
				log.Errorf("unable to close config watcher: %v", err.Error())
			}
			if gw != nil {
				gw.close <- true
			}
			if hw != nil {
				hw.close <- true
			}
		}()

		for {
			select {
			case <-cw.stop:
				stopFileWatcher <- true
				return
			case c := <-update:
				for _, listener := range cw.listeners {
					listener(c)
				}
			case evt := <-cw.watcher.Events:
				// temporary files ends with '~' in name
				if len(evt.Name) > 0 && !strings.HasSuffix(evt.Name, "~") {
					events = append(events, evt)
				}
			case <-ticker.C:
				m := make(map[string]bool)
				for _, evt := range events {
					if _, ok := m[evt.Name]; ok {
						continue
					}
					m[evt.Name] = true

					if b, err := isDir(evt.Name); err != nil {
						log.Errorf("unable to read event from %v: %v", evt.Name, err)
					} else if b && !skipPath(evt.Name) {
						if err := cw.watcher.Add(evt.Name); err != nil {
							log.Error(err)
						}
					} else if isValidConfigFile(evt.Name) {
						go cw.fw.add(evt.Name)
					}
				}

				events = make([]fsnotify.Event, 0)
			}
		}
	}()

	return nil
}

func (cw *ConfigWatcher) walkDir(path string, fi os.FileInfo, _ error) error {
	if fi.Mode().IsDir() {
		if skipPath(path) {
			return filepath.SkipDir
		}
		return cw.watcher.Add(path)
	} else if isValidConfigFile(path) {
		go cw.fw.add(path)
	}

	return nil
}
