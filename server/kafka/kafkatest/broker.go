package kafkatest

import (
	"mokapi/server/kafka"
	"mokapi/server/kafka/memory"
	"net"
)

type Broker struct {
	Listener net.Listener

	broker *kafka.BrokerServer
	client *Client
}

func NewBroker() *Broker {
	l, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		panic(err)
	}

	broker := kafka.NewBrokerServer(0, l.Addr().String())
	broker.Cluster = memory.NewCluster(memory.Schema{})

	b := &Broker{
		Listener: l,
		broker:   broker,
		client:   NewClient(l.Addr().String(), "kafkatest"),
	}

	go b.broker.Serve(l)

	return b
}

func (b *Broker) Cluster() kafka.Cluster {
	return b.broker.Cluster
}

func (b *Broker) SetCluster(c kafka.Cluster) {
	b.broker.Cluster = c
}

func (b *Broker) Client() *Client {
	return b.client
}

func (b *Broker) Close() {
	b.client.Close()
	b.broker.Close()
	b.Listener.Close()
}
