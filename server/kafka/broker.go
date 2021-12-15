package kafka

import (
	"fmt"
	"mokapi/config/dynamic/asyncApi/kafka"
	"mokapi/server/kafka/protocol"
	"mokapi/server/kafka/protocol/apiVersion"
	"mokapi/server/kafka/protocol/fetch"
	"mokapi/server/kafka/protocol/findCoordinator"
	"mokapi/server/kafka/protocol/heartbeat"
	"mokapi/server/kafka/protocol/joinGroup"
	"mokapi/server/kafka/protocol/listgroup"
	"mokapi/server/kafka/protocol/metaData"
	"mokapi/server/kafka/protocol/offset"
	"mokapi/server/kafka/protocol/offsetCommit"
	"mokapi/server/kafka/protocol/offsetFetch"
	"mokapi/server/kafka/protocol/produce"
	"mokapi/server/kafka/protocol/syncGroup"
	"net"
	"sync"
	"time"
)

type nullCluster struct {
}

func (c *nullCluster) Topic(string) Topic {
	return nil
}

func (c *nullCluster) Topics() []Topic {
	return nil
}

func (c *nullCluster) AddTopic(string) (Topic, error) {
	return nil, fmt.Errorf("not supported")
}

func (c *nullCluster) Group(string) Group {
	return nil
}

func (c *nullCluster) Groups() []Group {
	return nil
}

func (c *nullCluster) NewGroup(string) (Group, error) {
	return nil, fmt.Errorf("not supported")
}

func (c *nullCluster) Brokers() []Broker {
	return nil
}

type BrokerServer struct {
	Id      int
	Config  kafka.BrokerBindings
	Cluster Cluster
	Clients map[net.Conn]*ClientContext

	server    *protocol.Server
	balancers map[string]*groupBalancerNew
	lock      sync.RWMutex
}

func NewBrokerServer(id int, addr string) *BrokerServer {
	b := &BrokerServer{
		Id:      id,
		Cluster: &nullCluster{},
		Clients: make(map[net.Conn]*ClientContext),
	}
	b.server = &protocol.Server{
		Addr:    addr,
		Handler: b,
		ConnContext: func(ctx protocol.Context, conn net.Conn) protocol.Context {
			cctx := &ClientContext{
				ctx: ctx,
				close: func() {
					b.lock.Lock()
					defer b.lock.Unlock()

					delete(b.Clients, conn)
					ctx.Close()
				}}
			b.Clients[conn] = cctx
			return cctx
		},
	}
	return b
}

func (b *BrokerServer) ListenAndServe() error {
	return b.server.ListenAndServe()
}

func (b *BrokerServer) Serve(l net.Listener) error {
	return b.server.Serve(l)
}

func (b *BrokerServer) Close() {
	for _, balancer := range b.balancers {
		balancer.stop <- true
	}
	b.server.Close()
}

func (b *BrokerServer) ServeMessage(rw protocol.ResponseWriter, req *protocol.Request) {
	req.Context.WithValue("heartbeat", time.Now())

	var err error
	switch req.Message.(type) {
	case *produce.Request:
		err = b.produce(rw, req)
	case *fetch.Request:
		err = b.fetch(rw, req)
	case *offset.Request:
		err = b.offset(rw, req)
	case *metaData.Request:
		err = b.metadata(rw, req)
	case *offsetCommit.Request:
		err = b.offsetCommit(rw, req)
	case *offsetFetch.Request:
		err = b.offsetFetch(rw, req)
	case *findCoordinator.Request:
		err = b.findCoordinator(rw, req)
	case *joinGroup.Request:
		err = b.joingroup(rw, req)
	case *heartbeat.Request:
		err = b.heartbeat(rw, req)
	case *syncGroup.Request:
		err = b.syncgroup(rw, req)
	case *listgroup.Request:
		err = b.listgroup(rw, req)
	case *apiVersion.Request:
		err = b.apiversion(rw, req)
	default:
		err = fmt.Errorf("unsupported api key: %v", req.Header.ApiKey)
	}

	if err != nil {
		panic(fmt.Sprintf("kafka broker: %v", err))
	}
}

func (b *BrokerServer) getBalancer(group Group) *groupBalancerNew {
	b.lock.Lock()
	defer b.lock.Unlock()

	if b.balancers == nil {
		b.balancers = make(map[string]*groupBalancerNew)
	}
	balancer, ok := b.balancers[group.Name()]
	if ok {
		return balancer
	}

	balancer = newGroupBalancerNew(group)
	b.balancers[group.Name()] = balancer
	go balancer.run()
	return balancer
}

func getClientContext(req *protocol.Request) *ClientContext {
	return req.Context.(*ClientContext)
}
