package kafkatest

import (
	"fmt"
	"mokapi/kafka/protocol"
	"mokapi/kafka/protocol/apiVersion"
	"mokapi/kafka/protocol/fetch"
	"mokapi/kafka/protocol/findCoordinator"
	"mokapi/kafka/protocol/heartbeat"
	"mokapi/kafka/protocol/joinGroup"
	"mokapi/kafka/protocol/listgroup"
	"mokapi/kafka/protocol/metaData"
	"mokapi/kafka/protocol/offset"
	"mokapi/kafka/protocol/offsetCommit"
	"mokapi/kafka/protocol/offsetFetch"
	"mokapi/kafka/protocol/produce"
	"mokapi/kafka/protocol/syncGroup"
	"net"
	"reflect"
	"time"
)

// Client is not thread-safe
type Client struct {
	Addr    string
	Timeout time.Duration

	conn          net.Conn
	clientId      string
	correlationId int32
}

func NewClient(addr, clientId string) *Client {
	return &Client{Addr: addr, clientId: clientId, Timeout: time.Second * 10}
}

func (c *Client) Close() {
	if c.conn == nil {
		return
	}
	err := c.conn.Close()
	if err != nil {
		panic(err)
	}
}

func (c *Client) Send(r *protocol.Request) (*protocol.Response, error) {
	var err error
	if c.conn == nil {
		d := net.Dialer{Timeout: c.Timeout}
		c.conn, err = d.Dial("tcp", c.Addr)
		if err != nil {
			return nil, err
		}
	}

	r.Header.CorrelationId = c.correlationId
	c.correlationId++
	err = r.Write(c.conn)
	if err != nil {
		return nil, err
	}

	res := protocol.NewResponse(r.Header.ApiKey, r.Header.ApiVersion, r.Header.CorrelationId)
	c.conn.SetReadDeadline(time.Now().Add(c.Timeout))
	err = res.Read(c.conn)
	return res, err
}

func (c *Client) ApiVersion(version int, r *apiVersion.Request) (*apiVersion.Response, error) {
	res, err := c.Send(NewRequest(c.clientId, version, r))
	if err != nil {
		return nil, err
	}
	if msg, ok := res.Message.(*apiVersion.Response); ok {
		return msg, nil
	}
	return nil, fmt.Errorf("unexpected response message: %T", res.Message)
}

func (c *Client) Metadata(version int, r *metaData.Request) (*metaData.Response, error) {
	res, err := c.Send(NewRequest(c.clientId, version, r))
	if err != nil {
		return nil, err
	}
	if msg, ok := res.Message.(*metaData.Response); ok {
		return msg, nil
	}
	return nil, fmt.Errorf("unexpected response message: %T", res.Message)
}

func (c *Client) Produce(version int, r *produce.Request) (*produce.Response, error) {
	res, err := c.Send(NewRequest(c.clientId, version, r))
	if err != nil {
		return nil, err
	}
	if msg, ok := res.Message.(*produce.Response); ok {
		return msg, nil
	}
	return nil, fmt.Errorf("unexpected response message: %T", res.Message)
}

func (c *Client) Fetch(version int, r *fetch.Request) (*fetch.Response, error) {
	res, err := c.Send(NewRequest(c.clientId, version, r))
	if err != nil {
		return nil, err
	}
	if msg, ok := res.Message.(*fetch.Response); ok {
		return msg, nil
	}
	return nil, fmt.Errorf("unexpected response message: %T", res.Message)
}

func (c *Client) OffsetFetch(version int, r *offsetFetch.Request) (*offsetFetch.Response, error) {
	res, err := c.Send(NewRequest(c.clientId, version, r))
	if err != nil {
		return nil, err
	}
	if msg, ok := res.Message.(*offsetFetch.Response); ok {
		return msg, nil
	}
	return nil, fmt.Errorf("unexpected response message: %T", res.Message)
}

func (c *Client) Offset(version int, r *offset.Request) (*offset.Response, error) {
	res, err := c.Send(NewRequest(c.clientId, version, r))
	if err != nil {
		return nil, err
	}
	if msg, ok := res.Message.(*offset.Response); ok {
		return msg, nil
	}
	return nil, fmt.Errorf("unexpected response message: %T", res.Message)
}

func (c *Client) JoinGroup(version int, r *joinGroup.Request) (*joinGroup.Response, error) {
	res, err := c.Send(NewRequest(c.clientId, version, r))
	if err != nil {
		return nil, err
	}
	if msg, ok := res.Message.(*joinGroup.Response); ok {
		return msg, nil
	}
	return nil, fmt.Errorf("unexpected response message: %T", res.Message)
}

func (c *Client) SyncGroup(version int, r *syncGroup.Request) (*syncGroup.Response, error) {
	res, err := c.Send(NewRequest(c.clientId, version, r))
	if err != nil {
		return nil, err
	}
	if msg, ok := res.Message.(*syncGroup.Response); ok {
		return msg, nil
	}
	return nil, fmt.Errorf("unexpected response message: %T", reflect.ValueOf(res.Message).Elem().Type())
}

func (c *Client) Heartbeat(version int, r *heartbeat.Request) (*heartbeat.Response, error) {
	res, err := c.Send(NewRequest(c.clientId, version, r))
	if err != nil {
		return nil, err
	}
	if msg, ok := res.Message.(*heartbeat.Response); ok {
		return msg, nil
	}
	return nil, fmt.Errorf("unexpected response message: %T", res.Message)
}

func (c *Client) FindCoordinator(version int, r *findCoordinator.Request) (*findCoordinator.Response, error) {
	res, err := c.Send(NewRequest(c.clientId, version, r))
	if err != nil {
		return nil, err
	}
	if msg, ok := res.Message.(*findCoordinator.Response); ok {
		return msg, nil
	}
	return nil, fmt.Errorf("unexpected response message: %T", res.Message)
}

func (c *Client) OffsetCommit(version int, r *offsetCommit.Request) (*offsetCommit.Response, error) {
	res, err := c.Send(NewRequest(c.clientId, version, r))
	if err != nil {
		return nil, err
	}
	if msg, ok := res.Message.(*offsetCommit.Response); ok {
		return msg, nil
	}
	return nil, fmt.Errorf("unexpected response message: %T", res.Message)
}

func (c *Client) Listgroup(version int, r *listgroup.Request) (*listgroup.Response, error) {
	res, err := c.Send(NewRequest(c.clientId, version, r))
	if err != nil {
		return nil, err
	}
	if msg, ok := res.Message.(*listgroup.Response); ok {
		return msg, nil
	}
	return nil, fmt.Errorf("unexpected response message: %T", res.Message)
}

func (c *Client) JoinSyncGroup(member, group string, joinVersion, syncVersion int) error {
	join, err := c.JoinGroup(joinVersion, &joinGroup.Request{
		GroupId:      group,
		MemberId:     member,
		ProtocolType: "consumer",
		Protocols: []joinGroup.Protocol{{
			Name: "range",
		}},
	})
	if err != nil {
		return err
	} else if join.ErrorCode != protocol.None {
		return fmt.Errorf("join error code: %v", join.ErrorCode)
	}
	sync, err := c.SyncGroup(syncVersion, &syncGroup.Request{
		GroupId:      group,
		MemberId:     member,
		ProtocolType: "consumer",
		GroupAssignments: []syncGroup.GroupAssignment{
			{
				MemberId:   member,
				Assignment: []byte{},
			},
		},
	})
	if err != nil {
		return err
	} else if sync.ErrorCode != protocol.None {
		return fmt.Errorf("sync error code: %v", join.ErrorCode)
	}
	return nil
}
