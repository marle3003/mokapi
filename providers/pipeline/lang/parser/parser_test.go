package parser

import (
	"fmt"
	"mokapi/providers/pipeline/lang/ast"
	"mokapi/providers/pipeline/lang/types"
	"testing"
)

func TestParseExpr(t *testing.T) {
	src := "a + b"
	x, err := ParseExpr(src, ast.NewScope(map[string]types.Object{"a": types.NewString(""), "b": types.NewString("")}))
	if err != nil {
		t.Errorf("ParseExpr(%q):%v", src, err)
		if _, ok := x.(*ast.Binary); !ok {
			t.Errorf("ParseExpr(%q): got %T, expected *ast.BinaryExpr", src, x)
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
