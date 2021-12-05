package web

import (
	"fmt"
	"net/url"
	"strconv"
)

type httpError struct {
	StatusCode int
	message    string
}

func (h *httpError) Error() string {
	return h.message
}

func newHttpError(status int, message string) *httpError {
	return &httpError{
		StatusCode: status,
		message:    message,
	}
}

func newHttpErrorf(status int, format string, args ...interface{}) *httpError {
	return &httpError{
		StatusCode: status,
		message:    fmt.Sprintf(format, args...),
	}
}

func ParseAddress(s string) (host string, port int, path string, err error) {
	var u *url.URL
	u, err = url.Parse(s)
	if err != nil {
		return
	}

	host = u.Hostname()

	portString := u.Port()
	if len(portString) == 0 {
		switch u.Scheme {
		case "https":
			port = 443
		default:
			port = 80
		}
	} else {
		var p int64
		p, err = strconv.ParseInt(portString, 10, 32)
		if err != nil {
			return
		}
		port = int(p)
	}

	if len(u.Path) == 0 {
		path = "/"
	} else {
		path = u.Path
	}

	return
}
