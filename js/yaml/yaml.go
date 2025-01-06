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

func Require(vm *goja.Runtime, module *goja.Object) {
	o := vm.Get("mokapi/internal").(*goja.Object)
	host := o.Get("host").Export().(common.Host)
	f := &Module{
		rt:   vm,
		host: host,
	}
	obj := module.Get("exports").(*goja.Object)
	obj.Set("parse", f.Parse)
	obj.Set("stringify", f.Stringify)
}

func (m *Module) Parse(s string) (interface{}, error) {
	var i interface{}
	err := yaml.Unmarshal([]byte(s), &i)
	return i, err
}

func (m *Module) Stringify(i interface{}) (string, error) {
	b, err := yaml.Marshal(i)
	return string(b), err
}
