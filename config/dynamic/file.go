package dynamic

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

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
	p.close <- true
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

	config := NewConfiguration()
	error := loadFileConfig(filename, config)
	if error != nil {
		log.WithFields(log.Fields{"file": filename, "error": error}).Error("Error loading configuration")
		return
	}

	if config.Ldap == nil && config.OpenApi == nil {
		log.Debugf("No expected configuration found in %v", filename)
		return
	}

	channel <- ConfigMessage{ProviderName: "file", Config: config, Key: filename}
}

func (p *FileProvider) addWatcher(directory string, channel chan<- ConfigMessage) {
	p.close = make(chan bool)

	watcher, error := fsnotify.NewWatcher()
	if error != nil {
		log.Error("Error creating file watcher", error)
	}

	error = watcher.Add(directory)
	if error != nil {
		log.WithField("watchItem", directory).Error("Error adding watcher")
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

					log.WithField("item", evt.Name).Info("Item change event received from " + directory)

					fi, error := os.Stat(evt.Name)
					if error != nil {
						log.WithFields(log.Fields{"item": evt.Name, "error": error}).Info("Error on watching item")
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

	err := yaml.Unmarshal(data, element)
	if err != nil {
		return errors.Wrapf(err, "parsing YAML file %s", filename)
	}
	return nil
}
