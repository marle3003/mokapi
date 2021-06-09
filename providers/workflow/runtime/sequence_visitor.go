package runtime

import (
	"mokapi/providers/workflow/ast"
)

type sequenceVisitor struct {
	outer *visitor
	seq   *ast.SequenceExpr
}

func newSequenceVisitor(seq *ast.SequenceExpr, outer *visitor) *sequenceVisitor {
	return &sequenceVisitor{seq: seq, outer: outer}
}

func (v *sequenceVisitor) Visit(e ast.Expression) ast.Visitor {
	if e != nil {
		return v.outer.Visit(e)
	}

	v.outer.stack.Push(v.getArray())

	return nil
}

func (v *sequenceVisitor) getArray() interface{} {
	var values []interface{}
	for range v.seq.Values {
		val := v.outer.stack.Pop()
		values = append(values, val)

	}
	result := make([]interface{}, 0)
	for _, i := range reverse(values) {
		if a, isArray := i.([]interface{}); isArray {
			result = append(result, a...)
		} else {
			result = append(result, i)
		}
	}
	return result
}

func reverse(numbers []interface{}) []interface{} {
	for i := 0; i < len(numbers)/2; i++ {
		j := len(numbers) - i - 1
		numbers[i], numbers[j] = numbers[j], numbers[i]
	}
	return numbers
}
