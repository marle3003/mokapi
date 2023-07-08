package imap

import "context"

const clientKey = "client"

type ClientContext struct {
	Addr    string
	Session map[string]interface{}
}

func NewClientContext(ctx context.Context, addr string) context.Context {
	return context.WithValue(ctx, clientKey, &ClientContext{
		Addr:    addr,
		Session: map[string]interface{}{},
	})
}

func ClientFromContext(ctx context.Context) *ClientContext {
	c := ctx.Value(clientKey).(*ClientContext)
	return c
}
