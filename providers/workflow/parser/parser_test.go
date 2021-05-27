package parser

import (
	"mokapi/providers/workflow/ast"
	"mokapi/providers/workflow/token"
	"testing"
)

func TestNumber(t *testing.T) {
	src := "12"
	x := Parse(src)
	if n, ok := x.(*ast.Literal); !ok {
		t.Errorf("Parse(%q): got %T, expected *ast.Literal", src, x)
	} else if n.Kind != token.NUMBER {
		t.Errorf("Parse(%q): got %T, expected number", src, n.Kind)
	} else if n.Value != "12" {
		t.Errorf("Parse(%q): got %v, expected value 12", src, n.Value)
	}
}

func TestString(t *testing.T) {
	src := "\"Hello World\""
	x := Parse(src)
	if s, ok := x.(*ast.Literal); !ok {
		t.Errorf("Parse(%q): got %T, expected *ast.Literal", src, x)
	} else if s.Kind != token.STRING {
		t.Errorf("Parse(%q): got %T, expected string", src, s.Kind)
	} else if s.Value != "Hello World" {
		t.Errorf("Parse(%q): got %v, expected value 'Hello World'", src, s.Value)
	}
}

func TestIdentExpr(t *testing.T) {
	src := "a"
	x := Parse(src)
	if i, ok := x.(*ast.Identifier); !ok {
		t.Errorf("Parse(%q): got %T, expected *ast.Identifier", src, x)
	} else if i.Name != "a" {
		t.Errorf("Parse(%q): got %v, expected ident 'a'", src, i.Name)
	}
}

func TestSelectorExpr(t *testing.T) {
	src := "a.b"
	x := Parse(src)
	if s, ok := x.(*ast.Selector); !ok {
		t.Errorf("Parse(%q): got %T, expected *ast.Selector", src, x)
	} else {
		if ident, isIdent := s.X.(*ast.Identifier); !isIdent {
			t.Errorf("Parse(%q): got %T, expected *ast.Identifier", src, s.X)
		} else if ident.Name != "a" {
			t.Errorf("Parse(%q): got %v, expected ident 'a'", src, ident.Name)
		}
		if s.Selector.Name != "b" {
			t.Errorf("Parse(%q): got %v, expected selector 'b'", src, s.Selector.Name)
		}
	}
}

func TestBinaryExpr(t *testing.T) {
	src := "a == b"
	x := Parse(src)
	if b, ok := x.(*ast.Binary); !ok {
		t.Errorf("Parse(%q): got %T, expected *ast.Binary", src, x)
	} else {
		if b.Op != token.EQL {
			t.Errorf("Parse(%q): got %v, expected %v", src, b.Op, token.EQL)
		}

		if ident, isIdent := b.Lhs.(*ast.Identifier); !isIdent {
			t.Errorf("Parse(%q): got %T, expected *ast.Identifier", src, b.Lhs)
		} else if ident.Name != "a" {
			t.Errorf("Parse(%q): got %v, expected ident 'a'", src, ident.Name)
		}

		if ident, isIdent := b.Rhs.(*ast.Identifier); !isIdent {
			t.Errorf("Parse(%q): got %T, expected *ast.Identifier", src, b.Rhs)
		} else if ident.Name != "b" {
			t.Errorf("Parse(%q): got %v, expected ident 'a'", src, ident.Name)
		}
	}
}

func TestCallExpr(t *testing.T) {
	src := "a(b, c)"
	x := Parse(src)
	if f, ok := x.(*ast.CallExpr); !ok {
		t.Errorf("Parse(%q): got %T, expected *ast.CallExpr", src, x)
	} else {
		if f.Fun.Name != "a" {
			t.Errorf("Parse(%q): got %v, expected function 'a'", src, f.Fun.Name)
		}

		if len(f.Args) != 2 {
			t.Errorf("Parse(%q): got len of args %v, expected 2", src, len(f.Args))
		}

		if ident, isIdent := f.Args[0].(*ast.Identifier); !isIdent {
			t.Errorf("Parse(%q): got %T, expected *ast.Identifier", src, f.Args[0])
		} else if ident.Name != "b" {
			t.Errorf("Parse(%q): got %v, expected ident 'b'", src, ident.Name)
		}

		if ident, isIdent := f.Args[1].(*ast.Identifier); !isIdent {
			t.Errorf("Parse(%q): got %T, expected *ast.Identifier", src, f.Args[1])
		} else if ident.Name != "c" {
			t.Errorf("Parse(%q): got %v, expected ident 'c'", src, ident.Name)
		}
	}
}

func TestClosureExpr(t *testing.T) {
	src := "x => x == b"
	x := Parse(src)
	if f, ok := x.(*ast.Closure); !ok {
		t.Errorf("Parse(%q): got %T, expected *ast.CallExpr", src, x)
	} else {
		if _, isBinary := f.Func.(*ast.Binary); !isBinary {
			t.Errorf("Parse(%q): got %v, expected block *ast.Binary", src, f.Func)
		}

		if len(f.Args) != 1 {
			t.Errorf("Parse(%q): got len of args %v, expected 1", src, len(f.Args))
		}

		if ident, isIdent := f.Args[0].(*ast.Identifier); !isIdent {
			t.Errorf("Parse(%q): got %T, expected *ast.Identifier", src, f.Args[0])
		} else if ident.Name != "x" {
			t.Errorf("Parse(%q): got %v, expected ident 'x'", src, ident.Name)
		}
	}
}

func TestList(t *testing.T) {
	e := []string{"1", "2", "3", "4"}
	src := "[1, 2, 3, 4]"

	x := Parse(src)
	seq, isSeq := x.(*ast.SequenceExpr)
	if !isSeq {
		t.Errorf("Parse(%q): got %v, expected block *ast.SequenceExpr", src, x)
		return
	}
	if seq.IsMap {
		t.Errorf("ParseExpr(%q):expected not a map", src)
	}
	for i, v := range seq.Values {
		if l, isLit := v.(*ast.Literal); !isLit {
			t.Errorf("ParseExpr(%q): got %T, expected *ast.Literal", src, v)
		} else {
			if l.Value != e[i] {
				t.Errorf("ParseExpr(%q): got %v, expected %v", src, l.Value, e[i])
			}
		}
	}
}

func TestListRange(t *testing.T) {
	src := "[1...4]"

	x := Parse(src)
	seq, isSeq := x.(*ast.SequenceExpr)
	if !isSeq {
		t.Errorf("Parse(%q): got %v, expected block *ast.SequenceExpr", src, x)
		return
	}
	if seq.IsMap {
		t.Errorf("ParseExpr(%q):expected not a map", src)
	}
	for _, v := range seq.Values {
		if r, isRange := v.(*ast.RangeExpr); !isRange {
			t.Errorf("ParseExpr(%q): got %T, expected *ast.RangeExpr", src, v)
		} else {
			if l, isLit := r.Start.(*ast.Literal); !isLit {
				t.Errorf("ParseExpr(%q): got %T, expected *ast.Literal", src, r.Start)
			} else {
				if l.Value != "1" {
					t.Errorf("ParseExpr(%q): got %v, expected 1", src, l.Value)
				}
			}
			if l, isLit := r.End.(*ast.Literal); !isLit {
				t.Errorf("ParseExpr(%q): got %T, expected *ast.Literal", src, r.Start)
			} else {
				if l.Value != "4" {
					t.Errorf("ParseExpr(%q): got %v, expected 4", src, l.Value)
				}
			}
		}
	}
}

func TestMap(t *testing.T) {
	e := map[string]string{"a": "1", "b": "2", "c": "3", "z": "4"}
	src := "[a: 1, b: 2, c: 3, z: 4]"

	x := Parse(src)
	seq, isSeq := x.(*ast.SequenceExpr)
	if !isSeq {
		t.Errorf("Parse(%q): got %v, expected block *ast.SequenceExpr", src, x)
		return
	}
	if !seq.IsMap {
		t.Errorf("ParseExpr(%q):expected a map", src)
	}
	for _, v := range seq.Values {
		if kv, isLit := v.(*ast.KeyValueExpr); !isLit {
			t.Errorf("ParseExpr(%q): got %T, expected *ast.Literal", src, v)
		} else {
			if i, isIdent := kv.Key.(*ast.Identifier); !isIdent {
				t.Errorf("ParseExpr(%q): got %T, expected key as *ast.Ident", src, kv.Key)
			} else if val, ok := e[i.Name]; !ok {
				t.Errorf("ParseExpr(%q): key %v not found", src, i.Name)
			} else if ok {
				if l, isLit := kv.Value.(*ast.Literal); !isLit {
					t.Errorf("ParseExpr(%q): got %v, expected value as *ast.Literal", src, kv.Value)
				} else if l.Value != val {
					t.Errorf("ParseExpr(%q): got %v, expected value is %v", src, l.Value, val)
				}
			}
		}
	}
}
