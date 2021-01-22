package runtime

import (
	"github.com/pkg/errors"
	"mokapi/providers/pipeline/lang/ast"
	"mokapi/providers/pipeline/lang/parser"
	"mokapi/providers/pipeline/lang/token"
	"mokapi/providers/pipeline/lang/types"
)

type visitor interface {
	ast.Visitor
	AddError(pos token.Position, msg string)
	AddErrorf(pos token.Position, format string, args ...interface{})
	HasErrors() bool
	Err() error
	CloseScope()
	Stack() *stack
	Scope() *ast.Scope
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
	v := newExprVisitor(newStack(), scope)
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
		err = v.Err()
	}()

	ast.Walk(v, n)

	return
}

type pipelineVisitor struct {
	visitorErrorHandler
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
	if v.HasErrors() {
		return nil
	}

	switch n := node.(type) {
	case *ast.Pipeline:
		v.scope = n.Scope
		v.exprVisitor.scope = v.scope
		if n.Name != v.name {
			return nil
		} else {
			v.found = true
		}
	case *ast.Stage:
		v.scope = n.Scope
		v.exprVisitor.scope = n.Scope
		return newStageVisitor(n, v)
	case *ast.ExprStatement:
		// case 'when'
		return v.exprVisitor
	case *ast.VarsBlock:
		return v.exprVisitor
	case *ast.StepBlock:
		return v.exprVisitor
	}

	return v
}

func (v *pipelineVisitor) Stack() *stack {
	return v.stack
}

func (v *pipelineVisitor) Scope() *ast.Scope {
	return v.scope
}

func (v *pipelineVisitor) HasError() bool {
	return len(v.errors) != 0 || v.exprVisitor.HasErrors()
}

func (v *pipelineVisitor) Err() error {
	if v.HasErrors() {
		return v.Err()
	} else {
		return v.exprVisitor.Err()
	}
}

func (v *pipelineVisitor) CloseScope() {
	v.scope = v.scope.Outer
}

type visitorErrorHandler struct {
	errors parser.ErrorList
}

func (v *visitorErrorHandler) AddError(pos token.Position, msg string) {
	v.errors.Add(pos, msg)
}

func (v *visitorErrorHandler) AddErrorf(pos token.Position, format string, args ...interface{}) {
	v.errors.Addf(pos, format, args...)
}

func (v *visitorErrorHandler) HasErrors() bool {
	return len(v.errors) > 0
}

func (v *visitorErrorHandler) Err() error {
	return v.errors.Err()
}
