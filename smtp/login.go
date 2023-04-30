package smtp

import (
	"context"
	"net/textproto"
)

var (
	InvalidAuthCredentials = &SMTPStatus{
		Code:    535,
		Status:  EnhancedStatusCode{5, 7, 8},
		Message: "Authentication credentials invalid",
	}

	AuthSucceeded = &SMTPStatus{
		Code:    StatusAuthSucceeded,
		Status:  Success,
		Message: "Authentication succeeded",
	}

	AuthRequired = &SMTPStatus{
		Code:    AuthenticationRequire,
		Status:  SecurityError,
		Message: "Authentication required",
	}
)

type LoginRequest struct {
	Username string
	Password string
	ctx      context.Context
}

type LoginResponse struct {
	Result *SMTPStatus
}

func NewLoginRequest(username, password string, ctx context.Context) *LoginRequest {
	return &LoginRequest{
		Username: username,
		Password: password,
		ctx:      ctx,
	}
}

func (r *LoginRequest) Context() context.Context {
	return r.ctx
}

func (r *LoginRequest) WithContext(ctx context.Context) {
	r.ctx = ctx
}

func (r *LoginResponse) write(conn *textproto.Conn) error {
	return write(conn, r.Result.Code, r.Result.Status, r.Result.Message)
}
