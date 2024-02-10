package service

import (
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"mokapi/kafka"
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

func (b *KafkaBroker) Start() {
	go func() {
		err := b.server.ListenAndServe()
		if !errors.Is(err, kafka.ErrServerClosed) {
			log.Error(err)
		}
	}()
}

func (b *KafkaBroker) Stop() {
	b.server.Close()
}
