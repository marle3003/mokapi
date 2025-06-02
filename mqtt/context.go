package mqtt

import (
	"context"
)

const clientKey = "client"

type ClientContext struct {
	Addr     string
	ClientId string
	Session  *ClientSession
}

func ClientFromContext(ctx context.Context) *ClientContext {
	if ctx == nil {
		return nil
	}

	return ctx.Value(clientKey).(*ClientContext)
}

func NewClientContext(ctx context.Context, addr string) context.Context {
	return context.WithValue(ctx, clientKey, &ClientContext{Addr: addr})
}
