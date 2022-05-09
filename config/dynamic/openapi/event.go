package openapi

import (
	"mokapi/config/dynamic/openapi/parameter"
	"net/http"
	"strings"
)

type EventResponse struct {
	Headers    map[string]string `json:"headers"`
	StatusCode int               `json:"statusCode"`
	Body       string            `json:"body"`
	Data       interface{}       `json:"data"`
}

type EventRequest struct {
	Method string                 `json:"method"`
	Url    Url                    `json:"url"`
	Body   interface{}            `json:"body"`
	Path   map[string]interface{} `json:"path"`
	Query  map[string]interface{} `json:"query"`
	Header map[string]interface{} `json:"header"`
	Cookie map[string]interface{} `json:"cookie"`

	Key         string `json:"key"`
	OperationId string `json:"operationId"`
}

type Url struct {
	Scheme string `json:"scheme"`
	Host   string `json:"host"`
	Path   string `json:"path"`
	Query  string `json:"query"`
}

func NewEventResponse(status int) *EventResponse {
	return &EventResponse{
		Headers:    make(map[string]string),
		StatusCode: status,
	}
}

func EventRequestFrom(r *http.Request) *EventRequest {
	ctx := r.Context()
	endpointPath := ctx.Value("endpointPath").(string)
	op, _ := OperationFromContext(ctx)

	req := &EventRequest{
		Key:         endpointPath,
		OperationId: op.OperationId,
		Method:      r.Method,
		Path:        make(map[string]interface{}),
		Query:       make(map[string]interface{}),
		Header:      make(map[string]interface{}),
		Cookie:      make(map[string]interface{}),
	}

	req.Url = Url{
		Scheme: "",
		Host:   r.Host,
		Path:   r.URL.Path,
		Query:  r.URL.RawQuery,
	}

	if strings.HasPrefix(r.Proto, "HTTPS") {
		req.Url.Scheme = "https"
	} else if strings.HasPrefix(r.Proto, "HTTP") {
		req.Url.Scheme = "http"
	}

	params, _ := parameter.FromContext(ctx)
	for t, values := range params {
		for k, v := range values {
			switch t {
			case parameter.Path:
				req.Path[k] = v.Value
			case parameter.Query:
				req.Query[k] = v.Value
			case parameter.Header:
				req.Header[k] = v.Value
			case parameter.Cookie:
				req.Cookie[k] = v.Value
			}
		}
	}

	return req
}

func (r *EventRequest) String() string {
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

func (r *EventResponse) HasBody() bool {
	return len(r.Body) > 0 || r.Data != nil
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
