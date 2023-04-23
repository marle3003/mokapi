package js

import (
	"github.com/dop251/goja"
	"github.com/stretchr/testify/require"
	"testing"
)

type mapType struct{}

func Test(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		rt := goja.New()
		err := rt.Set("obj", mapToJSValue(rt, &mapType{}))
		require.NoError(t, err)
		_, err = rt.RunString("obj.foo()")
		require.NoError(t, err)
	})
	t.Run("error not exists", func(t *testing.T) {
		rt := goja.New()
		err := rt.Set("obj", mapToJSValue(rt, &mapType{}))
		require.NoError(t, err)
		_, err = rt.RunString("obj.foo2()")
		require.EqualError(t, err, "TypeError: Object has no member 'foo2' at <eval>:1:9(3)")
	})
	t.Run("error case", func(t *testing.T) {
		rt := goja.New()
		err := rt.Set("obj", mapToJSValue(rt, &mapType{}))
		require.NoError(t, err)
		_, err = rt.RunString("obj.Foo()")
		require.EqualError(t, err, "TypeError: Object has no member 'Foo' at <eval>:1:8(3)")
	})
}

func (mt *mapType) Foo() {}
