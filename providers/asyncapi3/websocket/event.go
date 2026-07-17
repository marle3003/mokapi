package websocket

import (
	engine "mokapi/engine/common"
)

func newEventChannel(ch *Channel) *engine.WebsocketEventChannel {
	ec := &engine.WebsocketEventChannel{
		Name: ch.Name,
		Conn: ch,
	}
	for _, c := range ch.clients {
		ec.Clients = append(ec.Clients, newEventClient(c))
	}
	return ec
}

func newEventClient(client *Client) *engine.WebsocketEventClient {
	c := &engine.WebsocketEventClient{
		RemoteAddress: client.RemoteAddr,
		Headers:       map[string]any{},
		Query:         map[string]any{},
		Conn:          client,
	}
	for k, v := range client.Query {
		c.Query[k] = v
	}
	for k, v := range client.Headers {
		c.Headers[k] = v
	}
	return c
}
