package logs

import "time"

type HttpRequestLog struct {
	Method      string
	Url         string
	Parameters  []HttpParamter
	ContentType string
	Body        string
}

type HttpResponseLog struct {
	HttpStatus  int
	ContentType string
	Body        string
}

type HttpLog struct {
	Id       string
	Service  string
	Time     time.Time
	Duration time.Duration
	Request  HttpRequestLog
	Response HttpResponseLog
}

type HttpParamter struct {
	Name  string
	Type  string
	Value string
	Raw   string
}
