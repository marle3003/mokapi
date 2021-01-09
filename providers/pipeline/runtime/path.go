package runtime

import (
	"fmt"
	"github.com/pkg/errors"
	"mokapi/providers/pipeline/lang"
	"mokapi/providers/pipeline/lang/types"
)

type pathVisitor struct {
	stack *stack
	outer visitor
	expr  *lang.PathExpr
}

func (v *pathVisitor) Visit(node lang.Node) lang.Visitor {
	if node != nil && !v.outer.hasErrors() {
		return v.outer.Visit(node)
	}

	n := len(v.expr.Args)
	args := make(map[string]types.Object, n)
	for i := n - 1; i >= 0; i-- {
		kv := v.stack.Pop().(*types.KeyValuePair)
		argName := kv.Key
		if len(argName) == 0 {
			argName = fmt.Sprintf("%v", i)
		}
		args[argName] = kv.Value
	}

	path := v.stack.Pop().String()

	obj := v.stack.Pop()
	if obj == nil {
		v.outer.addError(errors.Errorf("null reference"))
		return nil
	}

	p, ok := obj.(*types.Path)
	if !ok {
		p = types.NewPath(obj)
	}

	val, err := p.Resolve(path, args)
	if err != nil {
		v.outer.addError(err)
	} else {
		v.stack.Push(val)
	}

	return nil
}
