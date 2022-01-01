package common

import (
	"reflect"
)

var (
	configTypes []configType
)

type configType struct {
	header     string
	configType reflect.Type
}

func Register(header string, c interface{}) {
	val := reflect.ValueOf(c).Elem()
	configTypes = append(configTypes, configType{header, val.Type()})
}
