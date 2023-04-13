package js

import (
	"fmt"
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

func (m *fakerModule) Fake(v goja.Value) (interface{}, error) {
	s := &schema.Schema{}
	err := m.rt.ExportTo(v, &s)
	if err != nil {
		return nil, fmt.Errorf("expected parameter type of OpenAPI schema")
	}
	return m.generator.New(&schema.Ref{Value: s})
}
