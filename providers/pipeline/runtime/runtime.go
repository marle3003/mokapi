package runtime

import (
	"github.com/pkg/errors"
	"mokapi/providers/pipeline/lang"
	"mokapi/providers/pipeline/lang/types"
	"strings"
)

type visitor interface {
	lang.Visitor
	addError(err error)
	hasErrors() bool
	err() error
}

func RunPipeline(f *lang.File, name string, scope *Scope) error {
	v := newPipelineVisitor(name, scope)
	return run(f, v)
}

func runExpr(expr lang.Expression, scope *Scope) (result types.Object, err error) {
	v := &exprVisitor{scope: scope, stack: newStack()}
	err = run(expr, v)
	result = v.stack.Pop()
	return
}

func run(n lang.Node, v visitor) (err error) {
	defer func() {
		err = v.err()
	}()

	lang.Walk(v, n)

	return
}

type pipelineVisitor struct {
	visitorImpl
	scope       *Scope
	name        string
	exprVisitor *exprVisitor
}

func newPipelineVisitor(name string, scope *Scope) *pipelineVisitor {
	stack := newStack()
	return &pipelineVisitor{
		scope:       scope,
		name:        name,
		exprVisitor: &exprVisitor{scope: scope, stack: stack},
	}
}

func (v *pipelineVisitor) Visit(node lang.Node) lang.Visitor {
	if v.hasErrors() {
		return nil
	}

	switch n := node.(type) {
	case *lang.Pipeline:
		if n.Name != v.name {
			return nil
		}
	case *lang.StepBlock:
		return v.exprVisitor
	}

	return v
}

func (v *pipelineVisitor) hasError() bool {
	return len(v.errors) != 0 || v.exprVisitor.hasErrors()
}

func (v *pipelineVisitor) err() error {
	if v.hasErrors() {
		return v.err()
	} else {
		return v.exprVisitor.err()
	}
}

type visitorImpl struct {
	errors []error
}

func (v *visitorImpl) addError(err error) {
	v.errors = append(v.errors, err)
}

func (v *visitorImpl) hasErrors() bool {
	return len(v.errors) > 0
}

func (v *visitorImpl) err() error {
	if !v.hasErrors() {
		return nil
	}
	sb := strings.Builder{}
	for i, e := range v.errors {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(e.Error())
	}
	return errors.New(sb.String())
}
