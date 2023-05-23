package smtp

import (
	"context"
	"net/textproto"
)

var (
	InvalidAuthCredentials = Status{
		StatusCode:         535,
		EnhancedStatusCode: EnhancedStatusCode{5, 7, 8},
		Message:            "Authentication credentials invalid",
	}

	AuthSucceeded = Status{
		StatusCode:         StatusAuthSucceeded,
		EnhancedStatusCode: Success,
		Message:            "Authentication succeeded",
	}

	AuthRequired = Status{
		StatusCode:         AuthenticationRequire,
		EnhancedStatusCode: SecurityError,
		Message:            "Authentication required",
	}
)

type LoginRequest struct {
	Username string
	Password string
	ctx      context.Context
}

type LoginResponse struct {
	Result *Status
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

func (r *LoginRequest) NewResponse(result *Status) Response {
	return &LoginResponse{Result: result}
}

func (r *LoginResponse) write(conn *textproto.Conn) error {
	return write(conn, r.Result.StatusCode, r.Result.EnhancedStatusCode, r.Result.Message)
}
