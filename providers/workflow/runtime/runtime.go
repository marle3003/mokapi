package runtime

import (
	"mokapi/providers/workflow/ast"
	"mokapi/providers/workflow/parser"
)

func RunExpression(s string, ctx *WorkflowContext) (interface{}, error) {
	exp := parser.Parse(s)

	v := newVisitor(ctx)
	ast.Walk(v, exp)

	return v.stack.Pop(), nil
}
