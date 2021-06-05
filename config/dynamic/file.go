package dynamic

import (
	"encoding/json"
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"
)

type FileWatcher struct {
	Path    map[string]*fileHandler
	watcher *fsnotify.Watcher
	close   chan bool
	update  chan Config
	lock    sync.RWMutex
}

type fileHandler struct {
	f      func(string) (Config, error)
	events []ChangeEventHandler
}

func NewFileWatcher(update chan Config, close chan bool) *FileWatcher {
	return &FileWatcher{
		Path:   make(map[string]*fileHandler),
		close:  close,
		update: update,
		lock:   sync.RWMutex{},
	}
}

func (fw *FileWatcher) Read(path string, config Config, h ChangeEventHandler) error {
	fw.lock.Lock()
	defer fw.lock.Unlock()
	fh, ok := fw.Path[path]
	if !ok {
		fh = newFileHandler(config)
		fw.Path[path] = fh
		fw.watcher.Add(path)
	}
	fh.events = append(fh.events, h)
	v, err := fh.f(path)
	if err != nil {
		return err
	}

	vConfig := reflect.ValueOf(config).Elem()
	if vConfig.Kind() == reflect.Ptr {
		if vConfig.Type() == reflect.TypeOf(v) {
			vConfig.Set(reflect.ValueOf(v))
		} else {
			log.Debugf("TODO: FileWatcher.Read ** to *")
		}
	} else {
		vConfig.Set(reflect.ValueOf(v).Elem())
	}

	return nil
}

func (fw *FileWatcher) add(path string) {
	fw.lock.Lock()
	defer fw.lock.Unlock()

	if fh, ok := fw.Path[path]; !ok {
		ci := &configItem{}
		if err := loadFileConfig(path, ci); err != nil {
			log.Errorf("unable to read config %v: %v", path, err.Error())
		}
		if ci.item != nil {
			fh = newFileHandler(ci.item)
			fh.events = append(fh.events, ci.handler)
			fw.Path[path] = fh
			fw.watcher.Add(path)
			go func() {
				ci.handler(path, ci.item, fw)
				fw.update <- ci.item
			}()

		}
	}
}

func (fw *FileWatcher) Start() {
	if w, err := fsnotify.NewWatcher(); err != nil {
		log.Error("error creating file watcher", err)
		return
	} else {
		fw.watcher = w
	}

	ticker := time.NewTicker(time.Second)
	events := make([]fsnotify.Event, 0)

	go func() {
		defer func() {
			log.Error("Closing file watcher. Restart is required...")
			ticker.Stop()
			fw.watcher.Close()
		}()

		for {
			select {
			case <-fw.close:
				return
			case evt := <-fw.watcher.Events:
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

					log.Debugf("item change event received " + evt.Name)

					handler, ok := fw.Path[evt.Name]
					if !ok {
						log.Infof("No handler for '%v' found", evt.Name)
					}

					if config, err := handler.f(evt.Name); err != nil {
						log.Errorf("unable to read %v: %v", evt.Name, err.Error())
					} else {
						fw.onChanged(handler, evt.Name, config)
					}
				}

				events = make([]fsnotify.Event, 0)
			}
		}
	}()
}

func (fw *FileWatcher) onChanged(h *fileHandler, path string, config Config) {
	for _, e := range h.events {
		if b, c := e(path, config, fw); b {
			fw.update <- c
		}
	}
}

func newFileHandler(config interface{}) *fileHandler {
	val := reflect.ValueOf(config).Interface()
	return &fileHandler{f: func(path string) (Config, error) {
		err := loadFileConfig(path, val)
		return config, err
	}}
}

func loadFileConfig(filename string, element interface{}) error {
	log.Debugf("reading config %q", filename)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	switch filepath.Ext(filename) {
	case ".yml", ".yaml":
		err := yaml.Unmarshal(data, element)
		if err != nil {
			return errors.Wrapf(err, "parsing yaml file %s", filename)
		}
	case ".json":
		err := json.Unmarshal(data, element)
		if err != nil {
			return errors.Wrapf(err, "parsing json file %s", filename)
		}
	}

	return nil
}
