package mqtttest

import (
	"context"
	"mokapi/mqtt"
	"net"
)

func NewTestClientContext() (context.Context, net.Conn) {
	server, client := net.Pipe()
	ctx := mqtt.NewClientContext(context.Background(), client)
	return ctx, server
}
