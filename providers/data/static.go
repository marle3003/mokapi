package data

import (
	"fmt"
	"io/ioutil"
	"mokapi/service"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type StaticDataProvider struct {
	Path string
	data map[interface{}]interface{}
	stop chan bool
}

func NewStaticDataProvider(path string) *StaticDataProvider {
	provider := &StaticDataProvider{Path: path, stop: make(chan bool)}
	provider.init()
	return provider
}

func (provider *StaticDataProvider) Provide(parameters map[string]string, schema *service.Schema) (interface{}, error) {
	data := provider.getData(schema.Resource)
	data = filterData(data, parameters)
	return data, nil
}

func (p *StaticDataProvider) init() {
	go func() {
		p.data = make(map[interface{}]interface{})
		p.loadData()

		p.addWatcher()
	}()
}

func (p *StaticDataProvider) loadData() {
	fi, error := os.Stat(p.Path)
	if error != nil {
		log.WithFields(log.Fields{"path": p.Path, "error": error}).Info("Error in static data provider")
		return
	}
	switch mode := fi.Mode(); {
	case mode.IsDir():
		p.readDirectory(p.Path)
	case mode.IsRegular():
		p.readFile(p.Path)
	}
}

func (p *StaticDataProvider) Close() {
	p.stop <- true
}

func (p *StaticDataProvider) readDirectory(directory string) {
	fileList, err := ioutil.ReadDir(directory)
	if err != nil {
		log.Error("unable to read directory %s: %v", directory, err)
	}

	for _, item := range fileList {
		if item.IsDir() {
			p.readDirectory(filepath.Join(directory, item.Name()))
			continue
		}
		p.readFile(filepath.Join(directory, item.Name()))
	}
}

func (p *StaticDataProvider) readFile(file string) {
	newData := parseFile(file)
	for k, v := range newData {
		p.data[k] = v
	}
}

func (p *StaticDataProvider) addWatcher() {
	watcher, error := fsnotify.NewWatcher()
	if error != nil {
		log.Error("Error creating file watcher", error)
	}

	error = watcher.Add(p.Path)
	if error != nil {
		log.WithField("watchItem", p.Path).Error("Error adding watcher")
	}

	go func() {
		defer func() {
			log.Debug("Closing StaticDataProvider")
			watcher.Close()
		}()
		for {
			select {
			case evt := <-watcher.Events:
				log.WithField("item", evt.Name).Info("Item change event received")
				fi, error := os.Stat(evt.Name)
				if error != nil {
					log.WithFields(log.Fields{"item": evt.Name, "error": error}).Info("Error on watching item")
					return
				}
				switch mode := fi.Mode(); {
				case mode.IsDir():
					p.readDirectory(evt.Name)
				case mode.IsRegular():
					p.readFile(evt.Name)
				}
			case <-p.stop:
				return
			}
		}
	}()
}

func (provider *StaticDataProvider) getData(resource string) interface{} {
	if resource != "" {
		return convertData(provider.data[resource])
	}
	return convertData(provider.data)
}

func convertData(o interface{}) interface{} {
	if a, ok := o.([]interface{}); ok {
		var result []interface{}
		result = make([]interface{}, len(a))
		for i, e := range a {
			result[i] = convertData(e)
		}
		return result
	} else {
		return convertObject(o)
	}
}

func convertObject(o interface{}) interface{} {
	if m, ok := o.(map[interface{}]interface{}); ok {
		result := make(map[string]interface{}, len(m))
		for k, v := range m {
			propertyName := fmt.Sprint(k)
			result[propertyName] = convertData(v)
		}
		return result
	}
	return o
}

func filterData(data interface{}, parameters map[string]string) interface{} {
	if parameters == nil || len(parameters) == 0 {
		return data
	}

	if list, ok := data.([]interface{}); ok {
		result := make([]interface{}, 0)
		for _, d := range list {
			match := true
			o := d.(map[string]interface{})
			for p, v := range parameters {
				if o[p] != v {
					match = false
					break
				}
			}
			if match {
				result = append(result, o)
			}
		}
		return result
	}
	return data
}

func parseFile(filename string) map[interface{}]interface{} {
	data, error := ioutil.ReadFile(filename)
	if error != nil {
		log.WithFields(log.Fields{"Error": error, "Filename": filename}).Error("error reading file")
		return nil
	}

	newData := make(map[interface{}]interface{})

	err := yaml.Unmarshal(data, newData)
	if err != nil {
		log.WithFields(log.Fields{"Error": error, "Filename": filename}).Error("error parsing file")
	}

	return newData
}
