package js

import (
	"github.com/dop251/goja"
	"gopkg.in/yaml.v3"
	"mokapi/engine/common"
)

type yamlModule struct {
	host common.Host
	rt   *goja.Runtime
}

func newYaml(host common.Host, rt *goja.Runtime) interface{} {
	return &yamlModule{host: host, rt: rt}
}

func (m *yamlModule) Parse(s string) (interface{}, error) {
	var i interface{}
	err := yaml.Unmarshal([]byte(s), &i)
	return i, err
}

func (m *yamlModule) Stringify(i interface{}) (string, error) {
	b, err := yaml.Marshal(i)
	return string(b), err
}
