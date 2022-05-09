package convert

import (
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"
	"reflect"
	"testing"
)

func TestToLua(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T, l *lua.LState)
	}{
		{
			name: "bool",
			f: func(t *testing.T, l *lua.LState) {
				lv, err := ToLua(l, true)
				require.NoError(t, err)
				require.Equal(t, lua.LTrue, lv)
			},
		},
		{
			name: "string",
			f: func(t *testing.T, l *lua.LState) {
				lv, err := ToLua(l, "foo bar")
				require.NoError(t, err)
				require.Equal(t, lua.LString("foo bar"), lv)
			},
		},
		{
			name: "int",
			f: func(t *testing.T, l *lua.LState) {
				lv, err := ToLua(l, 42)
				require.NoError(t, err)
				require.Equal(t, lua.LNumber(42), lv)
			},
		},
		{
			name: "int",
			f: func(t *testing.T, l *lua.LState) {
				lv, err := ToLua(l, 2.71828)
				require.NoError(t, err)
				require.Equal(t, lua.LNumber(2.71828), lv)
			},
		},
		{
			name: "nil",
			f: func(t *testing.T, l *lua.LState) {
				lv, err := ToLua(l, nil)
				require.NoError(t, err)
				require.Equal(t, lua.LNil, lv)
			},
		},
		{
			name: "array",
			f: func(t *testing.T, l *lua.LState) {
				a := []string{"foo", "bar"}
				lv, err := ToLua(l, a)
				require.NoError(t, err)
				tbl, ok := lv.(*lua.LTable)
				require.True(t, ok)
				require.Equal(t, 2, tbl.Len())
				require.Equal(t, "foo", tbl.RawGet(lua.LNumber(1)).String())
				require.Equal(t, "bar", tbl.RawGet(lua.LNumber(2)).String())
			},
		},
		{
			name: "map",
			f: func(t *testing.T, l *lua.LState) {
				lv, err := ToLua(l, map[string]string{"foo": "123", "bar": "456"})
				require.NoError(t, err)
				tbl, ok := lv.(*lua.LTable)
				require.True(t, ok)
				require.Equal(t, "123", tbl.RawGet(lua.LString("foo")).String())
				require.Equal(t, "456", tbl.RawGet(lua.LString("bar")).String())
			},
		},
		{
			name: "struct",
			f: func(t *testing.T, l *lua.LState) {
				lv, err := ToLua(l, &struct {
					Name  string
					Valid bool
				}{"foobar", true})
				require.NoError(t, err)
				ud, ok := lv.(*lua.LUserData)
				require.True(t, ok)
				v := reflect.ValueOf(ud.Value).Elem()
				require.Equal(t, 2, v.NumField())
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			l := lua.NewState()
			tc.f(t, l)
		})
	}
}

func TestFromValue(t *testing.T) {
	testcases := []struct {
		name string
		f    func(l *lua.LState)
	}{
		{
			name: "nil",
			f: func(l *lua.LState) {
				var b bool
				err := FromLua(lua.LNil, &b)
				require.NoError(t, err)
				require.Equal(t, false, b)
			},
		},
		{
			name: "bool",
			f: func(l *lua.LState) {
				var b bool
				err := FromLua(lua.LTrue, &b)
				require.NoError(t, err)
				require.Equal(t, true, b)
			},
		},
		{
			name: "string",
			f: func(l *lua.LState) {
				var s string
				err := FromLua(lua.LString("foo bar"), &s)
				require.NoError(t, err)
				require.Equal(t, "foo bar", s)
			},
		},
		{
			name: "int",
			f: func(l *lua.LState) {
				var i int
				err := FromLua(lua.LNumber(42), &i)
				require.NoError(t, err)
				require.Equal(t, 42, i)
			},
		},
		{
			name: "int32",
			f: func(l *lua.LState) {
				var i int32
				err := FromLua(lua.LNumber(42), &i)
				require.NoError(t, err)
				require.Equal(t, int32(42), i)
			},
		},
		{
			name: "float",
			f: func(l *lua.LState) {
				var f float32
				err := FromLua(lua.LNumber(2.71828), &f)
				require.NoError(t, err)
				require.Equal(t, float32(2.71828), f)
			},
		},
		{
			name: "float to int",
			f: func(l *lua.LState) {
				var i int
				err := FromLua(lua.LNumber(2.71828), &i)
				require.NoError(t, err)
				require.Equal(t, 2, i)
			},
		},
		{
			name: "number to interface",
			f: func(l *lua.LState) {
				var i interface{}
				err := FromLua(lua.LNumber(2.71828), &i)
				require.NoError(t, err)
				require.Equal(t, 2.71828, i)
			},
		},
		{
			name: "not supported number",
			f: func(l *lua.LState) {
				var i uint32
				err := FromLua(lua.LNumber(42), &i)
				require.EqualError(t, err, "unable to convert number to uint32")
			},
		},
	}

	for _, testcase := range testcases {
		test := testcase
		t.Run(test.name, func(t *testing.T) {
			l := lua.NewState()
			test.f(l)
		})
	}
}
