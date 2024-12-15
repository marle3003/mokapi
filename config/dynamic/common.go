package dynamic

import (
	"mokapi/safe"
	"mokapi/version"
	"net/url"
	"reflect"
)

var (
	configTypes []*configType
)

type Provider interface {
	Read(u *url.URL) (*Config, error)
	Start(chan *Config, *safe.Pool) error
}

type configType struct {
	header       string
	configType   reflect.Type
	checkVersion func(version version.Version) bool
}

func AnyVersion(v version.Version) bool {
	return true
}

func Register(header string, checkVersion func(v version.Version) bool, c interface{}) {
	val := reflect.ValueOf(c).Elem()
	configTypes = append(configTypes, &configType{
		header:       header,
		checkVersion: checkVersion,
		configType:   val.Type()})
}
