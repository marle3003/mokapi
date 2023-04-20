package js

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/dop251/goja"
	"io"
	"mokapi/engine/common"
	"net/http"
)

type HttpClient interface {
	Do(r *http.Request) (*http.Response, error)
}

type httpModule struct {
	host   common.Host
	rt     *goja.Runtime
	client HttpClient
}

type requestArgs struct {
	Headers map[string]interface{}
}

type response struct {
	rt         *goja.Runtime
	Body       string              `json:"body"`
	StatusCode int                 `json:"statusCode"`
	Headers    map[string][]string `json:"headers"`
}

func newHttp(host common.Host, rt *goja.Runtime) interface{} {
	return &httpModule{host: host, rt: rt, client: host.HttpClient()}
}

func (m *httpModule) Get(url string, args goja.Value) (interface{}, error) {
	return m.doRequest("GET", url, "", args)
}

func (m *httpModule) Post(url, body string, args goja.Value) (interface{}, error) {
	return m.doRequest("POST", url, body, args)
}

func (m *httpModule) Put(url, body string, args goja.Value) (interface{}, error) {
	return m.doRequest("PUT", url, body, args)
}

func (m *httpModule) Head(url string, args goja.Value) (interface{}, error) {
	return m.doRequest("HEAD", url, "", args)
}

func (m *httpModule) Patch(url, body string, args goja.Value) (interface{}, error) {
	return m.doRequest("PATCH", url, body, args)
}

func (m *httpModule) Delete(url, body string, args goja.Value) (interface{}, error) {
	return m.doRequest("DELETE", url, body, args)
}

func (m *httpModule) Del(url, body string, args goja.Value) (interface{}, error) {
	return m.Delete(url, body, args)
}

func (m *httpModule) Options(url, body string, args goja.Value) (interface{}, error) {
	return m.doRequest("OPTIONS", url, body, args)
}

func (m *httpModule) doRequest(method, url, body string, args goja.Value) (interface{}, error) {
	rArgs := &requestArgs{Headers: make(map[string]interface{})}
	if args != nil && !goja.IsUndefined(args) && !goja.IsNull(args) {
		params := args.ToObject(m.rt)
		for _, k := range params.Keys() {
			switch k {
			case "headers":
				headers := params.Get(k)
				if goja.IsUndefined(headers) || goja.IsNull(headers) {
					continue
				}
				rArgs.Headers = headers.Export().(map[string]interface{})
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

	return m.parseResponse(res), nil
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

func (m *httpModule) parseResponse(r *http.Response) response {
	result := response{StatusCode: r.StatusCode, Headers: make(map[string][]string), rt: m.rt}
	if r.Body != nil {
		if b, err := io.ReadAll(r.Body); err == nil {
			result.Body = string(b)
		}
	}
	for k, v := range r.Header {
		result.Headers[k] = v
	}
	return result
}

func (r response) Json() interface{} {
	var i interface{}
	err := json.Unmarshal([]byte(r.Body), &i)
	if err != nil {
		panic(r.rt.ToValue(err.Error()))
	}
	return i
}
