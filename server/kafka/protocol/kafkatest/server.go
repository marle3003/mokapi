package kafkatest

import (
	"mokapi/server/kafka/protocol"
	"net"
	"sync"
)

type Server struct {
	Listener net.Listener
	server   *protocol.Server
	wg       sync.WaitGroup
}

type handler struct {
	fn func(rw protocol.ResponseWriter, req *protocol.Request)
}

func NewServer(handlerFunc func(rw protocol.ResponseWriter, req *protocol.Request)) *Server {
	l, err := net.Listen("tcp", "")
	if err != nil {
		panic(err)
	}
	return &Server{Listener: l, server: &protocol.Server{Handler: &handler{fn: handlerFunc}}}
}

func (s *Server) Start() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.server.Serve(s.Listener)
	}()
}

func (s *Server) Close() {
	s.server.Close()
	s.Listener.Close()
	s.wg.Wait()
}

func (h *handler) ServeMessage(rw protocol.ResponseWriter, req *protocol.Request) {
	h.fn(rw, req)
}
