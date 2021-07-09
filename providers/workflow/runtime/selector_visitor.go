package runtime

import (
	log "github.com/sirupsen/logrus"
	"mokapi/providers/workflow/ast"
	"mokapi/providers/workflow/path/objectpath"
)

type selectorVisitor struct {
	outer        *visitor
	resolvedRoot bool
}

func newSelectorVisitor(outer *visitor) *selectorVisitor {
	return &selectorVisitor{outer: outer}
}

func (v *selectorVisitor) Visit(e ast.Expression) ast.Visitor {
	if e != nil {
		switch t := e.(type) {
		case *ast.Identifier:
			v.outer.stack.Push(t.Name)
		case *ast.Selector:
			if ident, ok := t.X.(*ast.Identifier); ok {
				v.outer.Visit(ident)
				v.Visit(t.Selector)
				return v.Visit(nil)
			}
			return v
		}
		return nil
	}

	selector := v.outer.stack.Pop().(string)
	source := v.outer.stack.Pop()

	m, err := objectpath.Resolve(selector, source)
	if err != nil {
		log.Debugf("unable to resolve objectpath: %v", err)
	}
	v.outer.stack.Push(m)

	return nil
}
