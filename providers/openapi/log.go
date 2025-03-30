package openapi

import (
	"context"
	"encoding/json"
	"mokapi/engine/common"
	"mokapi/lib"
	"mokapi/providers/openapi/parameter"
	"mokapi/runtime/events"
	"net/http"
	"strings"
)

const logKey = "http_log"

type HttpLog struct {
	Request    *HttpRequestLog  `json:"request"`
	Response   *HttpResponseLog `json:"response"`
	Duration   int64            `json:"duration"`
	Deprecated bool             `json:"deprecated"`
	Actions    []*common.Action `json:"actions"`
}

type HttpRequestLog struct {
	Method      string          `json:"method"`
	Url         string          `json:"url"`
	Parameters  []HttpParameter `json:"parameters,omitempty"`
	ContentType string          `json:"contentType,omitempty"`
	Body        string          `json:"body,omitempty"`
}

type HttpResponseLog struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       string            `json:"body"`
	Size       int               `json:"size"`
}

type HttpParameter struct {
	Name  string  `json:"name"`
	Type  string  `json:"type"`
	Value string  `json:"value"`
	Raw   *string `json:"raw"`
}

func NewLogEventContext(r *http.Request, deprecated bool, traits events.Traits) (context.Context, error) {
	l := &HttpLog{
		Request: &HttpRequestLog{
			Method:      r.Method,
			Url:         lib.GetUrl(r),
			ContentType: r.Header.Get("content-type"),
		},
		Response:   &HttpResponseLog{Headers: make(map[string]string)},
		Deprecated: deprecated,
	}

	params, _ := parameter.FromContext(r.Context())
	if params != nil {
		for t, values := range params {
			for k, v := range values {
				value, _ := json.Marshal(v.Value)
				l.Request.Parameters = append(l.Request.Parameters, HttpParameter{
					Name:  k,
					Type:  string(t),
					Value: string(value),
					Raw:   v.Raw,
				})
			}
		}
	}
	for k, v := range r.Header {
		raw := strings.Join(v, ",")
		p := HttpParameter{
			Name: k,
			Type: string(parameter.Header),
			Raw:  &raw,
		}
		if params != nil {
			if pp, ok := params[parameter.Header][k]; ok {
				val, _ := json.Marshal(pp.Value)
				p.Value = string(val)
			}
		}
		l.Request.Parameters = append(l.Request.Parameters, p)
	}

	err := events.Push(l, traits.WithNamespace("http"))
	if err != nil {
		return nil, err
	}
	ctx := context.WithValue(r.Context(), logKey, l)

	return ctx, nil
}

func LogEventFromContext(ctx context.Context) (*HttpLog, bool) {
	l, ok := ctx.Value(logKey).(*HttpLog)
	return l, ok
}
