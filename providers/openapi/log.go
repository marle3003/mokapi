package openapi

import (
	"context"
	"encoding/json"
	"fmt"
	"mokapi/engine/common"
	"mokapi/lib"
	"mokapi/runtime/events"
	"net/http"
	"net/textproto"
	"strings"
)

const logKey = "http_log"

type HttpLog struct {
	Request    *HttpRequestLog  `json:"request"`
	Response   *HttpResponseLog `json:"response"`
	Duration   int64            `json:"duration"`
	Deprecated bool             `json:"deprecated"`
	Actions    []*common.Action `json:"actions"`
	Api        string           `json:"api"`
	Path       string           `json:"path"`
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

func NewLogEventContext(r *http.Request, deprecated bool, eh events.Handler, traits events.Traits) (context.Context, error) {
	l := &HttpLog{
		Request: &HttpRequestLog{
			Method:      r.Method,
			Url:         lib.GetUrl(r),
			ContentType: r.Header.Get("Content-Type"),
		},
		Response:   &HttpResponseLog{Headers: make(map[string]string)},
		Deprecated: deprecated,
		Api:        traits.GetName(),
		Path:       traits.Get("path"),
	}

	params, _ := FromContext(r.Context())
	if params != nil {
		l.Request.setParams("path", params.Path)
		l.Request.setParams("query", params.Query)
		l.Request.setParams("header", params.Header)
		l.Request.setParams("cookie", params.Cookie)
		if params.QueryString != nil {
			value, _ := json.Marshal(params.QueryString.Value)
			l.Request.Parameters = append(l.Request.Parameters, HttpParameter{
				Type:  "querystring",
				Value: string(value),
				Raw:   params.QueryString.Raw,
			})
		}
	}

	var parsedHeaders = map[string]bool{}
	if params != nil {
		parsedHeaders = getParsedHeaders(params.Header)
	}
	for k, v := range r.Header {
		raw := strings.Join(v, ",")
		param := HttpParameter{
			Name: k,
			Type: string(ParameterHeader),
			Raw:  &raw,
		}

		if _, ok := parsedHeaders[k]; ok {
			continue
		}
		l.Request.Parameters = append(l.Request.Parameters, param)

	}

	err := eh.Push(l, traits.WithNamespace("http"))
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

func (l *HttpLog) Title() string {
	return fmt.Sprintf("%s %s", l.Request.Method, l.Request.Url)
}

func (l *HttpRequestLog) setParams(name string, params map[string]RequestParameterValue) {
	for k, v := range params {
		value, _ := json.Marshal(v.Value)
		l.Parameters = append(l.Parameters, HttpParameter{
			Name:  k,
			Type:  name,
			Value: string(value),
			Raw:   v.Raw,
		})
	}
}

func getParsedHeaders(headers map[string]RequestParameterValue) map[string]bool {
	result := map[string]bool{}
	for k := range headers {
		result[textproto.CanonicalMIMEHeaderKey(k)] = true
	}
	return result
}
