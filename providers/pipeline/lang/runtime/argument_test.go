package runtime

import (
	"mokapi/providers/pipeline/lang/ast"
	"mokapi/providers/pipeline/lang/types"
	"testing"
)

func TestArgumentVisitor_EmptyName(t *testing.T) {
	data := []struct {
		arg *ast.Argument
	}{
		{&ast.Argument{
			Name: "",
		}},
		{&ast.Argument{
			Name: "aKey",
		}},
	}

	for _, d := range data {
		s := newStack()
		s.Push(types.NewNumber(12))

		hasErrorsMockFunc = func() bool { return false }
		stackMockFunc = func() *stack { return s }

		v := newArgumentVisitor(d.arg, mockVisitor{})
		v.Visit(nil)

		o := s.Pop()
		if kv, isKeyValue := o.(*types.KeyValuePair); !isKeyValue {
			t.Errorf("got %T, expected *types.KeyValuePair", o)
		} else {
			if kv.Key != d.arg.Name {
				t.Errorf("got key %q, expected %q", kv.Key, d.arg.Name)
			}
			if n, isNumber := kv.Value.(*types.Number); !isNumber {
				t.Errorf("got %T, expected *types.Number", kv.Value)
			} else if n.String() != "12" {
				t.Errorf("got val %q, expected %q", n, 12)
			}
		}
	}
}
