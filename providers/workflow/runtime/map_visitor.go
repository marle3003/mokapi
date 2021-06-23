package runtime

import (
	log "github.com/sirupsen/logrus"
	"mokapi/providers/workflow/ast"
)

type mapVisitor struct {
	outer *visitor
	m     *ast.MapExpr
}

func newMapVisitor(m *ast.MapExpr, outer *visitor) *mapVisitor {
	return &mapVisitor{m: m, outer: outer}
}

func (v *mapVisitor) Visit(e ast.Expression) ast.Visitor {
	if e != nil {
		return v.outer.Visit(e)
	}

	v.outer.stack.Push(v.getExpando())

	return nil
}

func (v *mapVisitor) getExpando() map[string]interface{} {
	expando := make(map[string]interface{})
	for range v.m.Values {
		val := v.outer.stack.Pop()
		key := v.outer.stack.Pop()
		if key == nil {
			log.Errorf("empty key for value %q is not allowed", val)
			continue
		}
		expando[key.(string)] = val
	}
	return expando
}
