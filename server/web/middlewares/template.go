package middlewares

import (
	"fmt"
	"io/ioutil"
	"mokapi/models"
	"mokapi/server/web"
	"reflect"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

type template struct {
	config *models.Template
	next   Middleware
}

func NewTemplate(config *models.Template, next Middleware) Middleware {
	m := &template{config: config, next: next}
	return m
}

func (m *template) ServeData(request *Request, context *web.HttpContext) {
	content, error := ioutil.ReadFile(m.config.Filename)
	if error != nil {
		log.WithFields(log.Fields{"Error": error, "Filename": m.config.Filename}).Error("error reading file")
		return
	}

	template := string(content)

	pat := regexp.MustCompile(`\{\{(.*)\}\}`)
	matches := pat.FindAllStringSubmatch(template, -1) // matches is [][]string

	for _, match := range matches {
		path := strings.TrimSpace(match[1])
		value, error := getValue(path, request.Data)
		if error != nil {
			log.Errorf("Error in template middleware: %v", error.Error())
			value = error.Error()
		}

		template = strings.ReplaceAll(template, match[0], value)
	}

	request.Data = template

	m.next.ServeData(request, context)
}

func getValue(path string, element interface{}) (string, error) {
	if element == nil {
		return "NULL", nil
	}
	properties := strings.Split(path, ".")
	currentElement := reflect.ValueOf(element)

	for _, property := range properties {
		k := currentElement.Kind()
		if k == reflect.Map {
			i := currentElement.Interface()
			m := i.(map[string]interface{})
			value := m[property]
			t := reflect.TypeOf(value)
			if t == nil {
				return "null", nil
			}
			currentElement = reflect.ValueOf(value)
		} else if k == reflect.Struct {
			currentElement = currentElement.FieldByNameFunc(func(f string) bool { return strings.ToLower(f) == property })
		} else {
			currentElement = currentElement.Elem().FieldByNameFunc(func(f string) bool { return strings.ToLower(f) == property })
		}

		if !currentElement.IsValid() {
			return "", fmt.Errorf("No value found in path %v", path)
		}
	}

	return fmt.Sprintf("%v", currentElement), nil
}
