package websocket

// ChannelParams exposes the parameters of a channel instance to the external tests.
func ChannelParams(c *Channel) map[string]string { return c.params() }

// NewClientOnChannel adds a client to the channel the way ServeHTTP does and returns the reason
// if the server closed it right away.
func NewClientOnChannel(c *Channel) (reason string, closedByServer bool) {
	client := &Client{Id: "test", channel: c, closeCh: make(chan struct{}), closing: &closeRequest{}}
	c.addClient(client)
	return client.serverCloseReason()
}
