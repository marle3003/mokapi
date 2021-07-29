package runtime

import (
	log "github.com/sirupsen/logrus"
	"mokapi/providers/workflow/ast"
	"mokapi/providers/workflow/path/objectpath"
)

type indexVisitor struct {
	outer *visitor
}

func newIndexVisitor(outer *visitor) *indexVisitor {
	return &indexVisitor{outer: outer}
}

func (v *indexVisitor) Visit(e ast.Expression) ast.Visitor {
	if e != nil {
		return v.outer.Visit(e)
	}

	index := v.outer.stack.Pop()
	source := v.outer.stack.Pop()

	m, err := objectpath.Resolve(index, source)
	if err != nil {
		log.Debugf("unable to resolve index path: %v", err)
	}
	v.outer.stack.Push(m)

	return nil
}
