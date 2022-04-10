package common

import (
	"mokapi/safe"
	"net/url"
	"reflect"
)

var (
	configTypes []configType
)

type Provider interface {
	Read(u *url.URL) (*Config, error)
	Start(chan *Config, *safe.Pool) error
}

type Parser interface {
	Parse(config *Config, reader Reader) error
}

type Reader interface {
	Read(u *url.URL, opts ...ConfigOptions) (*Config, error)
}

type configType struct {
	header     string
	configType reflect.Type
}

func Register(header string, c interface{}) {
	val := reflect.ValueOf(c).Elem()
	configTypes = append(configTypes, configType{header, val.Type()})
}
