package event

import (
	"mokapi/config/dynamic/mokapi"
	"mokapi/providers/workflow/path/urlpath"
	"net"
	"strings"
)

type Handler func(trigger mokapi.Trigger) bool

type HttpEvent struct {
	Method string
	Path   string
}

type SmtpEvent struct {
	Login    bool
	Received bool
	Logout   bool
	Address  string
}

func WithHttpEvent(evt HttpEvent) Handler {
	return func(t mokapi.Trigger) bool {

		if evt.isValid(t.Http) {
			return true
		}

		return false
	}
}

func WithSmtpEvent(evt SmtpEvent) Handler {
	return func(t mokapi.Trigger) bool {

		if evt.isValid(t.Smtp) {
			return true
		}

		return false
	}
}

func (e HttpEvent) isValid(t *mokapi.HttpTrigger) bool {
	if t == nil {
		return false
	}
	if len(t.Method) > 0 && t.Method != strings.ToLower(e.Method) {
		return false
	}
	if len(t.Path) > 0 && !urlpath.Match(t.Path, e.Path) {
		return false
	}

	return true
}

func (e SmtpEvent) isValid(t *mokapi.SmtpTrigger) bool {
	if t == nil {
		return false
	}
	if !matchAddress(t.Address, e.Address) {
		return false
	}

	if e.Login && t.Login {
		return true
	}
	if e.Received && t.Received {
		return true
	}
	if e.Logout && t.Logout {
		return true
	}

	return false
}

func matchAddress(a1, a2 string) bool {
	if len(a1) == 0 {
		return true
	}

	host1, port1, err1 := net.SplitHostPort(a1)
	host2, port2, err2 := net.SplitHostPort(a2)

	if err1 != nil || err2 != nil {
		return false
	}

	ip1 := net.ParseIP(host1)
	ip2 := net.ParseIP(host2)

	if host1 != "0.0.0.0" && ip1.Equal(ip2) {
		return false
	}

	if port1 != port2 {
		return false
	}

	return true
}
