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

				var m *sortedmap.LinkedHashMap[string, interface{}]
				err := FromLua(ud, &m)
				require.NoError(t, err)
				require.Equal(t, 3, m.Len())
				name, _ := m.Get("Name")
				require.Equal(t, "foo", name)
				n, _ := m.Get("Nil")
				require.Equal(t, nil, n)

				data, _ := m.Get("Data")
				require.IsType(t, &sortedmap.LinkedHashMap[string, interface{}]{}, data)

				xy, _ := data.(*sortedmap.LinkedHashMap[string, interface{}]).Get("xy")
				require.IsType(t, []interface{}{}, xy)
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
				require.Equal(t, []string{"xy"}, v.Data.(*sortedmap.LinkedHashMap[string, interface{}]).Keys())
				require.Equal(t, []interface{}{[]interface{}{5.0, 6.0}}, v.Data.(*sortedmap.LinkedHashMap[string, interface{}]).Values())
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
