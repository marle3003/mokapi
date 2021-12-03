package http

import (
	"bytes"
	"fmt"
	lua "github.com/yuin/gopher-lua"
	"io"
	luar "layeh.com/gopher-luar"
	"mokapi/lua/utils"
	"net/http"
	"strings"
	"time"
)

type Client interface {
	Do(*http.Request) (*http.Response, error)
}

type Module struct {
	client Client
}

func New(client Client) *Module {
	return &Module{client: client}
}

func (m *Module) Loader(state *lua.LState) int {
	exports := map[string]lua.LGFunction{
		"get":     m.get,
		"post":    m.post,
		"put":     m.put,
		"head":    m.head,
		"patch":   m.patch,
		"delete":  m.delete,
		"options": m.options,
	}

	mod := state.SetFuncs(state.NewTable(), exports)

	state.Push(mod)
	return 1
}

type response struct {
	Body       string
	StatusCode int
	Headers    map[string]string
}

func (m *Module) get(state *lua.LState) int {
	return m.doRequest(state, "get")
}

func (m *Module) post(state *lua.LState) int {
	return m.doRequest(state, "post")
}

func (m *Module) put(state *lua.LState) int {
	return m.doRequest(state, "put")
}

func (m *Module) head(state *lua.LState) int {
	return m.doRequest(state, "head")
}

func (m *Module) patch(state *lua.LState) int {
	return m.doRequest(state, "patch")
}

func (m *Module) delete(state *lua.LState) int {
	return m.doRequest(state, "delete")
}

func (m *Module) options(state *lua.LState) int {
	return m.doRequest(state, "options")
}

func (m *Module) doRequest(state *lua.LState, method string) int {
	url := state.CheckString(1)

	body := ""
	argsIndex := 3
	if lv, ok := state.Get(2).(lua.LString); ok {
		body = lv.String()
	} else {
		argsIndex = 2
	}

	args, err := getArgs(state, argsIndex)
	if err != nil {
		state.Push(luar.New(state, nil))
		state.Push(lua.LString(err.Error()))
		return 2
	}

	req, err := createRequest(method, url, body, args)

	if err != nil {
		state.Push(luar.New(state, nil))
		state.Push(lua.LString(err.Error()))
		return 2
	}

	r, err := m.client.Do(req)
	if err != nil {
		state.Push(luar.New(state, nil))
		state.Push(lua.LString(err.Error()))
		return 2
	}

	state.Push(luar.New(state, parseResponse(r)))
	return 1
}

func createRequest(method, url, body string, args map[string]interface{}) (*http.Request, error) {
	var br io.Reader
	if len(body) > 0 {
		br = bytes.NewBufferString(body)
	}

	req, err := http.NewRequest(method, url, br)
	if err != nil {
		return nil, err
	}

	for key, value := range args {
		switch key {
		case "headers":
			headers, ok := value.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid type of headers")
			}
			for k, v := range headers {
				if a, ok := v.([]interface{}); ok {
					for _, i := range a {
						req.Header.Add(k, fmt.Sprintf("%v", i))
					}
				} else {
					req.Header.Set(k, fmt.Sprintf("%v", v))
				}
			}
		}
	}

	return req, nil
}

func getArgs(state *lua.LState, index int) (map[string]interface{}, error) {
	args := make(map[string]interface{})
	if lv, ok := state.Get(2).(*lua.LTable); ok {
		tbl := utils.MapTable(lv)
		args, ok = tbl.(map[string]interface{})
		if !ok {
			return args, fmt.Errorf("invalid type of args")
		}
	}
	return args, nil
}

func parseResponse(r *http.Response) response {
	result := response{StatusCode: r.StatusCode, Headers: make(map[string]string)}
	if r.Body != nil {
		if b, err := io.ReadAll(r.Body); err == nil {
			result.Body = string(b)
		}
	}
	for k, v := range r.Header {
		result.Headers[k] = fmt.Sprintf("%v", v)
	}
	return result
}

func put(state *lua.LState) int {
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
