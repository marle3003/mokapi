package http

import (
	"bytes"
	"fmt"
	"github.com/dop251/goja"
	"io"
	"mokapi/engine/common"
	"net/http"
)

type Client interface {
	Do(r *http.Request) (*http.Response, error)
}

type Module struct {
	host   common.Host
	rt     *goja.Runtime
	client Client
}

type requestArgs struct {
	Headers map[string]interface{}
}

type response struct {
	Body       string
	StatusCode int
	Headers    map[string]string
}

func New(host common.Host, rt *goja.Runtime) interface{} {
	return &Module{host: host, rt: rt, client: host.HttpClient()}
}

func (m *Module) Get(url string, args goja.Value) (interface{}, error) {
	return m.doRequest("GET", url, "", args)
}

func (m *Module) Post(url, body string, args goja.Value) (interface{}, error) {
	return m.doRequest("POST", url, body, args)
}

func (m *Module) Put(url, body string, args goja.Value) (interface{}, error) {
	return m.doRequest("PUT", url, body, args)
}

func (m *Module) Head(url string, args goja.Value) (interface{}, error) {
	return m.doRequest("HEAD", url, "", args)
}

func (m *Module) Patch(url, body string, args goja.Value) (interface{}, error) {
	return m.doRequest("PATCH", url, body, args)
}

func (m *Module) Delete(url, body string, args goja.Value) (interface{}, error) {
	return m.doRequest("DELETE", url, body, args)
}

func (m *Module) Options(url, body string, args goja.Value) (interface{}, error) {
	return m.doRequest("OPTIONS", url, body, args)
}

func (m *Module) doRequest(method, url, body string, args goja.Value) (interface{}, error) {
	rArgs := &requestArgs{Headers: make(map[string]interface{})}
	if args != nil && !goja.IsUndefined(args) && !goja.IsNull(args) {
		params := args.ToObject(m.rt)
		for _, k := range params.Keys() {
			switch k {
			case "headers":
				tagsV := params.Get(k)
				if goja.IsUndefined(tagsV) || goja.IsNull(tagsV) {
					continue
				}
				tagsO := tagsV.ToObject(m.rt)
				for _, key := range tagsO.Keys() {
					rArgs.Headers[key] = tagsO.Get(key).String()
				}
			}
		}
	}

	req, err := createRequest(method, url, body, rArgs)

	if err != nil {
		return nil, err
	}

	res, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}

	return parseResponse(res), nil
}

func createRequest(method, url, body string, args *requestArgs) (*http.Request, error) {
	var br io.Reader
	if len(body) > 0 {
		br = bytes.NewBufferString(body)
	}

	req, err := http.NewRequest(method, url, br)
	if err != nil {
		return nil, err
	}

	for k, v := range args.Headers {
		if a, ok := v.([]interface{}); ok {
			for _, i := range a {
				req.Header.Add(k, fmt.Sprintf("%v", i))
			}
		} else {
			req.Header.Set(k, fmt.Sprintf("%v", v))
		}
	}

	return req, nil
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
