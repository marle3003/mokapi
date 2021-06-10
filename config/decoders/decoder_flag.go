package decoders

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type FlagDecoder struct {
}

func (f *FlagDecoder) Decode(flags map[string]string, element interface{}) error {
	for name, value := range flags {
		err := setValue(name, value, element)
		if err != nil {
			return err
		}
	}

	return nil
}

func setValue(name string, value string, element interface{}) error {
	path := strings.Split(name, ".")
	currentElement := reflect.ValueOf(element)

	for _, fieldName := range path {
		k := currentElement.Kind()
		if k != reflect.Struct {
			currentElement = currentElement.Elem().FieldByNameFunc(func(f string) bool { return strings.ToLower(f) == fieldName })
		} else {
			currentElement = currentElement.FieldByNameFunc(func(f string) bool { return strings.ToLower(f) == fieldName })
		}
		if !currentElement.IsValid() {
			return fmt.Errorf("no configuration entry found for %v with value %v", name, value)
		}
	}

	switch currentElement.Kind() {
	case reflect.String:
		currentElement.SetString(value)
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("value %v cannot be parsed as bool: %v", value, err.Error())
		}
		currentElement.SetBool(b)
	}

	return nil
}
