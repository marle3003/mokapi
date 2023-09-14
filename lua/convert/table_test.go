package convert

import (
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"
	"mokapi/sortedmap"
	"testing"
)

func TestFromLua_Table(t *testing.T) {
	testcases := []struct {
		name string
		f    func(l *lua.LState)
	}{
		{
			name: "[]string",
			f: func(l *lua.LState) {
				tbl := l.NewTable()
				tbl.Insert(1, lua.LString("foo"))
				tbl.Insert(2, lua.LString("bar"))
				var a []string
				err := FromLua(tbl, &a)
				require.NoError(t, err)
				require.Equal(t, []string{"foo", "bar"}, a)
			},
		},
		{
			name: "[]int",
			f: func(l *lua.LState) {
				tbl := l.NewTable()
				tbl.Insert(1, lua.LNumber(1))
				tbl.Insert(2, lua.LNumber(2))
				tbl.Insert(3, lua.LNumber(3))
				tbl.Insert(4, lua.LNumber(5))
				var a []int
				err := FromLua(tbl, &a)
				require.NoError(t, err)
				require.Equal(t, []int{1, 2, 3, 5}, a)
			},
		},
		{
			name: "[]interface from string",
			f: func(l *lua.LState) {
				tbl := l.NewTable()
				tbl.Insert(1, lua.LString("foo"))
				tbl.Insert(2, lua.LString("bar"))
				var a []interface{}
				err := FromLua(tbl, &a)
				require.NoError(t, err)
				require.Equal(t, []interface{}{"foo", "bar"}, a)
			},
		},
		{
			name: "[]interface from number",
			f: func(l *lua.LState) {
				tbl := l.NewTable()
				tbl.Insert(1, lua.LNumber(1))
				tbl.Insert(2, lua.LNumber(2))
				tbl.Insert(3, lua.LNumber(3))
				tbl.Insert(4, lua.LNumber(5))
				var a []interface{}
				err := FromLua(tbl, &a)
				require.NoError(t, err)
				require.Equal(t, []interface{}{1.0, 2.0, 3.0, 5.0}, a)
			},
		},
		{
			name: "interface as []interface{}",
			f: func(l *lua.LState) {
				tbl := l.NewTable()
				tbl.Insert(1, lua.LNumber(1))
				tbl.Insert(2, lua.LNumber(2))
				tbl.Insert(3, lua.LNumber(3))
				tbl.Insert(4, lua.LNumber(5))
				var a interface{}
				err := FromLua(tbl, &a)
				require.NoError(t, err)
				require.Equal(t, []interface{}{1.0, 2.0, 3.0, 5.0}, a)
			},
		},
		{
			name: "map",
			f: func(l *lua.LState) {
				tbl := l.NewTable()
				l.SetField(tbl, "foo", lua.LString("123"))
				l.SetField(tbl, "bar", lua.LString("58"))
				var m map[string]string
				err := FromLua(tbl, &m)
				require.NoError(t, err)
				require.Len(t, m, 2)
				require.Contains(t, m, "foo")
				require.Contains(t, m, "bar")
				require.Equal(t, "123", m["foo"])
				require.Equal(t, "58", m["bar"])
			},
		},
		{
			name: "map interface{}",
			f: func(l *lua.LState) {
				tbl := l.NewTable()
				l.SetField(tbl, "foo", lua.LString("123"))
				l.SetField(tbl, "bar", lua.LString("58"))
				var m map[string]interface{}
				err := FromLua(tbl, &m)
				require.NoError(t, err)
				require.Len(t, m, 2)
				require.Contains(t, m, "foo")
				require.Contains(t, m, "bar")
				require.Equal(t, "123", m["foo"])
				require.Equal(t, "58", m["bar"])
			},
		},
		{
			name: "sortedmap",
			f: func(l *lua.LState) {
				tbl := l.NewTable()
				l.SetField(tbl, "foo", lua.LString("123"))
				l.SetField(tbl, "bar", lua.LString("456"))
				var m *sortedmap.LinkedHashMap[string, interface{}]
				err := FromLua(tbl, &m)
				require.NoError(t, err)
				require.Equal(t, []string{"foo", "bar"}, m.Keys())
				require.Equal(t, []interface{}{"123", "456"}, m.Values())
			},
		},
		{
			name: "interface as sortedmap",
			f: func(l *lua.LState) {
				tbl := l.NewTable()
				l.SetField(tbl, "foo", lua.LString("123"))
				l.SetField(tbl, "bar", lua.LString("456"))
				var i interface{}
				err := FromLua(tbl, &i)
				require.NoError(t, err)
				m := i.(*sortedmap.LinkedHashMap[string, interface{}])
				require.Equal(t, []string{"foo", "bar"}, m.Keys())
				require.Equal(t, []interface{}{"123", "456"}, m.Values())
			},
		},
		{
			name: "empty table",
			f: func(l *lua.LState) {
				tbl := l.NewTable()
				var m *sortedmap.LinkedHashMap[string, interface{}]
				err := FromLua(tbl, &m)
				require.NoError(t, err)
				require.Equal(t, 0, m.Len())
			},
		},
		{
			name: "nil table",
			f: func(l *lua.LState) {
				var m *sortedmap.LinkedHashMap[string, interface{}]
				err := FromLua(lua.LNil, &m)
				require.NoError(t, err)
				require.Nil(t, m)
			},
		},
		{
			name: "struct",
			f: func(l *lua.LState) {
				tbl := l.NewTable()
				l.SetField(tbl, "name", lua.LString("foo"))
				l.SetField(tbl, "value", lua.LNumber(42))

				var f foo
				err := FromLua(tbl, &f)
				require.NoError(t, err)
				require.Equal(t, "foo", f.Name)
				require.Equal(t, 42, f.Value)
			},
		},
		{
			name: "pointer",
			f: func(l *lua.LState) {
				tbl := l.NewTable()
				l.SetField(tbl, "name", lua.LString("foo"))
				l.SetField(tbl, "value", lua.LNumber(42))

				f := &foo{}
				err := FromLua(tbl, &f)
				require.NoError(t, err)
				require.Equal(t, "foo", f.Name)
				require.Equal(t, 42, f.Value)
			},
		},
		{
			name: "pointer",
			f: func(l *lua.LState) {
				tbl := l.NewTable()
				l.SetField(tbl, "name", lua.LString("foo"))
				l.SetField(tbl, "value", lua.LNumber(42))

				var f *foo
				err := FromLua(tbl, &f)
				require.NoError(t, err)
				require.Equal(t, "foo", f.Name)
				require.Equal(t, 42, f.Value)
			},
		},
		{
			name: "nested",
			f: func(l *lua.LState) {
				tbl := l.NewTable()
				l.SetField(tbl, "name", lua.LString("foo"))
				l.SetField(tbl, "value", lua.LNumber(42))
				nested := l.NewTable()
				l.SetField(nested, "flag", lua.LTrue)
				l.SetField(tbl, "ref", nested)
				l.SetField(tbl, "ref2", nested)

				f := &foo{}
				err := FromLua(tbl, &f)
				require.NoError(t, err)
				require.Equal(t, "foo", f.Name)
				require.Equal(t, 42, f.Value)
				require.True(t, f.Ref.Flag)
				require.True(t, f.Ref2.Flag)
			},
		},
		{
			name: "struct with default",
			f: func(l *lua.LState) {
				tbl := l.NewTable()
				l.SetField(tbl, "name", lua.LString("foo"))

				f := &foo{Value: 42}
				err := FromLua(tbl, &f)
				require.NoError(t, err)
				require.Equal(t, "foo", f.Name)
				require.Equal(t, 42, f.Value)
				require.Nil(t, f.Ref)
				require.NotNil(t, f.Ref2)
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

type foo struct {
	Name  string
	Value int
	Ref   *bar
	Ref2  bar
}

type bar struct {
	Flag bool
}
