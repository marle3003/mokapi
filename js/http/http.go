package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/dop251/goja"
	"io"
	"mokapi/engine/common"
	"mokapi/js/eventloop"
	"mokapi/media"
	"net/http"
	"strings"
)

type Client interface {
	Do(r *http.Request) (*http.Response, error)
}

type Module struct {
	host common.Host
	rt   *goja.Runtime
	loop *eventloop.EventLoop
}

type RequestArgs struct {
	Headers map[string]interface{}
}

type Response struct {
	rt         *goja.Runtime
	Body       string              `json:"body"`
	StatusCode int                 `json:"statusCode"`
	Headers    map[string][]string `json:"headers"`
}

type fetchArgs struct {
	method string
}

func Require(vm *goja.Runtime, module *goja.Object) {
	o := vm.Get("mokapi/internal").(*goja.Object)
	host := o.Get("host").Export().(common.Host)
	loop := o.Get("loop").Export().(*eventloop.EventLoop)
	f := &Module{
		rt:   vm,
		host: host,
		loop: loop,
	}
	obj := module.Get("exports").(*goja.Object)
	obj.Set("get", f.Get)
	obj.Set("post", f.Post)
	obj.Set("put", f.Put)
	obj.Set("head", f.Head)
	obj.Set("patch", f.Patch)
	obj.Set("delete", f.Delete)
	obj.Set("del", f.Delete)
	obj.Set("options", f.Options)
	obj.Set("fetch", f.Fetch)
}

func (m *Module) Get(url string, args goja.Value) interface{} {
	return m.doRequest(http.MethodGet, url, "", args)
}

func (m *Module) Post(url string, body interface{}, args goja.Value) interface{} {
	return m.doRequest(http.MethodPost, url, body, args)
}

func (m *Module) Put(url string, body interface{}, args goja.Value) interface{} {
	return m.doRequest(http.MethodPut, url, body, args)
}

func (m *Module) Head(url string, args goja.Value) interface{} {
	return m.doRequest(http.MethodHead, url, "", args)
}

func (m *Module) Patch(url string, body interface{}, args goja.Value) interface{} {
	return m.doRequest(http.MethodPatch, url, body, args)
}

func (m *Module) Delete(url string, body interface{}, args goja.Value) interface{} {
	return m.doRequest(http.MethodDelete, url, body, args)
}

func (m *Module) Options(url string, body interface{}, args goja.Value) interface{} {
	return m.doRequest(http.MethodOptions, url, body, args)
}

func (m *Module) Fetch(url string, v goja.Value) *goja.Promise {
	p, resolve, reject := m.rt.NewPromise()
	go func() {
		defer func() {
			r := recover()
			if r != nil {
				m.loop.Run(func(vm *goja.Runtime) {
					reject(r)
				})
			}
		}()

		method := http.MethodGet
		var body interface{}
		if v != nil {
			obj := v.ToObject(m.rt)
			vMethod := obj.Get("method")
			if vMethod != nil {
				method = strings.ToUpper(vMethod.String())
			}
			vBody := obj.Get("body")
			if vBody != nil {
				body = vBody.Export()
			}
		}

		res := m.doRequest(method, url, body, v)
		m.loop.Run(func(vm *goja.Runtime) {
			resolve(res)
		})
	}()
	return p
}

func (m *Module) doRequest(method, url string, body interface{}, args goja.Value) Response {
	if len(url) == 0 {
		panic(m.rt.ToValue(fmt.Errorf("url cannot be empty")))
	}

	rArgs := &RequestArgs{Headers: make(map[string]interface{})}
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

	client := m.host.HttpClient()
	res, err := client.Do(req)
	if err != nil {
		panic(m.rt.ToValue(err.Error()))
	}

	return m.parseResponse(res)
}

func createRequest(method, url string, body interface{}, args *RequestArgs) (*http.Request, error) {
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
				req.Header[k] = append(req.Header[k], fmt.Sprintf("%v", i))
			}
		} else {
			req.Header[k] = []string{fmt.Sprintf("%v", v)}
		}
	}

	return req, nil
}

func (m *Module) parseResponse(r *http.Response) Response {
	result := Response{StatusCode: r.StatusCode, Headers: make(map[string][]string), rt: m.rt}
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

func encode(i interface{}, args *RequestArgs) (io.Reader, error) {
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
		return nil, fmt.Errorf("encoding request body failed: content type '%v' is not supported", contentType)
	}
}

func (r Response) Json() interface{} {
	var i interface{}
	err := json.Unmarshal([]byte(r.Body), &i)
	if err != nil {
		err = fmt.Errorf("response is not a valid JSON response: %w", err)
		panic(r.rt.ToValue(err.Error()))
	}
	return i
}
