package web

import (
	"fmt"
	"mokapi/providers/workflow/runtime"
	"strconv"
)

type Response struct {
	Headers    map[string]string
	StatusCode int
	Body       string
	Data       interface{}
}

type Request struct {
	Method string
	Body   interface{}
	Path   RequestParameter
	Query  RequestParameter
	Header RequestParameter
	Cookie RequestParameter
}

func (r *Response) Run(ctx *runtime.ActionContext) error {
	if data, ok := ctx.GetInput("data"); ok {
		r.Data = data
	}
	if body, ok := ctx.GetInputString("body"); ok {
		r.Body = body
	}

	if s, ok := ctx.GetInputString("contentType"); ok {
		r.Headers["Content-Type"] = s
	}

	if s, ok := ctx.GetInputString("statusCode"); ok {
		if i, err := strconv.Atoi(s); err != nil {
			return err
		} else {
			r.StatusCode = i
		}
	}

	return nil
}

func (r RequestParameter) Resolve(name string) (interface{}, error) {
	if v, ok := r[name]; ok {
		return v.Value, nil
	}

	return nil, fmt.Errorf("undefined field %q", name)
}
