package dynamic

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"mokapi/config/static"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

var (
	configTypes []configType
)

type configType struct {
	header  string
	config  reflect.Type
	handler ChangeEventHandler
}

type configItem struct {
	handler ChangeEventHandler
	item    Config
}

func NewEmptyEventHandler(parent Config) ChangeEventHandler {
	return func(path string, c Config, r ConfigReader) (bool, Config) { return true, parent }
}

func Register(header string, c Config, h ChangeEventHandler) {
	val := reflect.ValueOf(c).Elem()
	configTypes = append(configTypes, configType{header, val.Type(), h})
}

func isDir(path string) (bool, error) {
	if fi, err := os.Stat(path); err != nil {
		return false, err
	} else if fi.IsDir() {
		return true, nil
	}
	return false, nil
}

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
			fmt.Println("ERROR", err)
		}
	} else if len(cw.config.File.Filename) > 0 {
		go cw.fw.add(cw.config.File.Filename)
	} else {
		log.Info("file provider: directory and filename empty")
		return nil
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
		}()

		for {
			select {
			case <-cw.stop:
				stopFileWatcher <- true
				return
			case c := <-update:
				cw.onConfigChanged(c)
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
						if err := filepath.Walk(cw.config.File.Directory, cw.walkDir); err != nil {
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

func (cw *ConfigWatcher) onConfigChanged(c Config) {
	for _, listener := range cw.listeners {
		listener(c)
	}
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

func (ci *configItem) UnmarshalYAML(unmarshal func(interface{}) error) error {
	data := make(map[string]string)
	_ = unmarshal(data)

	for _, c := range configTypes {
		if _, ok := data[c.header]; ok {
			ci.item = reflect.New(c.config).Interface()
			ci.handler = c.handler
			err := unmarshal(ci.item)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func isValidConfigFile(path string) bool {
	if skipPath(path) {
		return false
	}
	switch filepath.Ext(path) {
	case ".yml", ".yaml", ".json", ".tmpl":
		return true
	default:
		return false
	}
}

func skipPath(path string) bool {
	name := filepath.Base(path)
	// TODO: make skip char configurable
	if strings.HasPrefix(name, "_") {
		log.Infof("skipping config %v", name)
		return true
	}
	return false
}
