package http

import (
	lua "github.com/yuin/gopher-lua"
	"io"
	luar "layeh.com/gopher-luar"
	"net/http"
	"strings"
	"time"
)

func Loader(state *lua.LState) int {
	exports := map[string]lua.LGFunction{
		"get":  get,
		"post": post,
	}

	mod := state.SetFuncs(state.NewTable(), exports)

	state.Push(mod)
	return 1
}

type response struct {
	Body       string
	StatusCode int
}

func get(state *lua.LState) int {
	url := state.CheckString(1)

	client := &http.Client{
		Timeout: time.Second * 30,
	}

	r, err := client.Get(url)
	response := response{StatusCode: r.StatusCode}
	if err == nil {
		if b, err := io.ReadAll(r.Body); err == nil {
			response.Body = string(b)
		}
	}

	state.Push(luar.New(state, response))
	if err != nil {
		state.Push(lua.LString(err.Error()))
		return 2
	}

	return 1
}

func post(state *lua.LState) int {
	url := state.CheckString(1)
	content := state.CheckString(2)
	contentType := "text/plain"
	if lv, ok := state.Get(3).(lua.LString); ok {
		contentType = string(lv)
	}

	client := &http.Client{
		Timeout: time.Second * 30,
	}

	r, err := client.Post(url, contentType, strings.NewReader(content))
	response := response{StatusCode: r.StatusCode}
	if err == nil {
		if b, err := io.ReadAll(r.Body); err == nil {
			response.Body = string(b)
		}
	}

	state.Push(luar.New(state, response))
	if err != nil {
		state.Push(lua.LString(err.Error()))
		return 2
	}

	return 1
}
