package common

import (
	"net/url"
	"reflect"
)

var (
	configTypes []configType
)

type Config struct {
	Url  *url.URL
	Data interface{}
}

type configType struct {
	header     string
	configType reflect.Type
}

func Register(header string, c interface{}) {
	val := reflect.ValueOf(c).Elem()
	configTypes = append(configTypes, configType{header, val.Type()})
}
