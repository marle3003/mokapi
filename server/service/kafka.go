package service

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/kafka"
	"mokapi/kafka/apiVersion"
	"net"
	"sync"
)

type KafkaBroker struct {
	server   *kafka.Server
	clusters map[string]kafka.Handler
	m        sync.RWMutex
}

func NewKafkaBroker(port string) *KafkaBroker {
	b := &KafkaBroker{
		server:   &kafka.Server{Addr: fmt.Sprintf(":%v", port)},
		clusters: make(map[string]kafka.Handler),
	}
	b.server.Handler = b
	return b
}

func (b *KafkaBroker) Add(addr string, handler kafka.Handler) {
	b.m.Lock()
	defer b.m.Unlock()

	if _, ok := b.clusters[addr]; !ok {
		log.Infof("add new kafka broker on %v", addr)
	}
	b.clusters[addr] = handler
}

func (b *KafkaBroker) Remove(host string) {
	b.m.Lock()
	defer b.m.Unlock()

	delete(b.clusters, host)
	if len(b.clusters) == 0 {
		b.Stop()
	}
}

func (b *KafkaBroker) ServeMessage(rw kafka.ResponseWriter, r *kafka.Request) {
	h, err := b.getHandler(r.Host)
	if err == nil {
		h.ServeMessage(rw, r)
	} else {
		if _, ok := r.Message.(*apiVersion.Request); ok {
			log.Errorf("received kafka message for unknown host: %v", r.Host)
			rw.Write(&apiVersion.Response{ErrorCode: kafka.UnknownServerError})
		} else {
			panic(fmt.Sprintf("received kafka message for unknown host: %v", r.Host))
		}
	}
}

func (b *KafkaBroker) Start() {
	go func() {
		err := b.server.ListenAndServe()
		if err != kafka.ErrServerClosed {
			log.Error(err)
		}
	}()
}

func (b *KafkaBroker) Stop() {
	b.server.Close()
}

func (b *KafkaBroker) getHandler(hostport string) (kafka.Handler, error) {
	if h, ok := b.clusters[hostport]; ok {
		return h, nil
	}

	host, port, err := net.SplitHostPort(hostport)
	if err != nil {
		return nil, err
	}

	if h, ok := b.clusters[fmt.Sprintf(":%v", port)]; ok {
		return h, nil
	}

	if isLocalhost(host) {
		for _, alias := range localhostAliases {
			if h, ok := b.clusters[fmt.Sprintf("%v:%v", alias, port)]; ok {
				return h, nil
			}
		}
	}

	return nil, fmt.Errorf("received kafka message for unknown host: %v", hostport)

}

func isLocalhost(host string) bool {
	for _, h := range localhostAliases {
		if host == h {
			return true
		}
	}
	return false
}

var localhostAliases = []string{"::1", "127.0.0.1", "localhost"}
