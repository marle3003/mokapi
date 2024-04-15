package kafka

import (
	"context"
	"github.com/google/uuid"
	"time"
)

const clientKey = "client"

type UUIDGenerator func() string

var uuidGenerator UUIDGenerator = func() string {
	return uuid.New().String()
}

func SetUUIDGenerator(g UUIDGenerator) {
	uuidGenerator = g
}

type ClientContext struct {
	Addr                   string
	ClientId               string
	ClientSoftwareName     string
	ClientSoftwareVersion  string
	Heartbeat              time.Time
	Member                 map[string]string
	Close                  func()
	AllowAutoTopicCreation bool
}

func (c *ClientContext) AddGroup(groupName, memberId string) {
	if c.Member == nil {
		c.Member = make(map[string]string)
	}
	c.Member[groupName] = memberId
}

func (c *ClientContext) GetOrCreateMemberId(groupName string) string {
	memberId := c.Member[groupName]
	if len(memberId) == 0 {
		memberId = c.ClientId
		if len(memberId) > 0 {
			memberId += "-"
		}
		memberId += uuidGenerator()
		c.Member[groupName] = memberId
	}
	return memberId
}

func ClientFromContext(req *Request) *ClientContext {
	return req.Context.Value(clientKey).(*ClientContext)
}

func NewClientContext(ctx context.Context, addr string) context.Context {
	return context.WithValue(ctx, clientKey, &ClientContext{Addr: addr, AllowAutoTopicCreation: true, Heartbeat: time.Now()})
}
