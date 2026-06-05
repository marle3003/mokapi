package openapi

import (
	"fmt"
	"mokapi/runtime/events"
	"net/http"
)

type HttpError struct {
	StatusCode int
	Header     http.Header
	Message    string

	Traits events.Traits
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

func (h *HttpError) WithTraits(t events.Traits) *HttpError {
	h.Traits = t
	return h
}
