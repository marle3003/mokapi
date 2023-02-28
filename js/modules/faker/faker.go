package faker

import (
	"encoding/json"
	"fmt"
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
	e := v.Export()
	s, err := toSchema(e)
	if err != nil {
		return nil, fmt.Errorf("expected parameter type of schema")
	}
	i := m.generator.New(&schema.Ref{Value: s})
	return i, nil
}

func toSchema(m interface{}) (*schema.Schema, error) {
	s := &schema.Schema{}
	b, _ := json.Marshal(m)
	err := json.Unmarshal(b, &s)
	return s, err
}
