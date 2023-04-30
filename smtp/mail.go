package smtp

import (
	"context"
	"net/textproto"
)

type SMTPStatus struct {
	Code    StatusCode
	Status  EnhancedStatusCode
	Message string
}

var (
	AddressRejected = &SMTPStatus{
		Code:    550,
		Status:  EnhancedStatusCode{5, 1, 0},
		Message: "Address rejected",
	}
	Ok = &SMTPStatus{
		Code:    250,
		Status:  Success,
		Message: "OK",
	}
)

type MailRequest struct {
	From string
	ctx  context.Context
}

type MailResponse struct {
	Result *SMTPStatus
}

func NewMailRequest(from string, ctx context.Context) *MailRequest {
	return &MailRequest{
		From: from,
		ctx:  ctx,
	}
}

func (r *MailRequest) Context() context.Context {
	return r.ctx
}

func (r *MailRequest) WithContext(ctx context.Context) {
	r.ctx = ctx
}

func (r *MailResponse) write(conn *textproto.Conn) error {
	return write(conn, r.Result.Code, r.Result.Status, r.Result.Message)
}
