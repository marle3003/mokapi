package store

import "fmt"

type Broker struct {
	id   int
	name string
	host string
	port int
}

func (b *Broker) Id() int {
	return b.id
}

func (b *Broker) Name() string {
	return b.name
}

func (b *Broker) Host() string {
	return b.host
}

func (b *Broker) Port() int {
	return b.port
}

func (b *Broker) Addr() string {
	return fmt.Sprintf("%v:%v", b.host, b.port)
}
