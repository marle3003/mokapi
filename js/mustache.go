package js

import (
	"github.com/dop251/goja"
	"mokapi/engine/common"
	"mokapi/lib/mustache"
)

type mustacheModule struct {
}

func newMustache(_ common.Host, _ *goja.Runtime) interface{} {
	return &mustacheModule{}
}

func (m *mustacheModule) Render(template string, data interface{}) (string, error) {
	return mustache.Render(template, data)
}
