package websocket

import (
	"context"
	"mokapi/providers/asyncapi3"
)

const serverKey = "server"

func NewServerContext(ctx context.Context, server *asyncapi3.Server) context.Context {
	return context.WithValue(ctx, serverKey, server)
}

func ServerFromContext(ctx context.Context) (*asyncapi3.Server, bool) {
	o, ok := ctx.Value(serverKey).(*asyncapi3.Server)
	return o, ok
}
