package store

type Client struct {
	Id     string
	Clean  bool
	Topics map[string]*SubscribedTopic
}

type SubscribedTopic struct {
	// may contain special topic wildcard characters
	Name string
	QoS  byte
}

func (c *Client) publish(topic string, payload []byte) {

}
