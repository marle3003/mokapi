package smtptest

import (
	"golang.org/x/net/nettest"
	"mokapi/smtp"
	"net"
	"time"
)

func NewServer(h smtp.HandlerFunc) (*smtp.Server, net.Conn, error) {
	l, err := nettest.NewLocalListener("tcp")
	if err != nil {
		return nil, nil, err
	}

	server := &smtp.Server{Handler: h, Addr: l.Addr().String()}
	go server.Serve(l)

	backoff := 50 * time.Millisecond
	var conn net.Conn
	for i := 0; i < 10; i++ {
		d := net.Dialer{Timeout: time.Second * 10}
		conn, err = d.Dial(l.Addr().Network(), l.Addr().String())
		if err != nil {
			time.Sleep(backoff)
			continue
		}
	}
	if err != nil {
		server.Close()
		return nil, nil, err
	}

	return server, conn, nil
}
