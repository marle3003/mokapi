package mustache

import (
	"fmt"
	lua "github.com/yuin/gopher-lua"
	"io/ioutil"
	luam "mokapi/lua"
	"mokapi/objectpath"
	"regexp"
	"strings"
)

var pattern = regexp.MustCompile(`{{\s*([\w\.]+)\s*}}`)

func Loader(state *lua.LState) int {
	exports := map[string]lua.LGFunction{
		"render":      renderApi,
		"render_file": renderFileApi,
	}

	mod := state.SetFuncs(state.NewTable(), exports)

	state.Push(mod)
	return 1
}

func renderFileApi(state *lua.LState) int {
	path := state.CheckString(1)
	data := luam.MapTable(state.CheckTable(2))

	template, err := ioutil.ReadFile(path)
	if err != nil {
		state.Push(lua.LNil)
		state.Push(lua.LString(err.Error()))
		return 2
	}

	s, err := render(string(template), data)
	if err != nil {
		state.Push(lua.LNil)
		state.Push(lua.LString(err.Error()))
		return 2
	}

	state.Push(lua.LString(s))

	return 1
}

func renderApi(state *lua.LState) int {
	template := state.CheckString(1)
	data := luam.MapTable(state.CheckTable(2))

	s, err := render(template, data)
	if err != nil {
		state.Push(lua.LNil)
		state.Push(lua.LString(err.Error()))
		return 2
	}

	state.Push(lua.LString(s))

	return 1
}

func render(template string, data interface{}) (string, error) {
	matches := pattern.FindAllStringSubmatch(template, -1)

	for _, match := range matches {
		path := strings.TrimSpace(match[1])
		var value interface{}
		var err error
		if path == "." {
			value = data
		} else {
			value, err = objectpath.Resolve(path, data)
			if err != nil {
				return "", err
			}
		}

		template = strings.ReplaceAll(template, match[0], fmt.Sprintf("%v", value))
	}

	return template, nil
}
