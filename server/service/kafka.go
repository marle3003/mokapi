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
	c, ok := b.clusters[r.Host]
	if ok {
		c.ServeMessage(rw, r)
		return
	}
	_, port, _ := net.SplitHostPort(r.Host)
	c, ok = b.clusters[fmt.Sprintf(":%v", port)]
	if ok {
		c.ServeMessage(rw, r)
		return
	}

	if _, ok := r.Message.(*apiVersion.Request); ok {
		log.Errorf("received kafka message for unknown host: %v", r.Host)
		rw.Write(&apiVersion.Response{ErrorCode: kafka.UnknownServerError})
	} else {
		panic(fmt.Sprintf("received kafka message for unknown host: %v", r.Host))
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
