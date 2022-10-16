package smtptest

import (
	"mokapi/smtp"
	"net"
	"sync"
)

type Server struct {
	Listener net.Listener
	server   *smtp.Server
	wg       sync.WaitGroup
}

func NewServer(handlerFunc smtp.HandlerFunc) *Server {
	l, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		panic(err)
	}
	return &Server{Listener: l, server: &smtp.Server{Handler: handlerFunc}}
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
