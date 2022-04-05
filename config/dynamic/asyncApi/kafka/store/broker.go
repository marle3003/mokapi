package store

import "fmt"

type Broker struct {
	Id   int
	Name string
	Host string
	Port int
}

func (b *Broker) Addr() string {
	return fmt.Sprintf("%v:%v", b.Host, b.Port)
}
