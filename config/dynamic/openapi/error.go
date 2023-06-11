package openapi

import "fmt"

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
