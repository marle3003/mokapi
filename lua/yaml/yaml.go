package yaml

import (
	lua "github.com/yuin/gopher-lua"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	luar "layeh.com/gopher-luar"
)

func Loader(state *lua.LState) int {
	exports := map[string]lua.LGFunction{
		"parse":     parse,
		"read_file": readFile,
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

func readFile(state *lua.LState) int {
	path := state.CheckString(1)

	content, err := ioutil.ReadFile(path)
	if err != nil {
		state.Push(lua.LNil)
		state.Push(lua.LString(err.Error()))
		return 2
	}

	result := make(map[string]interface{})
	err = yaml.Unmarshal([]byte(content), result)
	if err != nil {
		state.Push(lua.LNil)
		state.Push(lua.LString(err.Error()))
		return 2
	}

	state.Push(luar.New(state, result))
	return 1
}
