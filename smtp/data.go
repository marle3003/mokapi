package smtp

import (
	"context"
	"net/textproto"
)

type DataRequest struct {
	Message *Message
	ctx     context.Context
}

type DataResponse struct {
	Result *SMTPStatus
}

func NewDataRequest(msg *Message, ctx context.Context) *DataRequest {
	return &DataRequest{
		Message: msg,
		ctx:     ctx,
	}
}

func (r *DataRequest) Context() context.Context {
	return r.ctx
}

func (r *DataRequest) WithContext(ctx context.Context) {
	r.ctx = ctx
}

func (r *DataRequest) NewResponse(result *SMTPStatus) Response {
	return &DataResponse{Result: result}
}

func (r *DataResponse) write(conn *textproto.Conn) error {
	return write(conn, r.Result.Code, r.Result.Status, r.Result.Message)
}
