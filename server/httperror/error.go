package httperror

import "fmt"

type Error struct {
	StatusCode int
	message    string
}

func (h *Error) Error() string {
	return h.message
}

func New(status int, message string) *Error {
	return &Error{
		StatusCode: status,
		message:    message,
	}
}

func Newf(status int, format string, args ...interface{}) *Error {
	return &Error{
		StatusCode: status,
		message:    fmt.Sprintf(format, args...),
	}
}
