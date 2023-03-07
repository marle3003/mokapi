package openapi

import (
	"mokapi/config/dynamic/openapi/parameter"
	"mokapi/engine/common"
	"net/http"
	"strings"
)

func NewEventResponse(status int) *common.EventResponse {
	return &common.EventResponse{
		Headers:    make(map[string]string),
		StatusCode: status,
	}
}

func EventRequestFrom(r *http.Request) *common.EventRequest {
	ctx := r.Context()
	endpointPath := ctx.Value("endpointPath").(string)
	op, _ := OperationFromContext(ctx)

	req := &common.EventRequest{
		Key:         endpointPath,
		OperationId: op.OperationId,
		Method:      r.Method,
		Path:        make(map[string]interface{}),
		Query:       make(map[string]interface{}),
		Header:      make(map[string]interface{}),
		Cookie:      make(map[string]interface{}),
	}

	req.Url = common.Url{
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
