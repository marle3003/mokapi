package data

import (
	"fmt"
	"io/ioutil"
	"mokapi/models"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type StaticDataProvider struct {
	Path  string
	data  map[string]interface{}
	stop  chan bool
	watch bool
}

func NewStaticDataProvider(path string, watch bool) *StaticDataProvider {
	provider := &StaticDataProvider{Path: path, stop: make(chan bool), watch: watch}
	provider.init()
	return provider
}

func (provider *StaticDataProvider) Provide(name string, schema *models.Schema) (interface{}, error) {
	if data, ok := provider.data[name]; ok {
		return convertData(data), nil
	}
	return nil, nil
}

func (p *StaticDataProvider) init() {
	go func() {
		p.data = make(map[string]interface{})
		p.loadData()

		if p.watch {
			p.addWatcher()
		}
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

// func filterData(data interface{}, resource *models.Resource, context *Context) (interface{}, error) {
// 	if resource != nil && resource.Filter != nil {
// 		filter := resource.Filter
// 		if list, ok := data.([]interface{}); ok {
// 			result := make([]interface{}, 0)
// 			for _, d := range list {
// 				if ok, error := match(d, context, filter); ok && error == nil {
// 					result = append(result, d)
// 				}
// 			}
// 			if len(result) == 0 {
// 				return nil, nil
// 			}
// 			return result, nil
// 		}
// 	}

// 	return data, nil
// }

// func selectValue(data interface{}, context *Context, filter *parser.FilterExp) string {
// 	if filter.Tag == parser.FilterParameter {
// 		return context.Parameters[filter.Value]
// 	} else if filter.Tag == parser.FilterProperty {
// 		if data != nil {
// 			o := data.(map[string]interface{})
// 			if v, ok := o[filter.Value].(string); ok {
// 				return v
// 			}
// 		}
// 	} else if filter.Tag == parser.FilterBody {
// 		s, error := context.Body.Select(filter.Value)
// 		if error != nil {
// 			log.Error(error.Error())
// 		}
// 		return s

// 	} else if filter.Tag == parser.FilterConstant {
// 		return filter.Value
// 	}

// 	// todo: what happens if value is a object instead of string
// 	return ""
// }

func (p *StaticDataProvider) readFile(file string) {
	data, error := ioutil.ReadFile(file)
	if error != nil {
		log.WithFields(log.Fields{"Error": error, "Filename": file}).Error("error reading file")
		return
	}

	switch filepath.Ext(file) {
	case ".yml":
		newData := make(map[interface{}]interface{})

		err := yaml.Unmarshal(data, newData)
		if err != nil {
			log.WithFields(log.Fields{"Error": error, "Filename": file}).Error("error parsing file")
		}

		for k, v := range newData {
			if s, ok := k.(string); ok {
				p.data[s] = v
			} else {
				log.Errorf("Can not add key %v", k)
			}
		}
	default:
		key := filepath.Base(file)
		p.data[key] = string(data)
	}
}
