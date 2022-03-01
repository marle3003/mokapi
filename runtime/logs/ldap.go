package logs

import "time"

type LdapLog struct {
	Id       string
	Service  string
	Time     time.Time
	Duration time.Duration
	Request  *LdapRequestLog
	Response *LdapResponseLog
}

type LdapRequestLog struct {
	Method string
	Body   string
}

type LdapResponseLog struct {
	Body string
}

func NewLdapLog(service, method string) *HttpLog {
	return &HttpLog{
		Service: service,
		Request: &HttpRequestLog{
			Method: method,
		},
		Response: &HttpResponseLog{},
		Time:     time.Now(),
	}
}
