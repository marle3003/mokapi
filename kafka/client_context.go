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
	Close                 func()
	ServerAddress         string

	// clientId, heartbeat and member are written by the connection goroutine on every request
	// and read by the group balancer and the dashboard API. m is a leaf lock: nothing is
	// called while holding it.
	m         sync.RWMutex
	clientId  string
	heartbeat time.Time
	member    map[string]string // group name -> member id
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
	c.m.Lock()
	defer c.m.Unlock()
	c.setMember(groupName, memberId)
}

// MemberId returns the member id of this client in the given group.
func (c *ClientContext) MemberId(groupName string) (string, bool) {
	c.m.RLock()
	defer c.m.RUnlock()
	memberId, ok := c.member[groupName]
	return memberId, ok
}

func (c *ClientContext) HasGroup(groupName string) bool {
	_, ok := c.MemberId(groupName)
	return ok
}

// Groups returns a copy of the group name to member id mapping.
func (c *ClientContext) Groups() map[string]string {
	c.m.RLock()
	defer c.m.RUnlock()
	groups := make(map[string]string, len(c.member))
	for groupName, memberId := range c.member {
		groups[groupName] = memberId
	}
	return groups
}

func (c *ClientContext) GetOrCreateMemberId(groupName string) string {
	c.m.Lock()
	defer c.m.Unlock()

	memberId := c.member[groupName]
	if len(memberId) == 0 {
		memberId = c.clientId
		if len(memberId) > 0 {
			memberId += "-"
		}
		memberId += uuidGenerator()
		c.setMember(groupName, memberId)
	}
	return memberId
}

// setMember requires c.m to be held for writing.
func (c *ClientContext) setMember(groupName, memberId string) {
	if c.member == nil {
		c.member = make(map[string]string)
	}
	c.member[groupName] = memberId
}

func ClientFromContext(ctx context.Context) *ClientContext {
	return ctx.Value(clientKey).(*ClientContext)
}

func NewClientContext(ctx context.Context, addr, serverAddress string) context.Context {
	return context.WithValue(ctx, clientKey, &ClientContext{Addr: addr, ServerAddress: serverAddress, heartbeat: time.Now()})
}
