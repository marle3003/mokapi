package imap

import (
	"context"
	"crypto/tls"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"net/textproto"
	"runtime/debug"
	"strings"
	"syscall"
)

type conn struct {
	conn      net.Conn
	ctx       context.Context
	tpc       *textproto.Conn
	state     ConnState
	tlsConfig *tls.Config
	handler   Handler
}

func (c *conn) serve() {
	_, cancel := context.WithCancel(c.ctx)
	defer func() {
		r := recover()
		if r != nil {
			log.Debugf("smtp panic: %v", string(debug.Stack()))
			log.Errorf("smtp panic: %v", r)
		}
		cancel()
	}()

	if c.tpc == nil {
		c.tpc = textproto.NewConn(c.conn)
	}
	err := c.writeStatusCapability("*", ok, "Mokapi Ready")
	if err != nil {
		log.Errorf("failed to send greetings: %v", err)
		return
	}

	for {
		err = c.readCmd()
		if err != nil {
			switch {
			case clientDisconnected(err):
			default:
				log.Errorf("imap: %v", err)
			}
			return
		}
	}
}

func (c *conn) readCmd() error {
	line, err := c.tpc.ReadLine()
	if err != nil {
		return err
	}

	tag, cmd, param := parseLine(line)
	var res *response
	switch cmd {
	case "AUTHENTICATE":
		res = c.handleAuth(tag, param)
	case "CAPABILITY":
		res = c.handleCapability()
	case "STARTTLS":
		err = c.handleStartTLS(tag)
	case "SELECT":
		err = c.handleSelect(tag, param)
	default:
		log.Errorf("imap: unknown command: %v", line)
		res = &response{
			status: bad,
			text:   "Unknown command",
		}
	}
	if err != nil {
		return err
	}
	if res != nil {
		return c.writeResponse(tag, res)
	}
	return nil
}

func (c *conn) writeResponse(tag string, res *response) error {
	return c.tpc.PrintfLine("%v %v", tag, res)
}

func parseLine(line string) (tag, cmd, param string) {
	a := strings.SplitN(line, " ", 2)
	if len(a) != 2 {
		if len(a) == 0 {
			return line, "", ""
		} else {
			return a[0], a[1], ""
		}
	}
	tag = a[0]
	a = strings.SplitN(a[1], " ", 2)
	cmd = strings.ToUpper(a[0])
	if len(a) == 2 {
		param = strings.TrimSpace(a[1])
	}
	return
}

func clientDisconnected(err error) bool {
	return err == io.EOF || errors.Is(err, net.ErrClosed) || errors.Is(err, syscall.ECONNRESET)
}
