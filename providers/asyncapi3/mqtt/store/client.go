package store

import "mokapi/mqtt"

type Client struct {
	Id     string
	Clean  bool
	writer mqtt.ResponseWriter
}

func (c *Client) send(msg *Message) {
	//c.writer.Write(mqtt.PUBLISH, msg)
}
