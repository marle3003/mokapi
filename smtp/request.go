package smtp

import (
	"context"
	"fmt"
	"net"
)

var crnl = []byte{'\r', '\n'}

type Request struct {
	Context context.Context
	Proto   string // SMTP or ESMTP
	Cmd     Command
	Param   string
	Message *MailMessage
}

func (r *Request) Write(conn net.Conn) error {
	switch r.Cmd {
	case Hello:
		return r.writeHello(conn)
	}

	return fmt.Errorf("unknown command %v", r.Cmd)
}

func (r *Request) writeHello(conn net.Conn) error {
	msg := fmt.Sprintf("EHLO %v", r.Param)
	_, err := conn.Write([]byte(msg))
	conn.Write(crnl)
	return err
}
