package convert

import (
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"
	"mokapi/sortedmap"
	"testing"
)

func TestFromLua_UserData(t *testing.T) {
	testcases := []struct {
		name string
		fn   func(l *lua.LState)
	}{
		{
			name: "to from",
			fn: func(l *lua.LState) {
				f := &foo{Name: "foo"}
				var err error

				lv := l.NewUserData()
				lv.Value = &struct {
					Name lua.LValue
				}{
					lua.LString("bar"),
				}

				func(i interface{}) {
					err = FromLua(lv, i)
					require.NoError(t, err)
				}(f)
				require.Equal(t, "bar", f.Name)
			},
		},
		{
			name: "userdata to sortedmap",
			fn: func(l *lua.LState) {
				tblXy := l.NewTable()
				tblXy.Append(lua.LNumber(5))
				tblXy.Append(lua.LNumber(6))

				tblData := l.NewTable()
				l.SetField(tblData, "xy", tblXy)

				ud := l.NewUserData()
				ud.Value = &struct {
					Name lua.LValue
					Data lua.LValue
					Nil  lua.LValue
				}{
					lua.LString("foo"),
					tblData,
					lua.LNil,
				}

				var m *sortedmap.LinkedHashMap
				err := FromLua(ud, &m)
				require.NoError(t, err)
				require.Equal(t, 3, m.Len())
				require.Equal(t, "foo", m.Get("Name"))
				require.Equal(t, nil, m.Get("Nil"))

				data, ok := m.Get("Data").(*sortedmap.LinkedHashMap)
				require.True(t, ok, "should be LinkedHashMap")

				xy, ok := data.Get("xy").([]interface{})
				require.True(t, ok, "should be []interface{}")
				require.Equal(t, []interface{}{5.0, 6.0}, xy)
			},
		},
		{
			name: "userdata with hint",
			fn: func(l *lua.LState) {
				type r struct {
					Name string
					Data interface{}
					Nil  interface{}
				}

				tblXy := l.NewTable()
				tblXy.Append(lua.LNumber(5))
				tblXy.Append(lua.LNumber(6))

				tblData := l.NewTable()
				l.SetField(tblData, "xy", tblXy)

				ud := l.NewUserData()
				ud.Value = &struct {
					Name lua.LValue
					Data lua.LValue
					Nil  lua.LValue
				}{
					lua.LString("foo"),
					tblData,
					lua.LNil,
				}

				v := &r{}
				err := FromLua(ud, &v)
				require.NoError(t, err)
				require.Equal(t, "foo", v.Name)
				require.Equal(t, []interface{}{"xy"}, v.Data.(*sortedmap.LinkedHashMap).Keys())
				require.Equal(t, []interface{}{[]interface{}{5.0, 6.0}}, v.Data.(*sortedmap.LinkedHashMap).Values())
				require.Nil(t, v.Nil)
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
