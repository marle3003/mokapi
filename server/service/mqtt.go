package service

import (
	"fmt"
	"mokapi/mqtt"
	"net"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type MqttBroker struct {
	server *mqtt.Server
	ln     net.Listener
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

func (b *MqttBroker) Start() error {
	ln, err := net.Listen("tcp", b.server.Addr)
	if err != nil {
		return err
	}
	b.ln = ln

	go func() {
		err := b.server.Serve(ln)
		if !errors.Is(err, mqtt.ErrServerClosed) {
			log.Error(err)
		}
	}()
	return nil
}

func (b *MqttBroker) Stop() {
	b.server.Close()
	// Serve only learns about the listener once its goroutine runs, so Close alone can leave the port open.
	if b.ln != nil {
		_ = b.ln.Close()
	}
}
