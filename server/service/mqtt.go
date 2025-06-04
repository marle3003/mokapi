package service

import (
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"mokapi/mqtt"
)

type MqttBroker struct {
	server *mqtt.Server
}

func NewMqttBroker(port string, handler mqtt.Handler) *MqttBroker {
	b := &MqttBroker{
		server: &mqtt.Server{Addr: fmt.Sprintf(":%v", port), Handler: handler},
	}
	return b
}

func (b *MqttBroker) Addr() string {
	return b.server.Addr
}

func (b *MqttBroker) Start() {
	go func() {
		err := b.server.ListenAndServe()
		if !errors.Is(err, mqtt.ErrServerClosed) {
			log.Error(err)
		}
	}()
}

func (b *MqttBroker) Stop() {
	b.server.Close()
}
