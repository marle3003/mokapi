package utils

import (
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"
	"mokapi/sortedmap"
	"reflect"
	"testing"
)

func TestMapTable(t *testing.T) {
	l := lua.NewState()
	tbl := l.NewTable()
	l.SetField(tbl, "foo", lua.LString("123"))
	l.SetField(tbl, "bar", lua.LString("456"))
	o := MapTable(tbl)
	_ = o
}

func TestToValue(t *testing.T) {
	testcases := []struct {
		name   string
		val    interface{}
		verify func(t *testing.T, lv lua.LValue)
	}{
		{
			name: "bool",
			val:  true,
			verify: func(t *testing.T, lv lua.LValue) {
				require.Equal(t, lua.LTrue, lv)
			},
		},
		{
			name: "array",
			val:  []string{"foo", "bar"},
			verify: func(t *testing.T, lv lua.LValue) {
				tbl, ok := lv.(*lua.LTable)
				require.True(t, ok)
				require.Equal(t, 2, tbl.Len())
				require.Equal(t, "foo", tbl.RawGet(lua.LNumber(1)).String())
				require.Equal(t, "bar", tbl.RawGet(lua.LNumber(2)).String())
			},
		},
		{
			name: "map",
			val:  map[string]string{"foo": "123", "bar": "456"},
			verify: func(t *testing.T, lv lua.LValue) {
				tbl, ok := lv.(*lua.LTable)
				require.True(t, ok)
				require.Equal(t, "123", tbl.RawGet(lua.LString("foo")).String())
				require.Equal(t, "456", tbl.RawGet(lua.LString("bar")).String())
			},
		},
		{
			name: "map",
			val: &struct {
				Name  string
				Valid bool
			}{"foobar", true},
			verify: func(t *testing.T, lv lua.LValue) {
				ud, ok := lv.(*lua.LUserData)
				require.True(t, ok)
				v := reflect.ValueOf(ud.Value).Elem()
				require.Equal(t, 2, v.NumField())
			},
		},
	}

	for _, testcase := range testcases {
		test := testcase
		t.Run(test.name, func(t *testing.T) {
			l := lua.NewState()
			v := ToValue(l, test.val)
			test.verify(t, v)
		})
	}
}

func TestFromValue(t *testing.T) {
	testcases := []struct {
		name string
		fn   func(l *lua.LState)
	}{
		{
			name: "bool",
			fn: func(l *lua.LState) {
				v := FromValue(lua.LTrue)
				require.Equal(t, true, v)
			},
		},
		{
			name: "array",
			fn: func(l *lua.LState) {
				tbl := l.NewTable()
				tbl.Insert(1, lua.LString("foo"))
				tbl.Insert(2, lua.LString("bar"))
				v := FromValue(tbl)
				require.Equal(t, []interface{}{"foo", "bar"}, v)
			},
		},
		{
			name: "map",
			fn: func(l *lua.LState) {
				tbl := l.NewTable()
				l.SetField(tbl, "foo", lua.LString("123"))
				l.SetField(tbl, "bar", lua.LString("456"))
				v := FromValue(tbl)
				m, ok := v.(*sortedmap.LinkedHashMap)
				require.True(t, ok, "should be LinkedHashMap")
				require.Equal(t, []interface{}{"123", "456"}, m.Values())
			},
		},
	}

	for _, testcase := range testcases {
		test := testcase
		t.Run(test.name, func(t *testing.T) {
			l := lua.NewState()
			test.fn(l)
		})
	}
}

func TestMap_UserData(t *testing.T) {
	r := &struct {
		Name string
		Data interface{}
	}{}

	l := lua.NewState()
	tblXy := l.NewTable()
	tblXy.Append(lua.LNumber(5))
	tblXy.Append(lua.LNumber(6))

	tblData := l.NewTable()
	l.SetField(tblData, "xy", tblXy)

	ud := l.NewUserData()
	ud.Value = &struct {
		Name lua.LValue
		Data lua.LValue
	}{
		lua.LString("foo"),
		tblData,
	}

	err := Map(r, ud)
	require.NoError(t, err)
	require.Equal(t, "foo", r.Name)
	require.Equal(t, []interface{}{"xy"}, r.Data.(*sortedmap.LinkedHashMap).Keys())
	require.Equal(t, []interface{}{[]interface{}{int64(5), int64(6)}}, r.Data.(*sortedmap.LinkedHashMap).Values())
}

func TestMap_Table(t *testing.T) {
	r := &struct {
		Name string
		Data interface{}
	}{}

	l := lua.NewState()
	tblXy := l.NewTable()
	tblXy.Append(lua.LNumber(5))
	tblXy.Append(lua.LNumber(6))

	tblData := l.NewTable()
	l.SetField(tblData, "xy", tblXy)

	tbl := l.NewTable()
	l.SetField(tbl, "name", lua.LString("foo"))
	l.SetField(tbl, "Data", tblData)

	err := Map(r, tbl)
	require.NoError(t, err)
	require.Equal(t, "foo", r.Name)
	require.Equal(t, []interface{}{"xy"}, r.Data.(*sortedmap.LinkedHashMap).Keys())
	require.Equal(t, []interface{}{[]interface{}{int64(5), int64(6)}}, r.Data.(*sortedmap.LinkedHashMap).Values())
}
