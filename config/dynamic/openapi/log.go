package openapi

import (
	"context"
	"encoding/json"
	"io"
	"mokapi/config/dynamic/openapi/parameter"
	"mokapi/runtime/events"
	"net/http"
)

const logKey = "http_log"

type HttpLog struct {
	Request  *HttpRequestLog  `json:"request"`
	Response *HttpResponseLog `json:"response"`
}

type HttpRequestLog struct {
	Method      string         `json:"method"`
	Url         string         `json:"url"`
	Parameters  []HttpParamter `json:"parameters,omitempty"`
	ContentType string         `json:"contentType,omitempty"`
	Body        string         `json:"body,omitempty"`
}

type HttpResponseLog struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       string            `json:"body"`
}

type HttpParamter struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
	Raw   string `json:"raw"`
}

func NewLogEventContext(r *http.Request, traits events.Traits) (context.Context, error) {
	body, _ := io.ReadAll(r.Body)
	l := &HttpLog{
		Request: &HttpRequestLog{
			Method:      r.Method,
			Url:         r.URL.String(),
			ContentType: r.Header.Get("content-type"),
			Body:        string(body),
		},
		Response: &HttpResponseLog{Headers: make(map[string]string)},
	}
	params, _ := parameter.FromContext(r.Context())
	for t, values := range params {
		for k, v := range values {
			value, _ := json.Marshal(v.Value)
			l.Request.Parameters = append(l.Request.Parameters, HttpParamter{
				Name:  k,
				Type:  string(t),
				Value: string(value),
				Raw:   v.Raw,
			})
		}
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
