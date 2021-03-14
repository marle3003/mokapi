package dynamic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Masterminds/sprig"
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"mokapi/config/static"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"text/template"
	"time"
)

type ChangeEventHandler func(path string, c Config, r ConfigReader) (bool, Config)

type Config interface {
}

type ConfigReader interface {
	Read(path string, c Config, h ChangeEventHandler) error
}

type fileHandler struct {
	f      func(string) (Config, error)
	events []ChangeEventHandler
}

type FileWatcher struct {
	Path    map[string]*fileHandler
	watcher *fsnotify.Watcher
	close   chan bool
	update  chan Config
	lock    sync.RWMutex
}

type configType struct {
	header  string
	config  reflect.Type
	handler ChangeEventHandler
}

var (
	configTypes      []configType
	NullEventHandler = func(path string, c Config, r ConfigReader) (bool, Config) { return false, nil }
)

func NewEmptyEventHandler(parent Config) ChangeEventHandler {
	return func(path string, c Config, r ConfigReader) (bool, Config) { return true, parent }
}

func Register(header string, c Config, h ChangeEventHandler) {
	val := reflect.ValueOf(c).Elem()
	configTypes = append(configTypes, configType{header, val.Type(), h})
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
		vConfig.Set(reflect.ValueOf(v))
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

func NewFileWatcher(update chan Config, close chan bool) *FileWatcher {
	return &FileWatcher{
		Path:   make(map[string]*fileHandler),
		close:  close,
		update: update,
		lock:   sync.RWMutex{},
	}
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
	return &ConfigWatcher{config: config}
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

	if err := filepath.Walk(cw.config.File.Directory, cw.walkDir); err != nil {
		fmt.Println("ERROR", err)
	}

	ticker := time.NewTicker(time.Second)
	events := make([]fsnotify.Event, 0)

	go func() {
		defer func() {
			log.Error("Closing config watcher. Restart is required...")
			ticker.Stop()
			cw.watcher.Close()
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
						cw.watcher.Add(evt.Name)
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

type configItem struct {
	handler ChangeEventHandler
	item    Config
}

func (ci *configItem) UnmarshalYAML(unmarshal func(interface{}) error) error {
	data := make(map[string]string)
	unmarshal(data)

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
	case ".yml", ".yaml", ".json":
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

func loadFileConfig(filename string, element interface{}) error {
	log.Infof("reading config %q", filename)
	data, error := ioutil.ReadFile(filename)
	if error != nil {
		return error
	}

	content := string(data)

	funcMap := sprig.TxtFuncMap()
	funcMap["extractUsername"] = extractUsername
	tmpl := template.New(filename).Funcs(funcMap)

	_, error = tmpl.Parse(content)
	if error != nil {
		return error
	}

	var buffer bytes.Buffer
	error = tmpl.Execute(&buffer, false)
	if error != nil {
		return error
	}

	renderedTemplate := buffer.Bytes()

	switch filepath.Ext(filename) {
	case ".yml", ".yaml":
		err := yaml.Unmarshal(renderedTemplate, element)
		if err != nil {
			return errors.Wrapf(err, "parsing yaml file %s", filename)
		}
	case ".json":
		err := json.Unmarshal(renderedTemplate, element)
		if err != nil {
			return errors.Wrapf(err, "parsing json file %s", filename)
		}
	}

	return nil
}

func extractUsername(s string) string {
	slice := strings.Split(s, "\\")
	return slice[len(slice)-1]
}
