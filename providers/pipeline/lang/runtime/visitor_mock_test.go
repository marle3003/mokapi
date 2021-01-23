package runtime

import (
	"mokapi/providers/pipeline/lang/ast"
	"mokapi/providers/pipeline/lang/token"
)

var stackMockFunc func() *stack
var scopeMockFunc func() *ast.Scope
var closeScopeMockFunc func()
var visitMockFunc func(ast.Node) ast.Visitor
var addErrorMockFunc func(token.Position, string)
var addErrorfMockFunc func(token.Position, string, ...interface{})
var hasErrorsMockFunc func() bool
var errMockFunc func() error

type mockVisitor struct {
	scope       *ast.Scope
	name        string
	exprVisitor *exprVisitor
	stack       *stack
	found       bool
}

func (v mockVisitor) Stack() *stack {
	return stackMockFunc()
}

func (v mockVisitor) Scope() *ast.Scope {
	return scopeMockFunc()
}

func (v mockVisitor) CloseScope() {
	closeScopeMockFunc()
}

func (v mockVisitor) Visit(node ast.Node) ast.Visitor {
	return visitMockFunc(node)
}

func (v mockVisitor) AddError(pos token.Position, msg string) {
	addErrorMockFunc(pos, msg)
}

func (v mockVisitor) AddErrorf(pos token.Position, format string, args ...interface{}) {
	addErrorfMockFunc(pos, format, args...)
}

func (v mockVisitor) HasErrors() bool {
	return hasErrorsMockFunc()
}

func (v mockVisitor) Err() error {
	return errMockFunc()
}
