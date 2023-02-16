package faker

import (
	"encoding/json"
	"github.com/dop251/goja"
	"mokapi/config/dynamic/openapi/schema"
	"mokapi/engine/common"
	"reflect"
)

type Module struct {
	generator *schema.Generator
}

func New(_ common.Host, _ *goja.Runtime) interface{} {
	return &Module{generator: schema.NewGenerator()}
}

func (m *Module) Fake(v goja.Value) (interface{}, error) {
	t := v.ExportType()
	if t.Kind() == reflect.Struct {

	}
	e := v.Export()
	s := toSchema(e)
	i := m.generator.New(&schema.Ref{Value: s})
	return i, nil
}

func toSchema(m interface{}) *schema.Schema {
	s := &schema.Schema{}
	b, _ := json.Marshal(m)
	_ = json.Unmarshal(b, &s)
	return s
}
