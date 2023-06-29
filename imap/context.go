package imap

import "context"

const clientKey = "client"

type ClientContext struct {
	Addr string
}

func NewClientContext(ctx context.Context, addr string) context.Context {
	return context.WithValue(ctx, clientKey, &ClientContext{Addr: addr})
}
