package pipeline

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

type MustacheStep struct {
}

type MustacheStepExecution struct {
	Template string      `step:"format,position=0,required"`
	Data     interface{} `step:"data,position=1,required"`
}

func (e *MustacheStep) Start() StepExecution {
	return &MustacheStepExecution{}
}

func (e *MustacheStepExecution) Run(_ StepContext) (interface{}, error) {
	pat := regexp.MustCompile(`{{(.*)}}`)
	matches := pat.FindAllStringSubmatch(e.Template, -1) // matches is [][]string

	s := e.Template
	for _, match := range matches {
		path := strings.TrimSpace(match[1])
		value, err := getValue(path, e.Data)
		if err != nil {
			return nil, err
		}

		s = strings.ReplaceAll(s, match[0], value)
	}

	return s, nil
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
