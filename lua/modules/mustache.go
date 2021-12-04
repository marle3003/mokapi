package modules

import (
	lua "github.com/yuin/gopher-lua"
	"mokapi/lib/mustache"
	"mokapi/lua/utils"
)

func MustacheLoader(state *lua.LState) int {
	exports := map[string]lua.LGFunction{
		"render": renderApi,
	}

	mod := state.SetFuncs(state.NewTable(), exports)

	state.Push(mod)
	return 1
}

func renderApi(state *lua.LState) int {
	template := state.CheckString(1)
	data := utils.MapTable(state.CheckTable(2))

	s, err := mustache.Render(template, data)
	if err != nil {
		state.Push(lua.LNil)
		state.Push(lua.LString(err.Error()))
		return 2
	}

	state.Push(lua.LString(s))

	return 1
}
