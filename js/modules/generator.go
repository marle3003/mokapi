package modules

import (
	"encoding/json"
	"github.com/dop251/goja"
	"mokapi/config/dynamic/openapi/schema"
	"mokapi/engine/common"
)

type Generator struct {
	rt        *goja.Runtime
	generator *schema.Generator
}

func NewGenerator(host common.Host, rt *goja.Runtime) interface{} {
	return &Generator{generator: schema.NewGenerator(), rt: rt}
}

func (g *Generator) New(v goja.Value) (interface{}, error) {
	e := v.Export().(map[string]interface{})
	s := toSchema(e)
	i := g.generator.New(&schema.Ref{Value: s})
	return g.rt.ToValue(i), nil
}

func toSchema(m map[string]interface{}) *schema.Schema {
	s := &schema.Schema{}
	b, _ := json.Marshal(m)
	_ = json.Unmarshal(b, &s)
	return s
}
