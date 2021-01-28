package runtime

import (
	"mokapi/providers/pipeline/lang/ast"
	"mokapi/providers/pipeline/lang/types"
	"testing"
)

func TestStringFormat(t *testing.T) {
	data := []struct {
		format string
		out    string
		scope  *ast.Scope
	}{
		{"foobar", "foobar", ast.NewScope(map[string]types.Object{})},
		{"${foo}bar", "foobar", ast.NewScope(map[string]types.Object{"foo": types.NewString("foo")})},
		{"foo $bar", "foo bar", ast.NewScope(map[string]types.Object{"bar": types.NewString("bar")})},
		{"foo ${bar}", "foo bar", ast.NewScope(map[string]types.Object{"bar": types.NewString("bar")})},
		{"foo ${pi}", "foo 3.141", ast.NewScope(map[string]types.Object{"pi": types.NewNumber(3.141)})},
		{"foo ${math.pi}", "foo 3.141", func() *ast.Scope {
			e := types.NewExpando()
			e.SetField("pi", types.NewNumber(3.141))
			return ast.NewScope(map[string]types.Object{"math": e})
		}()},
		{"${x.foo} ${x.bar}", "foo bar", func() *ast.Scope {
			e := types.NewExpando()
			e.SetField("foo", types.NewString("foo"))
			e.SetField("bar", types.NewString("bar"))
			return ast.NewScope(map[string]types.Object{"x": e})
		}()},
	}

	for _, d := range data {
		s, err := format(d.format, d.scope)
		if err != nil {
			t.Errorf("convert(%q):%v", d.format, err.Error())
		}
		if s != d.out {
			t.Errorf("format(%q): got %q, expected %q", d.format, s, d.out)
		}
	}
}
