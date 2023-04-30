package smtp

import (
	"context"
	"net/textproto"
)

var (
	TooManyRecipients = &SMTPStatus{
		Code:   StatusActionAborted,
		Status: EnhancedStatusCode{4, 5, 3},
	}
)

type RcptRequest struct {
	To  string
	ctx context.Context
}

type RcptResponse struct {
	Result *SMTPStatus
}

func NewRcptRequest(to string, ctx context.Context) *RcptRequest {
	return &RcptRequest{
		To:  to,
		ctx: ctx,
	}
}

func (r *RcptRequest) Context() context.Context {
	return r.ctx
}

func (r *RcptRequest) WithContext(ctx context.Context) {
	r.ctx = ctx
}

func (r *RcptResponse) write(conn *textproto.Conn) error {
	return write(conn, r.Result.Code, r.Result.Status, r.Result.Message)
}

func TooManyRecipientsWithMessage(message string) *SMTPStatus {
	return &SMTPStatus{
		Code:    TooManyRecipients.Code,
		Status:  TooManyRecipients.Status,
		Message: message,
	}
}
