package kafkatest

import (
	"mokapi/kafka"
	"net"
	"sync"
)

type Server struct {
	Listener net.Listener
	server   *kafka.Server
	wg       sync.WaitGroup
}

type handler struct {
	fn func(rw kafka.ResponseWriter, req *kafka.Request)
}

func NewServer(handlerFunc func(rw kafka.ResponseWriter, req *kafka.Request)) *Server {
	l, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		panic(err)
	}
	return &Server{Listener: l, server: &kafka.Server{Handler: &handler{fn: handlerFunc}}}
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

func (h *handler) ServeMessage(rw kafka.ResponseWriter, req *kafka.Request) {
	h.fn(rw, req)
}
