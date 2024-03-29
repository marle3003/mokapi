package smtp

import (
	"context"
	"net/textproto"
)

var (
	TooManyRecipients = Status{
		StatusCode:         StatusActionAborted,
		EnhancedStatusCode: EnhancedStatusCode{4, 5, 3},
	}
)

type RcptRequest struct {
	To  string
	ctx context.Context
}

type RcptResponse struct {
	Result *Status
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

func (r *RcptRequest) NewResponse(result *Status) Response {
	return &RcptResponse{Result: result}
}

func (r *RcptResponse) write(conn *textproto.Conn) error {
	return write(conn, r.Result.StatusCode, r.Result.EnhancedStatusCode, r.Result.Message)
}
