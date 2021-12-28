package kafka

import (
	"context"
	"fmt"
	"mokapi/kafka/protocol"
	"mokapi/kafka/protocol/apiVersion"
	"mokapi/kafka/protocol/createTopics"
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
	"mokapi/kafka/store"
	"net"
	"sync"
	"time"
)

type Broker struct {
	Id      int
	Addr    string
	Store   *store.Store
	Clients map[net.Conn]context.Context

	server    *protocol.Server
	balancers map[string]*groupBalancerNew
	lock      sync.RWMutex
}

func NewBroker(id int, addr string) *Broker {
	b := &Broker{
		Id:      id,
		Addr:    addr,
		Store:   &store.Store{},
		Clients: make(map[net.Conn]context.Context),
	}
	b.server = &protocol.Server{
		Addr:    addr,
		Handler: b,
		ConnContext: func(ctx context.Context, conn net.Conn) context.Context {
			ctx = context.WithValue(ctx, "client", &ClientContext{})
			b.Clients[conn] = ctx
			go func() {
				for {
					select {
					case <-ctx.Done():
						b.lock.Lock()
						delete(b.Clients, conn)
						b.lock.Unlock()
					}
				}
			}()
			return ctx
		},
	}
	return b
}

func (b *Broker) ListenAndServe() error {
	return b.server.ListenAndServe()
}

func (b *Broker) Serve(l net.Listener) error {
	if b.Addr != l.Addr().String() {
		b.Addr = l.Addr().String()
	}
	return b.server.Serve(l)
}

func (b *Broker) Close() {
	for _, balancer := range b.balancers {
		balancer.stop <- true
	}
	b.server.Close()
}

func (b *Broker) ServeMessage(rw protocol.ResponseWriter, req *protocol.Request) {
	client := getClientContext(req)
	client.heartbeat = time.Now()

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
	case *createTopics.Request:
		err = b.createtopics(rw, req)
	default:
		err = fmt.Errorf("unsupported api key: %v", req.Header.ApiKey)
	}

	if err != nil && err.Error() != "use of closed network connection" {
		panic(fmt.Sprintf("kafka broker: %v", err))
	}
}

func (b *Broker) getBalancer(group *store.Group) *groupBalancerNew {
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
	return req.Context.Value("client").(*ClientContext)
}
