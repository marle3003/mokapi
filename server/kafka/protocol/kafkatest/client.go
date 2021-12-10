package kafkatest

import (
	"fmt"
	"mokapi/server/kafka/protocol"
	"mokapi/server/kafka/protocol/fetch"
	"mokapi/server/kafka/protocol/findCoordinator"
	"mokapi/server/kafka/protocol/heartbeat"
	"mokapi/server/kafka/protocol/joinGroup"
	"mokapi/server/kafka/protocol/metaData"
	"mokapi/server/kafka/protocol/offset"
	"mokapi/server/kafka/protocol/offsetFetch"
	"mokapi/server/kafka/protocol/produce"
	"mokapi/server/kafka/protocol/syncGroup"
	"net"
	"reflect"
)

// Client is not thread-safe
type Client struct {
	conn          net.Conn
	clientId      string
	correlationId int32
}

func NewClient(addr, clientId string) *Client {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		panic(err)
	}
	return &Client{conn: conn, clientId: clientId}
}

func (c *Client) Close() {
	err := c.conn.Close()
	if err != nil {
		panic(err)
	}
}

func (c *Client) Send(r *protocol.Request) (*protocol.Response, error) {
	r.Header.CorrelationId = c.correlationId
	c.correlationId++
	err := r.Write(c.conn)
	if err != nil {
		return nil, err
	}

	res := protocol.NewResponse(r.Header.ApiKey, r.Header.ApiVersion, r.Header.CorrelationId)
	err = res.Read(c.conn)
	return res, err
}

func (c *Client) Metadata(version int, r *metaData.Request) (*metaData.Response, error) {
	res, err := c.Send(NewRequest(c.clientId, version, r))
	if err != nil {
		return nil, err
	}
	if msg, ok := res.Message.(*metaData.Response); ok {
		return msg, nil
	}
	return nil, fmt.Errorf("unexpected response message: %t", res.Message)
}

func (c *Client) Produce(version int, r *produce.Request) (*produce.Response, error) {
	res, err := c.Send(NewRequest(c.clientId, version, r))
	if err != nil {
		return nil, err
	}
	if msg, ok := res.Message.(*produce.Response); ok {
		return msg, nil
	}
	return nil, fmt.Errorf("unexpected response message: %t", res.Message)
}

func (c *Client) Fetch(version int, r *fetch.Request) (*fetch.Response, error) {
	res, err := c.Send(NewRequest(c.clientId, version, r))
	if err != nil {
		return nil, err
	}
	if msg, ok := res.Message.(*fetch.Response); ok {
		return msg, nil
	}
	return nil, fmt.Errorf("unexpected response message: %t", res.Message)
}

func (c *Client) OffsetFetch(version int, r *offsetFetch.Request) (*offsetFetch.Response, error) {
	res, err := c.Send(NewRequest(c.clientId, version, r))
	if err != nil {
		return nil, err
	}
	if msg, ok := res.Message.(*offsetFetch.Response); ok {
		return msg, nil
	}
	return nil, fmt.Errorf("unexpected response message: %t", res.Message)
}

func (c *Client) Offset(version int, r *offset.Request) (*offset.Response, error) {
	res, err := c.Send(NewRequest(c.clientId, version, r))
	if err != nil {
		return nil, err
	}
	if msg, ok := res.Message.(*offset.Response); ok {
		return msg, nil
	}
	return nil, fmt.Errorf("unexpected response message: %t", res.Message)
}

func (c *Client) JoinGroup(version int, r *joinGroup.Request) (*joinGroup.Response, error) {
	res, err := c.Send(NewRequest(c.clientId, version, r))
	if err != nil {
		return nil, err
	}
	if msg, ok := res.Message.(*joinGroup.Response); ok {
		return msg, nil
	}
	return nil, fmt.Errorf("unexpected response message: %t", res.Message)
}

func (c *Client) SyncGroup(version int, r *syncGroup.Request) (*syncGroup.Response, error) {
	res, err := c.Send(NewRequest(c.clientId, version, r))
	if err != nil {
		return nil, err
	}
	if msg, ok := res.Message.(*syncGroup.Response); ok {
		return msg, nil
	}
	return nil, fmt.Errorf("unexpected response message: %t", reflect.ValueOf(res.Message).Elem().Type())
}

func (c *Client) Heartbeat(version int, r *heartbeat.Request) (*heartbeat.Response, error) {
	res, err := c.Send(NewRequest(c.clientId, version, r))
	if err != nil {
		return nil, err
	}
	if msg, ok := res.Message.(*heartbeat.Response); ok {
		return msg, nil
	}
	return nil, fmt.Errorf("unexpected response message: %t", res.Message)
}

func (c *Client) FindCoordinator(version int, r *findCoordinator.Request) (*findCoordinator.Response, error) {
	res, err := c.Send(NewRequest(c.clientId, version, r))
	if err != nil {
		return nil, err
	}
	if msg, ok := res.Message.(*findCoordinator.Response); ok {
		return msg, nil
	}
	return nil, fmt.Errorf("unexpected response message: %t", res.Message)
}
