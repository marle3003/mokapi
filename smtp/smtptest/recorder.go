package smtptest

import (
	"mokapi/smtp"
)

type ResponseRecorder struct {
	Response smtp.Response
}

func NewRecorder() *ResponseRecorder {
	return &ResponseRecorder{}
}

func (r *ResponseRecorder) Write(res smtp.Response) error {
	r.Response = res
	return nil
}
