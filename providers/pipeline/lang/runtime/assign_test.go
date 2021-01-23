package runtime

import (
	"mokapi/providers/pipeline/lang/ast"
	"mokapi/providers/pipeline/lang/token"
	"mokapi/providers/pipeline/lang/types"
	"testing"
)

func TestAssignVisitor(t *testing.T) {
	s := newStack()
	scope := ast.NewScope(map[string]types.Object{})

	type test func(*testing.T, string)
	type setup func()
	data := []struct {
		test   string
		assign *ast.Assignment
		check  test
		init   setup
	}{
		{"x := 12",
			&ast.Assignment{
				Lhs: &ast.Ident{Name: "x"},
				Tok: token.DEFINE,
				Rhs: &ast.Literal{Value: "12"},
			}, func(t *testing.T, test string) {
				if v, isDefined := scope.Symbol("x"); !isDefined {
					t.Errorf("test: %q: expected symbol 'x' is defined in scope", test)
				} else if v == nil {
					t.Errorf("test: %q: got value nil, expected '12'", test)
				} else if v.String() != "12" {
					t.Errorf("test: %q: got value %q, expected '12'", test, v.String())
				}
			},
			func() {
				s.Push(nil)
				s.Push(types.NewNumber(12))
			},
		},
		{"x = 12",
			&ast.Assignment{
				Lhs: &ast.Ident{Name: "x"},
				Tok: token.ASSIGN,
				Rhs: &ast.Literal{Value: "12"},
			}, func(t *testing.T, test string) {
				if v, isDefined := scope.Symbol("x"); !isDefined {
					t.Errorf("test: %q: expected symbol 'x' is defined in scope", test)
				} else if v == nil {
					t.Errorf("test: %q: got value nil, expected '12'", test)
				} else if v.String() != "12" {
					t.Errorf("test: %q: got value %q, expected '12'", test, v.String())
				}
			},
			func() {
				n := types.NewNumber(0)
				scope.SetSymbol("x", n)
				s.Push(n)
				s.Push(types.NewNumber(12))
			},
		},
		{"x.b = 12",
			&ast.Assignment{
				Lhs: &ast.PathExpr{X: &ast.Ident{Name: "x"}, Path: &ast.Ident{Name: "b"}},
				Tok: token.ASSIGN,
				Rhs: &ast.Literal{Value: "12"},
			}, func(t *testing.T, test string) {
				if v, isDefined := scope.Symbol("x"); !isDefined {
					t.Errorf("test: %q: expected symbol 'x' is defined in scope", test)
				} else if v == nil {
					t.Errorf("test: %q: got value nil, expected '12'", test)
				} else if f, err := v.GetField("b"); err != nil {
					t.Errorf("test: %q:%v", test, err.Error())
				} else if f.String() != "12" {
					t.Errorf("test: %q: got value %q, expected '12'", test, v.String())
				}
			},
			func() {
				e := types.NewExpando()
				scope.SetSymbol("x", e)
				s.Push(e)
				s.Push(types.NewString("b"))
				s.Push(types.NewNumber(12))
			},
		},
	}

	hasErrorsMockFunc = func() bool { return false }
	stackMockFunc = func() *stack { return s }
	scopeMockFunc = func() *ast.Scope { return scope }

	for _, d := range data {
		s.Reset()
		scope = ast.NewScope(map[string]types.Object{})
		d.init()
		v := newAssignVisitor(d.assign, mockVisitor{})
		v.Visit(nil)
		d.check(t, d.test)
	}
}
