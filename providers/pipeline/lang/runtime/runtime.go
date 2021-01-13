package runtime

import (
	"github.com/pkg/errors"
	"mokapi/providers/pipeline/lang/ast"
	"mokapi/providers/pipeline/lang/types"
	"strings"
)

type visitor interface {
	ast.Visitor
	addError(err error)
	hasErrors() bool
	err() error
	closeScope()
}

func RunPipeline(f *ast.File, name string) (err error) {
	v := newPipelineVisitor(name, f.Scope)

	defer func() {
		if !v.found {
			err = errors.Errorf("pipeline '%v' not found", name)
		}
	}()

	err = run(f, v)

	return
}

func runExpr(expr ast.Expression, scope *ast.Scope) (result types.Object, err error) {
	v := &exprVisitor{scope: scope, stack: newStack()}
	err = run(expr, v)
	result = v.stack.Pop()
	return
}

func runBlock(block *ast.Block, scope *ast.Scope) (result types.Object, err error) {
	v := &exprVisitor{scope: scope, stack: newStack()}
	err = run(block, v)
	result = v.stack.Pop()
	return
}

func run(n ast.Node, v visitor) (err error) {
	defer func() {
		err = v.err()
	}()

	ast.Walk(v, n)

	return
}

type pipelineVisitor struct {
	visitorImpl
	scope       *ast.Scope
	name        string
	exprVisitor *exprVisitor
	stack       *stack
	found       bool
}

func newPipelineVisitor(name string, scope *ast.Scope) *pipelineVisitor {
	stack := newStack()
	return &pipelineVisitor{
		scope:       scope,
		name:        name,
		stack:       stack,
		exprVisitor: &exprVisitor{scope: scope, stack: stack},
	}
}

func (v *pipelineVisitor) Visit(node ast.Node) ast.Visitor {
	if v.hasErrors() {
		return nil
	}

	switch n := node.(type) {
	case *ast.Pipeline:
		if n.Name != v.name {
			return nil
		} else {
			v.found = true
		}
	case *ast.Stage:
		v.scope = n.Scope
		v.exprVisitor.scope = n.Scope
		return &stageVisitor{stack: v.stack, outer: v, stage: n}
	case *ast.ExprStatement:
		// case 'when'
		return v.exprVisitor
	case *ast.StepBlock:
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

func (v *pipelineVisitor) closeScope() {
	v.scope = v.scope.Outer
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

func (v *visitorImpl) closeScope() {

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
