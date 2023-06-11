package ldaptest

import (
	"context"
	"mokapi/ldap"
)

func NewRequest(messageId int64, msg ldap.Message) *ldap.Request {
	return &ldap.Request{
		Context:   context.Background(),
		MessageId: messageId,
		Message:   msg,
	}
}

type ResponseRecorder struct {
	Message ldap.Message
}

type Response struct {
	Message ldap.Message
}

func NewRecorder() *ResponseRecorder {
	return &ResponseRecorder{}
}

func (r *ResponseRecorder) Write(msg ldap.Message) error {
	r.Message = msg
	return nil
}
