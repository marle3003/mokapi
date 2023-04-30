package smtp

import "context"

const clientKey = "client"

type ClientContext struct {
	Client string
	Addr   string
	Proto  string
	From   string
	To     []string
	Auth   string
}

func ClientFromContext(ctx context.Context) *ClientContext {
	return ctx.Value(clientKey).(*ClientContext)
}

func NewClientContext(ctx context.Context, addr string) context.Context {
	return context.WithValue(ctx, clientKey, &ClientContext{Addr: addr})
}

func (c *ClientContext) Reset() {
	c.From = ""
	c.To = nil
}
