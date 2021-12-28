package kafkatest

import (
	"mokapi/kafka"
	"mokapi/kafka/store"
	"net"
)

type Broker struct {
	Listener net.Listener

	broker *kafka.Broker
	client *Client
}

type BrokerOptions func(c *config)

type config struct {
	addr string
}

func NewBroker(opts ...BrokerOptions) *Broker {
	c := &config{addr: "127.0.0.1:"}
	for _, o := range opts {
		o(c)
	}
	l, err := net.Listen("tcp", c.addr)
	if err != nil {
		panic(err)
	}

	broker := kafka.NewBroker(0, l.Addr().String())
	broker.Store = &store.Store{}

	b := &Broker{
		Listener: l,
		broker:   broker,
		client:   NewClient(l.Addr().String(), "kafkatest"),
	}

	go b.broker.Serve(l)

	return b
}

func (b *Broker) Store() *store.Store {
	return b.broker.Store
}

func (b *Broker) SetStore(s *store.Store) {
	b.broker.Store = s
}

func (b *Broker) Client() *Client {
	return b.client
}

func (b *Broker) Close() {
	b.client.Close()
	b.broker.Close()
	b.Listener.Close()
}

func (b *Broker) HostPort() (string, int) {
	addr := b.Listener.Addr().(*net.TCPAddr)
	return addr.IP.String(), addr.Port
}

func WithPort(addr string) BrokerOptions {
	return func(c *config) {
		c.addr = addr
	}
}
