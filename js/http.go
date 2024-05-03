package js

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/dop251/goja"
	"io"
	"mokapi/engine/common"
	"mokapi/media"
	"net/http"
)

type HttpClient interface {
	Do(r *http.Request) (*http.Response, error)
}

type httpModule struct {
	host   common.Host
	rt     *goja.Runtime
	client HttpClient
	runner *runner
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

type fetchArgs struct {
	method string
}

func newHttp(host common.Host, rt *goja.Runtime, runner *runner) interface{} {
	return &httpModule{host: host, rt: rt, client: host.HttpClient(), runner: runner}
}

func (m *httpModule) Get(url string, args goja.Value) interface{} {
	return m.doRequest("GET", url, "", args)
}

func (m *httpModule) Post(url string, body interface{}, args goja.Value) interface{} {
	return m.doRequest("POST", url, body, args)
}

func (m *httpModule) Put(url string, body interface{}, args goja.Value) interface{} {
	return m.doRequest("PUT", url, body, args)
}

func (m *httpModule) Head(url string, args goja.Value) interface{} {
	return m.doRequest("HEAD", url, "", args)
}

func (m *httpModule) Patch(url string, body interface{}, args goja.Value) interface{} {
	return m.doRequest("PATCH", url, body, args)
}

func (m *httpModule) Delete(url string, body interface{}, args goja.Value) interface{} {
	return m.doRequest("DELETE", url, body, args)
}

func (m *httpModule) Del(url string, body interface{}, args goja.Value) interface{} {
	return m.Delete(url, body, args)
}

func (m *httpModule) Options(url string, body interface{}, args goja.Value) interface{} {
	return m.doRequest("OPTIONS", url, body, args)
}

func (m *httpModule) Fetch(url string, v goja.Value) *goja.Promise {
	p, resolve, _ := m.rt.NewPromise()
	go func() {
		args := getFetchArgs(v)
		res := m.doRequest(args.method, url, nil, nil)
		m.runner.Run(func(vm *goja.Runtime) {
			resolve(res)
		})
	}()
	return p
}

func (m *httpModule) doRequest(method, url string, body interface{}, args goja.Value) interface{} {
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
		panic(m.rt.ToValue(err.Error()))
	}

	res, err := m.client.Do(req)
	if err != nil {
		panic(m.rt.ToValue(err.Error()))
	}

	return m.parseResponse(res)
}

func createRequest(method, url string, body interface{}, args *requestArgs) (*http.Request, error) {
	r, err := encode(body, args)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, r)
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

func encode(i interface{}, args *requestArgs) (io.Reader, error) {
	if s, ok := i.(string); ok {
		return bytes.NewBufferString(s), nil
	}

	h, ok := args.Headers["Content-Type"]
	if !ok {
		h = "application/json"
	}
	contentType := fmt.Sprintf("%v", h)

	ct := media.ParseContentType(contentType)
	switch {
	case ct.Subtype == "json" || ct.Subtype == "problem+json":
		b, err := json.Marshal(i)
		if err != nil {
			return nil, err
		}
		return bytes.NewReader(b), nil
	default:
		return nil, fmt.Errorf("encoding %v is not supported", contentType)
	}
}

func (r response) Json() interface{} {
	var i interface{}
	err := json.Unmarshal([]byte(r.Body), &i)
	if err != nil {
		panic(r.rt.ToValue(err.Error()))
	}
	return i
}

func getFetchArgs(v goja.Value) fetchArgs {
	return fetchArgs{method: http.MethodGet}
}
