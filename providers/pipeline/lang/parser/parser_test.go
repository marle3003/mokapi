package parser

import (
	"fmt"
	"mokapi/providers/pipeline/lang/ast"
	"mokapi/providers/pipeline/lang/token"
	"mokapi/providers/pipeline/lang/types"
	"reflect"
	"testing"
)

func TestAddExpr(t *testing.T) {
	src := "a + b"
	x, err := ParseExpr(src, ast.NewScope(map[string]types.Object{"a": types.NewString(""), "b": types.NewString("")}))
	if err != nil {
		t.Errorf("ParseExpr(%q):%v", src, err)
	} else {
		if b, ok := x.(*ast.Binary); !ok {
			t.Errorf("ParseExpr(%q): got %T, expected *ast.BinaryExpr", src, x)
		} else {
			if b.Op != token.ADD {
				t.Errorf("ParseExpr(%q): got %v, expected %v", src, b.Op, token.ADD)
			}
			if i, isIdent := b.Lhs.(*ast.Ident); !isIdent {
				t.Errorf("ParseExpr(%q): got %v, expected ident", src, reflect.TypeOf(b.Lhs))
			} else if i.Name != "a" {
				t.Errorf("ParseExpr(%q): got %v, expected ident 'a'", src, i.Name)
			}
			if b.Op != token.ADD {
				t.Errorf("ParseExpr(%q): got %v, expected %v", src, b.Op, token.ADD)
			}
			if i, isIdent := b.Rhs.(*ast.Ident); !isIdent {
				t.Errorf("ParseExpr(%q): got %v, expected ident", src, reflect.TypeOf(b.Rhs))
			} else if i.Name != "b" {
				t.Errorf("ParseExpr(%q): got %v, expected ident 'b'", src, i.Name)
			}
		}
	}
}

func TestPath(t *testing.T) {
	src := "a.b"
	x, err := ParseExpr(src, ast.NewScope(map[string]types.Object{"a": types.NewString(""), "b": types.NewString("")}))
	if err != nil {
		t.Errorf("ParseExpr(%q):%v", src, err)
	} else {
		if p, ok := x.(*ast.PathExpr); !ok {
			t.Errorf("ParseExpr(%q): got %T, expected *ast.PathExpr", src, x)
		} else {
			if p.Lhs {
				t.Errorf("ParseExpr(%q): expected lhs expr", src)
			}
			if i, isIdent := p.X.(*ast.Ident); !isIdent {
				t.Errorf("ParseExpr(%q): got %v, expected ident", src, reflect.TypeOf(p.X))
			} else if i.Name != "a" {
				t.Errorf("ParseExpr(%q): got %v, expected ident 'a'", src, i.Name)
			}
			if i, isIdent := p.Path.(*ast.Ident); !isIdent {
				t.Errorf("ParseExpr(%q): got %v, expected ident", src, reflect.TypeOf(p.Path))
			} else if i.Name != "b" {
				t.Errorf("ParseExpr(%q): got %v, expected ident 'a'", src, i.Name)
			}
			if len(p.Args) != 0 {
				t.Errorf("ParseExpr(%q): got len of args %v, expected 0", src, len(p.Args))
			}
		}
	}
}

func TestPathChildren(t *testing.T) {
	src := "a.'*'"
	x, err := ParseExpr(src, ast.NewScope(map[string]types.Object{"a": types.NewString(""), "b": types.NewString("")}))
	if err != nil {
		t.Errorf("ParseExpr(%q):%v", src, err)
	} else {
		if p, ok := x.(*ast.PathExpr); !ok {
			t.Errorf("ParseExpr(%q): got %T, expected *ast.PathExpr", src, x)
		} else {
			if p.Lhs {
				t.Errorf("ParseExpr(%q): expected lhs expr", src)
			}
			if i, isIdent := p.X.(*ast.Ident); !isIdent {
				t.Errorf("ParseExpr(%q): got %v, expected ident", src, reflect.TypeOf(p.X))
			} else if i.Name != "a" {
				t.Errorf("ParseExpr(%q): got %v, expected ident 'a'", src, i.Name)
			}
			if i, isLiteral := p.Path.(*ast.Literal); !isLiteral {
				t.Errorf("ParseExpr(%q): got %v, expected ident", src, reflect.TypeOf(p.Path))
			} else if i.Value != "*" {
				t.Errorf("ParseExpr(%q): got %v, expected ident '*'", src, i.Value)
			}
			if len(p.Args) != 0 {
				t.Errorf("ParseExpr(%q): got len of args %v, expected 0", src, len(p.Args))
			}
		}
	}
}

func TestPathDeep(t *testing.T) {
	src := "a.'**'"
	x, err := ParseExpr(src, ast.NewScope(map[string]types.Object{"a": types.NewString(""), "b": types.NewString("")}))
	if err != nil {
		t.Errorf("ParseExpr(%q):%v", src, err)
	} else {
		if p, ok := x.(*ast.PathExpr); !ok {
			t.Errorf("ParseExpr(%q): got %T, expected *ast.PathExpr", src, x)
		} else {
			if p.Lhs {
				t.Errorf("ParseExpr(%q): expected lhs expr", src)
			}
			if i, isIdent := p.X.(*ast.Ident); !isIdent {
				t.Errorf("ParseExpr(%q): got %v, expected ident", src, reflect.TypeOf(p.X))
			} else if i.Name != "a" {
				t.Errorf("ParseExpr(%q): got %v, expected ident 'a'", src, i.Name)
			}
			if i, isLiteral := p.Path.(*ast.Literal); !isLiteral {
				t.Errorf("ParseExpr(%q): got %v, expected ident", src, reflect.TypeOf(p.Path))
			} else if i.Value != "**" {
				t.Errorf("ParseExpr(%q): got %v, expected ident '**'", src, i.Value)
			}
			if len(p.Args) != 0 {
				t.Errorf("ParseExpr(%q): got len of args %v, expected 0", src, len(p.Args))
			}
		}
	}
}

func TestPathClosure(t *testing.T) {
	src := "a.findAll {x => x + 1}"
	x, err := ParseExpr(src, ast.NewScope(map[string]types.Object{"a": types.NewString("")}))
	if err != nil {
		t.Errorf("ParseExpr(%q):%v", src, err)
	} else {
		if p, ok := x.(*ast.PathExpr); !ok {
			t.Errorf("ParseExpr(%q): got %T, expected *ast.PathExpr", src, x)
		} else {
			if _, ok := p.Path.(*ast.Ident); !ok {
				t.Errorf("ParseExpr(%q): got %T, expected *ast.Ident", src, x)
			}
			if len(p.Args) != 1 {
				t.Errorf("ParseExpr(%q): got len of args %v, expected 1", src, len(p.Args))
			}
			if c, isClosure := p.Args[0].Value.(*ast.Closure); !isClosure {
				t.Errorf("ParseExpr(%q): got %T, expected arg *ast.Closure", src, x)
			} else {
				if len(c.Params) != 1 {
					t.Errorf("ParseExpr(%q): got len of args %v, expected 1", src, len(c.Params))
				}
				if c.Params[0].Name != "x" {
					t.Errorf("ParseExpr(%q): got %v, expected ident 'x'", src, c.Params[0].Name)
				}
				if len(c.Block.Stmts) != 1 {
					t.Errorf("ParseExpr(%q): got len of stmts %v, expected 1", src, len(c.Block.Stmts))
				}
			}

		}
	}
}

func TestAssign(t *testing.T) {
	src := "a := 12"
	scope := ast.NewScope(map[string]types.Object{})
	p := newParser([]byte(src), scope)
	p.scanner.InsertLineEnd = true

	x := p.parseStatement()
	if p.errors.Err() != nil {
		t.Errorf("ParseExpr(%q):%v", src, p.errors.Err())

	}
	if a, ok := x.(*ast.Assignment); !ok {
		t.Errorf("ParseExpr(%q): got %T, expected *ast.BinaryExpr", src, x)
	} else {
		if a.Tok != token.DEFINE {
			t.Errorf("ParseExpr(%q): got %v, expected %v", src, a.Tok, token.DEFINE)
		} else {
			if _, isDefined := scope.Symbol("a"); !isDefined {
				t.Errorf("ParseExpr(%q): expected symbol 'a' is defined in scope", src)
			}
		}
		if l, isLiteral := a.Rhs.(*ast.Literal); !isLiteral {
			t.Errorf("ParseExpr(%q): got %T, expected *ast.Literal", src, a.Rhs)
		} else if l.Value != "12" {
			t.Errorf("ParseExpr(%q): got %v, expected 12", src, l.Value)
		} else if l.Kind != token.NUMBER {
			t.Errorf("ParseExpr(%q): got %v, expected kind of literal %v", src, l.Kind, token.NUMBER)
		}
	}
}

func TestList(t *testing.T) {
	e := []string{"1", "2", "3", "4"}
	src := "[1, 2, 3, 4]"
	scope := ast.NewScope(map[string]types.Object{})
	p := newParser([]byte(src), scope)
	p.scanner.InsertLineEnd = true

	x := p.parseSequence()
	if p.errors.Err() != nil {
		t.Errorf("ParseExpr(%q):%v", src, p.errors.Err())

	}
	if x.IsMap {
		t.Errorf("ParseExpr(%q):expected not a map", src)
	}
	for i, v := range x.Values {
		if l, isLit := v.(*ast.Literal); !isLit {
			t.Errorf("ParseExpr(%q): got %T, expected *ast.Literal", src, v)
		} else {
			if l.Value != e[i] {
				t.Errorf("ParseExpr(%q): got %v, expected %v", src, l.Value, e[i])
			}
		}
	}
}

func TestMap(t *testing.T) {
	e := map[string]string{"a": "1", "b": "2", "c": "3", "z": "4"}
	src := "[a: 1, b: 2, c: 3, z: 4]"
	scope := ast.NewScope(map[string]types.Object{})
	p := newParser([]byte(src), scope)
	p.scanner.InsertLineEnd = true

	x := p.parseSequence()
	if p.errors.Err() != nil {
		t.Errorf("ParseExpr(%q):%v", src, p.errors.Err())

	}
	if !x.IsMap {
		t.Errorf("ParseExpr(%q):expected a map", src)
	}
	for _, v := range x.Values {
		if kv, isLit := v.(*ast.KeyValueExpr); !isLit {
			t.Errorf("ParseExpr(%q): got %T, expected *ast.Literal", src, v)
		} else {
			if i, isIdent := kv.Key.(*ast.Ident); !isIdent {
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

func TestParseFile(t *testing.T) {
	for _, v := range valids {
		_, err := ParseFile(v.source, ast.NewScope(v.vars))
		if err != nil {
			t.Errorf("ParseExpr(%q):%v", v.source, err)
		}
	}
}

type def struct {
	source string
	vars   []string
}

type code struct {
	source []byte
	vars   map[string]types.Object
}

var scipts = [...]def{
	{"t := params.body.'*'.findAll {x => x.'*'.any {y => y == 29}}", []string{"params"}},
	{"m := [x: t]", []string{"t"}},
	{"echo m.x", []string{"m"}},
	{"a := [1,2,3,4,5]", []string{}},
	{"a := [1,2,3,4,5]\na[0]", []string{}},
	{"wsdl := readFile './data/xml.xml'", []string{"readFile"}},
}

var valids = func() []code {
	var values []code
	for _, s := range scipts {
		vars := map[string]types.Object{}
		for _, v := range s.vars {
			vars[v] = types.NewString("")
		}
		source := fmt.Sprintf("pipeline(){stages{stage(){steps{%v\n}}}}", s.source)
		values = append(values, code{source: []byte(source), vars: vars})
	}
	return values
}()
