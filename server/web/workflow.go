package web

import (
	"fmt"
	"mokapi/config/dynamic/openapi"
	"strings"
)

type Response struct {
	Headers    map[string]string `js:"headers"`
	StatusCode int               `js:"statusCode"`
	Body       string            `js:"body"`
	Data       interface{}       `js:"data"`
}

type Request struct {
	Method string                 `js:"method"`
	Url    Url                    `js:"url"`
	Body   interface{}            `js:"body"`
	Path   map[string]interface{} `js:"path"`
	Query  map[string]interface{} `js:"query"`
	Header map[string]interface{} `js:"header"`
	Cookie map[string]interface{} `js:"cookie"`

	Key         string `js:"key"`
	OperationId string `js:"operationId"`
}

type Url struct {
	Scheme string
	Host   string
	Path   string
	Query  string
}

func newRequest(ctx *HttpContext) *Request {
	r := &Request{
		Key:         ctx.EndpointPath,
		OperationId: ctx.Operation.OperationId,
		Method:      ctx.Request.Method,
		Path:        make(map[string]interface{}),
		Query:       make(map[string]interface{}),
		Header:      make(map[string]interface{}),
		Cookie:      make(map[string]interface{}),
	}
	for t, values := range ctx.Parameters {
		for k, v := range values {
			switch t {
			case openapi.PathParameter:
				r.Path[k] = v.Value
			case openapi.QueryParameter:
				r.Query[k] = v.Value
			case openapi.HeaderParameter:
				r.Header[k] = v.Value
			case openapi.CookieParameter:
				r.Cookie[k] = v.Value
			}
		}
	}

	r.Url = Url{
		Scheme: "",
		Host:   ctx.Request.Host,
		Path:   ctx.Request.URL.Path,
		Query:  ctx.Request.URL.RawQuery,
	}

	if strings.HasPrefix(ctx.Request.Proto, "HTTPS") {
		r.Url.Scheme = "https"
	} else if strings.HasPrefix(ctx.Request.Proto, "HTTP") {
		r.Url.Scheme = "http"
	}

	return r
}

func (r *Request) String() string {
	return r.Method + " " + r.Url.String()
}

func (u *Url) String() string {
	sb := strings.Builder{}
	sb.WriteString(u.Scheme)
	if sb.Len() > 0 {
		sb.WriteString("://")
	}
	sb.WriteString(u.Host)
	sb.WriteString(u.Path)
	sb.WriteString(u.Query)
	return sb.String()
}

//func (r *Response) Run(ctx *runtime.ActionContext) error {
//	if data, ok := ctx.GetInput("data"); ok {
//		r.Data = data
//	}
//
//	if headers, ok := ctx.GetInput("headers"); ok {
//		if m, ok := headers.(map[string]interface{}); ok {
//			for k, v := range m {
//				r.Headers[k] = fmt.Sprintf("%v", v)
//			}
//		}
//	}
//
//	if body, ok := ctx.GetInputString("body"); ok {
//		r.Body = body
//	}
//
//	if s, ok := ctx.GetInputString("contentType"); ok {
//		r.Headers["Content-Type"] = s
//	}
//
//	if s, ok := ctx.GetInputString("statusCode"); ok {
//		if i, err := strconv.Atoi(s); err != nil {
//			return err
//		} else {
//			r.StatusCode = i
//		}
//	}
//
//	return nil
//}

func (r RequestParameter) Resolve(name string) (interface{}, error) {
	if v, ok := r[name]; ok {
		return v.Value, nil
	}

	return nil, fmt.Errorf("undefined field %q", name)
}
