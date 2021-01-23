package runtime

import (
	"mokapi/providers/pipeline/lang/ast"
	"mokapi/providers/pipeline/lang/token"
	"mokapi/providers/pipeline/lang/types"
	"testing"
)

func TestBinaryVisitor(t *testing.T) {
	s := newStack()
	scope := ast.NewScope(map[string]types.Object{})

	type test func(*testing.T, string)
	type setup func()
	data := []struct {
		test   string
		binary *ast.Binary
		check  test
		init   setup
	}{
		{"1 + 1",
			&ast.Binary{
				Op:  token.ADD,
				Rhs: &ast.Literal{},
			}, func(t *testing.T, test string) {
				v := s.Pop()
				if n, isNum := v.(*types.Number); !isNum {
					t.Errorf("test %q: got %T, expected *ast.Number", test, v)
				} else if n.Val() != 2 {
					t.Errorf("test %q: got %v, expected 2", test, n)
				}
			},
			func() {
				s.Push(types.NewNumber(1))
				s.Push(types.NewNumber(1))
			},
		},
		{"1 - 1",
			&ast.Binary{
				Op:  token.SUB,
				Rhs: &ast.Literal{},
			}, func(t *testing.T, test string) {
				v := s.Pop()
				if n, isNum := v.(*types.Number); !isNum {
					t.Errorf("test %q: got %T, expected *ast.Number", test, v)
				} else if n.Val() != 0 {
					t.Errorf("test %q: got %v, expected 0", test, n)
				}
			},
			func() {
				s.Push(types.NewNumber(1))
				s.Push(types.NewNumber(1))
			},
		},
		{"1 * 2",
			&ast.Binary{
				Op:  token.MUL,
				Rhs: &ast.Literal{},
			}, func(t *testing.T, test string) {
				v := s.Pop()
				if n, isNum := v.(*types.Number); !isNum {
					t.Errorf("test %q: got %T, expected *ast.Number", test, v)
				} else if n.Val() != 2 {
					t.Errorf("test %q: got %v, expected 2", test, n)
				}
			},
			func() {
				s.Push(types.NewNumber(1))
				s.Push(types.NewNumber(2))
			},
		},
		{"4 / 2",
			&ast.Binary{
				Op:  token.QUO,
				Rhs: &ast.Literal{},
			}, func(t *testing.T, test string) {
				v := s.Pop()
				if n, isNum := v.(*types.Number); !isNum {
					t.Errorf("test %q: got %T, expected *ast.Number", test, v)
				} else if n.Val() != 2 {
					t.Errorf("test %q: got %v, expected 2", test, n)
				}
			},
			func() {
				s.Push(types.NewNumber(4))
				s.Push(types.NewNumber(2))
			},
		},
		{"1 + 1 * 4 + 6 / 3 - 1",
			&ast.Binary{
				Op: token.ADD,
				Rhs: &ast.Binary{
					Op: token.MUL,
					Rhs: &ast.Binary{
						Op: token.ADD,
						Rhs: &ast.Binary{
							Op: token.QUO,
							Rhs: &ast.Binary{
								Op:  token.SUB,
								Rhs: &ast.Literal{},
							},
						},
					},
				},
			}, func(t *testing.T, test string) {
				v := s.Pop()
				if n, isNum := v.(*types.Number); !isNum {
					t.Errorf("test %q: got %T, expected *ast.Number", test, v)
				} else if n.Val() != 6 {
					t.Errorf("test %q: got %v, expected 6", test, n)
				}
			},
			func() {
				s.Push(types.NewNumber(1))
				s.Push(types.NewNumber(1))
				s.Push(types.NewNumber(4))
				s.Push(types.NewNumber(6))
				s.Push(types.NewNumber(3))
				s.Push(types.NewNumber(1))
			},
		},
	}

	hasErrorsMockFunc = func() bool { return false }
	stackMockFunc = func() *stack { return s }
	scopeMockFunc = func() *ast.Scope { return scope }
	visitMockFunc = func(n ast.Node) ast.Visitor { return nil }

	for _, d := range data {
		s.Reset()
		scope = ast.NewScope(map[string]types.Object{})
		d.init()

		mock := mockVisitor{}
		v := newBinaryVisitor(d.binary, mock)

		var walk func(bin *ast.Binary)
		walk = func(bin *ast.Binary) {
			v.Visit(bin.Rhs)
			if b, isBinary := bin.Rhs.(*ast.Binary); isBinary {
				walk(b)
			}
			v.Visit(nil)
		}
		walk(d.binary)

		d.check(t, d.test)
	}
}
