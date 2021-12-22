package kafka

import "time"

type ClientContext struct {
	clientId              string
	clientSoftwareName    string
	clientSoftwareVersion string
	heartbeat             time.Time
	member                map[string]string
	close                 func()
}

func (c *ClientContext) AddGroup(groupName, memberId string) {
	if c.member == nil {
		c.member = make(map[string]string)
	}
	c.member[groupName] = memberId
}
