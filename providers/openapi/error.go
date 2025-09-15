package openapi

import (
	"fmt"
	"net/http"
)

type HttpError struct {
	StatusCode int
	Header     http.Header
	Message    string
}

func (h *HttpError) Error() string {
	return h.Message
}

func newHttpError(status int, message string) *HttpError {
	return &HttpError{
		StatusCode: status,
		Message:    message,
	}
}

func newHttpErrorf(status int, format string, args ...interface{}) *HttpError {
	return &HttpError{
		StatusCode: status,
		Message:    fmt.Sprintf(format, args...),
	}
}

func newMethodNotAllowedErrorf(methods []string, format string, args ...interface{}) *HttpError {
	return &HttpError{
		StatusCode: http.StatusMethodNotAllowed,
		Header: http.Header{
			"Allow": methods,
		},
		Message: fmt.Sprintf(format, args...),
	}
}
