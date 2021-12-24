package schema

import (
	"net"
	"net/url"
	"strconv"
)

type Cluster struct {
	Topics  []Topic
	Brokers []Broker
}

func New() Cluster {
	return Cluster{}
}

type Topic struct {
	Name       string
	Partitions []Partition
}

type Partition struct {
	Index    int
	Replicas []int
}

type Broker struct {
	Id   int
	Host string
	Port int
}

func NewBroker(id int, addr string) Broker {
	host, port := parseHostAndPort(addr)
	return Broker{Id: id, Host: host, Port: port}
}

func parseHostAndPort(s string) (host string, port int) {
	var err error
	var portString string
	host, portString, err = net.SplitHostPort(s)
	if err != nil {
		u, err := url.Parse(s)
		if err != nil || u.Host == "" {
			u, err = url.Parse("//" + s)
			if err != nil {
				return "", 9092
			}
		}

		host = u.Host
		portString = u.Port()
	}

	if len(portString) == 0 {
		port = 9092
	} else {
		var p int64
		p, err = strconv.ParseInt(portString, 10, 32)
		if err != nil {
			return
		}
		port = int(p)
	}

	return
}
