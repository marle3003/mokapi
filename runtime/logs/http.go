package logs

import (
	"context"
	"github.com/google/uuid"
	"time"
)

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

type HttpLog struct {
	Id       string           `json:"id"`
	Service  string           `json:"service"`
	Time     int64            `json:"time"`
	Duration time.Duration    `json:"duration"`
	Request  *HttpRequestLog  `json:"request"`
	Response *HttpResponseLog `json:"response"`
}

type HttpParamter struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
	Raw   string `json:"raw"`
}

func NewHttpLog(method, url string) *HttpLog {
	return &HttpLog{
		Id: uuid.New().String(),
		Request: &HttpRequestLog{
			Method: method,
			Url:    url,
		},
		Response: &HttpResponseLog{
			Headers: make(map[string]string),
		},
		Time: time.Now().Unix(),
	}
}

func NewHttpLogContext(ctx context.Context, log *HttpLog) context.Context {
	return context.WithValue(ctx, "log", log)
}

func HttpLogFromContext(ctx context.Context) (*HttpLog, bool) {
	m, ok := ctx.Value("log").(*HttpLog)
	return m, ok
}
