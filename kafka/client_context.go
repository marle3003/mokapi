package kafka

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
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
	Addr                  string
	ClientSoftwareName    string
	ClientSoftwareVersion string
	Member                map[string]string
	Close                 func()
	ServerAddress         string

	// clientId and heartbeat are written by the connection goroutine on every request
	// and read by the group balancer and the dashboard API.
	m         sync.RWMutex
	clientId  string
	heartbeat time.Time
}

func (c *ClientContext) ClientId() string {
	c.m.RLock()
	defer c.m.RUnlock()
	return c.clientId
}

func (c *ClientContext) SetClientId(clientId string) {
	c.m.Lock()
	defer c.m.Unlock()
	c.clientId = clientId
}

func (c *ClientContext) Heartbeat() time.Time {
	c.m.RLock()
	defer c.m.RUnlock()
	return c.heartbeat
}

func (c *ClientContext) SetHeartbeat(t time.Time) {
	c.m.Lock()
	defer c.m.Unlock()
	c.heartbeat = t
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
		memberId = c.ClientId()
		if len(memberId) > 0 {
			memberId += "-"
		}
		memberId += uuidGenerator()
		c.Member[groupName] = memberId
	}
	return memberId
}

func ClientFromContext(ctx context.Context) *ClientContext {
	return ctx.Value(clientKey).(*ClientContext)
}

func NewClientContext(ctx context.Context, addr, serverAddress string) context.Context {
	return context.WithValue(ctx, clientKey, &ClientContext{Addr: addr, ServerAddress: serverAddress, heartbeat: time.Now()})
}
