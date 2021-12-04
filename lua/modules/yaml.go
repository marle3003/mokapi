package modules

import (
	lua "github.com/yuin/gopher-lua"
	"gopkg.in/yaml.v3"
	luar "layeh.com/gopher-luar"
)

func YamlLoader(state *lua.LState) int {
	exports := map[string]lua.LGFunction{
		"parse": parse,
	}

	mod := state.SetFuncs(state.NewTable(), exports)

	state.Push(mod)
	return 1
}

func parse(state *lua.LState) int {
	content := state.CheckString(1)

	result := make(map[string]interface{})
	err := yaml.Unmarshal([]byte(content), result)

	state.Push(luar.New(state, result))
	state.Push(luar.New(state, err))

	return 2
}
