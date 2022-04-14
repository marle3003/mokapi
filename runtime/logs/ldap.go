package logs

import "time"

type LdapLog struct {
	Id       string
	Service  string
	Time     int64
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

func NewLdapLog(service, method string) *LdapLog {
	return &LdapLog{
		Service: service,
		Request: &LdapRequestLog{
			Method: method,
		},
		Response: &LdapResponseLog{},
		Time:     time.Now().Unix(),
	}
}
