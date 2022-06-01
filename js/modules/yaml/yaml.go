package yaml

import (
	"github.com/dop251/goja"
	"gopkg.in/yaml.v3"
	"mokapi/engine/common"
)

type Module struct {
	host common.Host
	rt   *goja.Runtime
}

func New(host common.Host, rt *goja.Runtime) interface{} {
	return &Module{host: host, rt: rt}
}

func (m *Module) Parse(s string) (interface{}, error) {
	result := make(map[string]interface{})
	err := yaml.Unmarshal([]byte(s), result)
	return result, err
}

func (m *Module) Stringify(i interface{}) (string, error) {
	b, err := yaml.Marshal(i)
	return string(b), err
}
