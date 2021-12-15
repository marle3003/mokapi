package kafka

import (
	"mokapi/server/kafka/protocol"
)

type ClientContext struct {
	ctx                   protocol.Context
	clientId              string
	clientSoftwareName    string
	clientSoftwareVersion string
	member                map[string]string
	close                 func()
}

func (c *ClientContext) AddGroup(groupName, memberId string) {
	if c.member == nil {
		c.member = make(map[string]string)
	}
	c.member[groupName] = memberId
}

func (c *ClientContext) WithValue(key string, val interface{}) {
	switch key {
	case "ClientId":
		c.clientId = val.(string)
	case "ClientSoftwareName":
		c.clientSoftwareName = val.(string)
	case "ClientSoftwareVersion":
		c.clientSoftwareVersion = val.(string)
	default:
		c.ctx.WithValue(key, val)
	}
}

func (c *ClientContext) Value(key string) interface{} {
	switch key {
	case "ClientId":
		return c.clientId
	case "ClientSoftwareName":
		return c.clientSoftwareName
	case "ClientSoftwareVersion":
		return c.clientSoftwareVersion
	default:
		return c.ctx.Value(key)
	}
}

func (c *ClientContext) Close() {
	if c.close != nil {
		c.close()
	}
	c.ctx.Close()
}
