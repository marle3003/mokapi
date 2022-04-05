package kafkatest

import (
	"fmt"
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/asyncApi/asyncapitest"
	"mokapi/kafka"
	"mokapi/try"
	"net"
	"strconv"
)

type Broker struct {
	Addr   string
	client *Client
	server *kafka.Server

	cfg *asyncApi.Config
}

type BrokerOptions func(c *config)

type config struct {
	addr    string
	handler kafka.Handler
}

func NewBroker(opts ...BrokerOptions) *Broker {
	c := &config{addr: "127.0.0.1:"}
	for _, o := range opts {
		o(c)
	}
	p, err := try.GetFreePort()
	if err != nil {
		panic(err)
	}
	addr := fmt.Sprintf("127.0.0.1:%v", p)

	cfg := asyncapitest.NewConfig(asyncapitest.WithServer("test", "kafka", addr))

	b := &Broker{
		Addr:   addr,
		server: &kafka.Server{Addr: addr, Handler: c.handler},
		client: NewClient(addr, "kafkatest"),
		cfg:    cfg,
	}

	go func() {
		err = b.server.ListenAndServe()
		if err != nil && err != kafka.ErrServerClosed {
			panic(err)
		}
	}()

	return b
}

func (b *Broker) Client() *Client {
	return b.client
}

func (b *Broker) Close() {
	b.client.Close()
	b.server.Close()
}

//func (b *Broker) AddRecords(topic string, records ...protocol.Record) {
//	b.client.Produce(3, &produce.Request{
//		Topics: []produce.RequestTopic{
//			{
//				Name: topic, Partitions: []produce.RequestPartition{
//					{
//						Index: 0,
//						Record: protocol.RecordBatch{
//							Records: records,
//						},
//					},
//				},
//			}},
//	})
//}

func (b *Broker) HostPort() (string, int) {
	h, ps, _ := net.SplitHostPort(b.Addr)
	p, _ := strconv.Atoi(ps)
	return h, p
}

func WithPort(addr string) BrokerOptions {
	return func(c *config) {
		c.addr = addr
	}
}

func WithHandler(h kafka.Handler) BrokerOptions {
	return func(c *config) {
		c.handler = h
	}
}
