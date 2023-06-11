package smtp

import (
	"context"
	"net/textproto"
)

type MailRequest struct {
	From string
	ctx  context.Context
}

type MailResponse struct {
	Result *Status
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

func (r *MailRequest) NewResponse(result *Status) Response {
	return &MailResponse{Result: result}
}

func (r *MailResponse) write(conn *textproto.Conn) error {
	return write(conn, r.Result.StatusCode, r.Result.EnhancedStatusCode, r.Result.Message)
}
