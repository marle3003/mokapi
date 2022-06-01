package mustache

import (
	"github.com/dop251/goja"
	"mokapi/engine/common"
	"mokapi/lib/mustache"
)

type Module struct {
	host common.Host
	rt   *goja.Runtime
}

func New(host common.Host, rt *goja.Runtime) interface{} {
	return &Module{host: host, rt: rt}
}

func (m *Module) Render(template string, data interface{}) (string, error) {
	return mustache.Render(template, data)
}
