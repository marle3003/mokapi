package openapi

import (
	"fmt"
	"net/http"
)

type httpError struct {
	StatusCode int
	Header     http.Header
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

func newMethodNotAllowedErrorf(status int, methods []string, format string, args ...interface{}) *httpError {
	return &httpError{
		StatusCode: status,
		Header: http.Header{
			"Allow": methods,
		},
		message: fmt.Sprintf(format, args...),
	}
}
