package kafka

import (
	"github.com/google/uuid"
	"time"
)

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

func (c *ClientContext) GetOrCreateMemberId(groupName string) string {
	memberId := c.member[groupName]
	if len(memberId) == 0 {
		memberId = c.clientSoftwareName
		if len(memberId) > 0 {
			memberId += "-"
		}
		memberId += uuid.New().String()
		c.member[groupName] = memberId
	}
	return memberId
}
