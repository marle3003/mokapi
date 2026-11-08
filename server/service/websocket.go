package service

import (
	"errors"
	"fmt"
	"net"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type WebsocketServer struct {
	server *http.Server
	ln     net.Listener
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

func (b *WebsocketServer) Start() error {
	ln, err := net.Listen("tcp", b.server.Addr)
	if err != nil {
		return err
	}
	b.ln = ln

	go func() {
		err := b.server.Serve(ln)
		if !errors.Is(err, http.ErrServerClosed) {
			log.Error(err)
		}
	}()
	return nil
}

func (b *WebsocketServer) Stop() {
	_ = b.server.Close()
	// Serve only learns about the listener once its goroutine runs, so Close alone can leave the port open.
	if b.ln != nil {
		_ = b.ln.Close()
	}
}
