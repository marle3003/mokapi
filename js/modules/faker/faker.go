package faker

import (
	"encoding/json"
	"github.com/dop251/goja"
	"mokapi/config/dynamic/openapi/schema"
	"mokapi/engine/common"
)

type Module struct {
	generator *schema.Generator
}

func New(_ common.Host, _ *goja.Runtime) interface{} {
	return &Module{generator: schema.NewGenerator()}
}

func (m *Module) Fake(v goja.Value) (interface{}, error) {
	e := v.Export().(map[string]interface{})
	s := toSchema(e)
	i := m.generator.New(&schema.Ref{Value: s})
	return i, nil
}

func toSchema(m map[string]interface{}) *schema.Schema {
	s := &schema.Schema{}
	b, _ := json.Marshal(m)
	_ = json.Unmarshal(b, &s)
	return s
}
