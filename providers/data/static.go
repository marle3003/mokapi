package data

import (
	"fmt"
	"io/ioutil"
	"mokapi/providers/parser"
	"mokapi/service"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type StaticDataProvider struct {
	Path  string
	data  map[interface{}]interface{}
	stop  chan bool
	watch bool
}

func NewStaticDataProvider(path string, watch bool) *StaticDataProvider {
	provider := &StaticDataProvider{Path: path, stop: make(chan bool), watch: watch}
	provider.init()
	return provider
}

func (provider *StaticDataProvider) Provide(parameters map[string]string, schema *service.Schema) (interface{}, error) {
	data := provider.getData(schema.Resource)

	filtered, error := filterData(data, schema, parameters)
	if error != nil {
		return nil, error
	}

	return filtered, nil
}

func (p *StaticDataProvider) init() {
	go func() {
		p.data = make(map[interface{}]interface{})
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

func (provider *StaticDataProvider) getData(r *service.Resource) interface{} {
	if r != nil && r.Name != "" {
		return convertData(provider.data[r.Name])
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

func filterData(data interface{}, schema *service.Schema, parameters map[string]string) (interface{}, error) {
	if parameters == nil || len(parameters) == 0 {
		return data, nil
	}

	if schema.Resource != nil && schema.Resource.Filter != nil {
		filter := schema.Resource.Filter
		if list, ok := data.([]interface{}); ok {
			result := make([]interface{}, 0)
			for _, d := range list {
				if ok, error := match(d, parameters, filter); ok && error == nil {
					result = append(result, d)
				}
			}
			if len(result) == 0 {
				return nil, nil
			}
			return result, nil
		}
	}

	return data, nil
}

func match(data interface{}, parameters map[string]string, filter *parser.FilterExp) (bool, error) {
	switch filter.Tag {
	case parser.FilterEqualityMatch:
		left := selectValue(data, parameters, filter.Children[0])
		right := selectValue(data, parameters, filter.Children[1])

		return left == right, nil
	case parser.FilterLike:
		left := selectValue(data, parameters, filter.Children[0])
		right := selectValue(data, parameters, filter.Children[1])

		s := strings.ReplaceAll(right, "%", ".*")
		regex := regexp.MustCompile(s)
		match := regex.FindStringSubmatch(left)

		return len(match) > 0, nil
	}

	return false, fmt.Errorf("Unsupported filter tag %v", filter.Tag)
}

func selectValue(data interface{}, parameters map[string]string, filter *parser.FilterExp) string {
	if filter.Tag == parser.FilterParameter {
		return parameters[filter.Value]
	}

	o := data.(map[string]interface{})

	if v, ok := o[filter.Value].(string); ok {
		return v
	}

	// todo: what happens if value is a object instead of string
	return ""
}

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
			p.data[k] = v
		}
	default:
		key := filepath.Base(file)
		p.data[key] = string(data)
	}
}
