package runtime

import (
	"mokapi/providers/pipeline/lang/ast"
	"mokapi/providers/pipeline/lang/types"
)

type sequenceVisitor struct {
	outer visitor
	seq   *ast.SequenceExpr
}

func newSequenceVisitor(seq *ast.SequenceExpr, outer visitor) *sequenceVisitor {
	return &sequenceVisitor{seq: seq, outer: outer}
}

func (v *sequenceVisitor) Visit(node ast.Node) ast.Visitor {
	if v.outer.HasErrors() {
		return nil
	}
	if node != nil {
		return v.outer.Visit(node)
	}

	if len(v.seq.Values) == 0 {
		v.outer.Stack().Push(v.getArray())
	} else {
		if v.seq.IsMap {
			v.outer.Stack().Push(v.getExpando())
		} else {
			v.outer.Stack().Push(v.getArray())
		}
	}

	return nil
}

func (v *sequenceVisitor) getArray() *types.Array {
	var values []types.Object
	for range v.seq.Values {
		val := v.outer.Stack().Pop()
		values = append(values, val)

	}
	result := types.NewArray()
	for _, i := range reverse(values) {
		if a, isArray := i.(*types.Array); isArray {
			result.AddRange(a)
		} else {
			result.Add(i)
		}
	}
	return result
}

func (v *sequenceVisitor) getExpando() *types.Expando {
	expando := types.NewExpando()
	for range v.seq.Values {
		val := v.outer.Stack().Pop()
		key := v.outer.Stack().Pop()
		expando.SetField(key.String(), val)
	}
	return expando
}

func reverse(numbers []types.Object) []types.Object {
	for i := 0; i < len(numbers)/2; i++ {
		j := len(numbers) - i - 1
		numbers[i], numbers[j] = numbers[j], numbers[i]
	}
	return numbers
}
