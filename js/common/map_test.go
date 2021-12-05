package common

import (
	"github.com/dop251/goja"
	"mokapi/test"
	"testing"
)

type mapType struct{}

func Test(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		rt := goja.New()
		err := rt.Set("obj", Map(rt, &mapType{}))
		test.Ok(t, err)
		_, err = rt.RunString("obj.foo()")
		test.Ok(t, err)
	})
	t.Run("error not exists", func(t *testing.T) {
		rt := goja.New()
		err := rt.Set("obj", Map(rt, &mapType{}))
		test.Ok(t, err)
		_, err = rt.RunString("obj.foo2()")
		test.EqualError(t, "TypeError: Object has no member 'foo2' at <eval>:1:9(3)", err)
	})
	t.Run("error case", func(t *testing.T) {
		rt := goja.New()
		err := rt.Set("obj", Map(rt, &mapType{}))
		test.Ok(t, err)
		_, err = rt.RunString("obj.Foo()")
		test.EqualError(t, "TypeError: Object has no member 'Foo' at <eval>:1:8(3)", err)
	})
}

func (mt *mapType) Foo() {}
