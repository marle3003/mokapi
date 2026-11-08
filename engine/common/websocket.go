package common

type WebsocketEventType = string

const (
	WebsocketConnectEventType WebsocketEventType = "connect"
	WebsocketCloseEventType   WebsocketEventType = "close"
	WebsocketMessageEventType WebsocketEventType = "message"
)

type WebsocketFilter struct {
	Type    WebsocketEventType
	Api     string
	Channel string
}

type WebsocketEvent struct {
	Type    WebsocketEventType     `json:"type"`
	Api     string                 `json:"api"`
	Channel *WebsocketEventChannel `json:"channel"`
	Client  *WebsocketEventClient  `json:"client"`
}

type WebsocketConnectEvent struct {
	WebsocketEvent

	Conn WebsocketClientConn `json:"-"`
}

type WebsocketCloseEvent struct {
	WebsocketEvent

	Reason   string `json:"reason"`
	ClosedBy string `json:"closedBy"`

	Conn WebsocketClientConn `json:"-"`
}

type WebsocketMessageEvent struct {
	WebsocketEvent

	Message any                 `json:"message"`
	Conn    WebsocketClientConn `json:"-"`
}

type WebsocketEventChannel struct {
	Key     string                  `json:"key"`
	Name    string                  `json:"name"`
	Clients []*WebsocketEventClient `json:"clients"`
	Params  map[string]string       `json:"params"`
	Conn    WebsocketChannelConn    `json:"-"`
}

type WebsocketEventClient struct {
	RemoteAddress string         `json:"remoteAddress"`
	Query         map[string]any `json:"query"`
	Headers       map[string]any `json:"headers"`

	Conn WebsocketClientConn `json:"-"`
}

type WebsocketChannelConn interface {
	Broadcast(message any) error
}

type WebsocketClientConn interface {
	Send(message any)
}

func (c *WebsocketEventClient) Send(message any) {
	c.Conn.Send(message)
}

func (c *WebsocketConnectEvent) Reply(message any) {
	c.Conn.Send(message)
}

func (c *WebsocketCloseEvent) Reply(message any) {
	c.Conn.Send(message)
}

func (c *WebsocketMessageEvent) Reply(message any) {
	c.Conn.Send(message)
}

func (c *WebsocketEvent) Broadcast(message any) error {
	return c.Channel.Conn.Broadcast(message)
}

func (c *WebsocketMessageEvent) Broadcast(message any) error {
	return c.Channel.Conn.Broadcast(message)
}

type WebsocketClient interface {
	Send(args WebsocketSendArgs) (*WebsocketSendResult, error)
}

type WebsocketSendArgs struct {
	Api         string
	ChannelName string
	// MessageId is the key of the message under channels.<channel>.messages
	MessageId  string
	Message    any
	ScriptFile string
	Retry      RetryArgs
}

type WebsocketSendResult struct {
	Api     string
	Channel string
	Message any
}
