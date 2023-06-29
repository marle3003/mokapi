package imap

import (
	"encoding/base64"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"mokapi/sasl"
	"net"
	"net/textproto"
	"strings"
	"syscall"
)

type ConnState uint8

const (
	Init ConnState = iota
	NotAuthenticated
	Authenticated
	Selected
	Logout
)

type conn struct {
	tpc   *textproto.Conn
	state ConnState
}

func (c *conn) serve() {
	c.tpc.PrintfLine("* OK [CAPABILITY IMAP4rev1 STARTTLS AUTH=PLAIN] Mokapi Ready")

	for {
		err := c.readCmd()
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
	default:
		log.Errorf("imap: unknown command: %v", line)
		res = &response{
			status: bad,
			text:   "Unknown command",
		}
	}
	return c.writeResponse(tag, res)
}

func (c *conn) writeCapabilities(tag string) {

}

func (c *conn) handleAuth(tag, param string) *response {
	params := strings.SplitN(param, " ", 2)
	mechanism := params[0]
	var resp []byte
	var err error
	if len(params) > 1 {
		resp, err = base64.StdEncoding.DecodeString(params[1])
		if err != nil {
			return &response{}
		}
	}
	var saslServer sasl.Server
	mechanism = strings.ToUpper(mechanism)
	switch mechanism {
	case "PLAIN":
		saslServer = sasl.NewPlainServer(func(identity, username, password string) error {
			return nil
		})
	default:
		return &response{
			status: no,
			text:   "Unsupported authentication mechanism",
		}
	}

	for {
		if err != nil {
			return &response{
				status: bad,
				text:   err.Error(),
			}
		}

		challenge, err := saslServer.Next(resp)
		if err != nil {
			return &response{
				status: bad,
				text:   err.Error(),
			}
		}

		if !saslServer.HasNext() {
			break
		}

		err = c.tpc.PrintfLine(fmt.Sprintf("+ %s", challenge))
		if err != nil {
			return &response{
				status: bad,
				text:   err.Error(),
			}
		}

		line, err := c.tpc.ReadLine()
		if err != nil {
			return &response{
				status: bad,
				text:   err.Error(),
			}
		}
		resp, err = base64.StdEncoding.DecodeString(line)
		if err != nil {
			return &response{
				status: bad,
				text:   "Invalid response",
			}
		}
	}

	return &response{
		status: ok,
		text:   "Authenticated",
	}
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
