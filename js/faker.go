package js

import (
	"github.com/dop251/goja"
	"mokapi/config/dynamic/openapi/schema"
	"mokapi/engine/common"
)

type fakerModule struct {
	generator *schema.Generator
	rt        *goja.Runtime
}

func newFaker(_ common.Host, rt *goja.Runtime) interface{} {
	return &fakerModule{generator: schema.NewGenerator(), rt: rt}
}

func (m *fakerModule) Fake(v goja.Value) interface{} {
	s := &schema.Schema{}
	err := m.rt.ExportTo(v, &s)
	if err != nil {
		panic(m.rt.ToValue("expected parameter type of OpenAPI schema"))
	}
	i, err := m.generator.New(&schema.Ref{Value: s})
	if err != nil {
		panic(m.rt.ToValue(err.Error()))
	}
	return i
}
