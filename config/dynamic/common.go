package dynamic

import (
	"mokapi/safe"
	"mokapi/version"
	"net/url"
	"reflect"
	"sync"
)

var (
	configTypes   []*configType
	configTypesMu sync.RWMutex
)

type Provider interface {
	Read(u *url.URL) (*Config, error)
	Start(chan ConfigEvent, *safe.Pool) error
}

type configType struct {
	header       string
	configType   reflect.Type
	checkVersion func(version version.Version) bool
}

func AnyVersion(_ version.Version) bool {
	return true
}

func Register(header string, checkVersion func(v version.Version) bool, c interface{}) {
	val := reflect.ValueOf(c).Elem()
	configTypesMu.Lock()
	defer configTypesMu.Unlock()
	configTypes = append(configTypes, &configType{
		header:       header,
		checkVersion: checkVersion,
		configType:   val.Type()})
}

func Reset() {
	configTypesMu.Lock()
	defer configTypesMu.Unlock()
	configTypes = nil
}
