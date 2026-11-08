package service

import (
	"fmt"
	"mokapi/kafka"
	"net"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type KafkaBroker struct {
	server *kafka.Server
}

func NewKafkaBroker(port string, handler kafka.Handler) *KafkaBroker {
	b := &KafkaBroker{
		server: &kafka.Server{Addr: fmt.Sprintf(":%v", port), Handler: handler},
	}
	return b
}

func (b *KafkaBroker) Addr() string {
	return b.server.Addr
}

func (b *KafkaBroker) Start() error {
	ln, err := net.Listen("tcp", b.server.Addr)
	if err != nil {
		return err
	}

	go func() {
		err := b.server.Serve(ln)
		if !errors.Is(err, kafka.ErrServerClosed) {
			log.Error(err)
		}
	}()
	return nil
}

func (b *KafkaBroker) Stop() {
	b.server.Close()
}
