package service

import (
	"errors"
	"fmt"
	"mokapi/mqtt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type WebsocketServer struct {
	server *http.Server
}

func NewWebsocketServer(port string, handler http.Handler) *WebsocketServer {
	b := &WebsocketServer{
		server: &http.Server{Addr: fmt.Sprintf(":%v", port), Handler: handler},
	}
	return b
}

func (b *WebsocketServer) Addr() string {
	return b.server.Addr
}

func (b *WebsocketServer) Start() {
	go func() {
		err := b.server.ListenAndServe()
		if !errors.Is(err, mqtt.ErrServerClosed) {
			log.Error(err)
		}
	}()
}

func (b *WebsocketServer) Stop() {
	_ = b.server.Close()
}
