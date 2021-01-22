package dynamic

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig"
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type FileProvider struct {
	Filename  string
	Directory string

	close chan bool
}

func (p *FileProvider) ProvideService(channel chan<- ConfigMessage) {
	p.loadService(channel)
}

func (p *FileProvider) Close() {
	if p.close != nil {
		p.close <- true
	}
}

func (p *FileProvider) loadService(channel chan<- ConfigMessage) {
	if len(p.Filename) > 0 {
		f, error := filepath.Abs(p.Filename)
		if error != nil {
			log.WithField("filename", p.Filename).Error("Can not resolve filepath")
			return
		}
		p.loadServiceFromFile(f, channel)
		p.addWatcher(f, channel)
	} else {
		d, error := filepath.Abs(p.Directory)
		if error != nil {
			log.WithField("directory", p.Directory).Error("Can not resolve directory")
			return
		}
		p.loadServiceFromDirectory(d, channel)
	}
}

func (p *FileProvider) loadServiceFromDirectory(directory string, channel chan<- ConfigMessage) {
	if skip(directory) {
		return
	}
	fileList, err := ioutil.ReadDir(directory)
	if err != nil {
		log.Error("unable to read directory %s: %v", directory, err)
	}

	for _, item := range fileList {
		if item.IsDir() {
			p.loadServiceFromDirectory(filepath.Join(directory, item.Name()), channel)
			continue
		}
		p.loadServiceFromFile(filepath.Join(directory, item.Name()), channel)
	}

	p.addWatcher(directory, channel)
}

func (p *FileProvider) loadServiceFromFile(filename string, channel chan<- ConfigMessage) {
	if skip(filename) {
		return
	}
	switch filepath.Ext(filename) {
	case ".yml", ".yaml", ".json":
		// continue loading from file
	default:
		return
	}

	config := NewConfigurationItem()
	error := loadFileConfig(filename, config)
	if error != nil {
		log.WithFields(log.Fields{"file": filename, "error": error}).Error("error loading configuration")
		return
	}

	if config.Ldap == nil && config.OpenApi == nil {
		log.Debugf("no expected configuration found in %v", filename)
		return
	}

	channel <- ConfigMessage{ProviderName: "file", Config: config, Key: filename}
}

func (p *FileProvider) addWatcher(directory string, channel chan<- ConfigMessage) {
	p.close = make(chan bool)

	watcher, error := fsnotify.NewWatcher()
	if error != nil {
		log.Error("error creating file watcher", error)
	}

	error = watcher.Add(directory)
	if error != nil {
		log.WithField("watchItem", directory).Error("error adding watcher")
	}

	ticker := time.NewTicker(time.Second)
	events := make([]fsnotify.Event, 0)

	go func() {
		defer func() {
			ticker.Stop()
			watcher.Close()
		}()

		for {
			select {
			case <-p.close:
				return
			case evt := <-watcher.Events:
				if len(evt.Name) > 0 {
					events = append(events, evt)
				}
			case <-ticker.C:
				m := make(map[string]struct{})
				for _, evt := range events {
					if _, ok := m[evt.Name]; ok {
						continue
					}
					m[evt.Name] = struct{}{}

					log.WithField("item", evt.Name).Debugf("item change event received from " + directory)

					fi, error := os.Stat(evt.Name)
					if error != nil {
						log.WithFields(log.Fields{"item": evt.Name, "error": error}).Info("error on watching item")
						return
					}
					switch mode := fi.Mode(); {
					case mode.IsDir():
						p.loadServiceFromDirectory(evt.Name, channel)
					case mode.IsRegular():
						p.loadServiceFromFile(evt.Name, channel)
					}
				}

				events = make([]fsnotify.Event, 0)
			}
		}
	}()
}

func loadFileConfig(filename string, element interface{}) error {
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

func skip(path string) bool {
	name := filepath.Base(path)
	// TODO: make skip char configurable
	if strings.HasPrefix(name, "_") {
		log.Infof("skipping config %v", name)
		return true
	}
	return false
}
