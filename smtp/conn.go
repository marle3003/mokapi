package smtp

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
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
	server *Server
	conn   net.Conn
	ctx    context.Context
	tpc    *textproto.Conn
}

func (c *conn) serve() {
	ctx, cancel := context.WithCancel(c.ctx)
	defer func() {
		r := recover()
		if r != nil {
			log.Debugf("smtp panic: %v", string(debug.Stack()))
			log.Errorf("smtp panic: %v", r)
		}
		cancel()
		c.server.closeConn(c.conn)
	}()

	if c.tpc == nil {
		c.tpc = textproto.NewConn(c.conn)
	}
	client := ClientFromContext(ctx)

	c.tpc.PrintfLine("220 localhost ESMTP Service Ready")

	for {
		line, err := c.tpc.ReadLine()
		if err != nil {
			switch {
			case clientDisconnected(err):
				return
			default:
				log.Errorf("smtp: %v", err)
				return
			}
		}

		cmd, param := parseLine(line)

		switch cmd {
		case "EHLO":
			client.Client = param
			client.Proto = "ESMTP"

			exts := []string{"AUTH LOGIN PLAIN"}
			if c.server.TLSConfig != nil {
				if _, ok := c.conn.(*tls.Conn); !ok {
					exts = append(exts, "STARTTLS")
				}
			}
			args := []string{"Hello " + param}
			args = append(args, exts...)
			write(c.tpc, StatusOk, Undefined, args...)
		case "AUTH":
			c.serveAuth(param)
		case "MAIL":
			c.reset()
			c.serveMail(param)
		case "RCPT":
			c.serveRcpt(param)
		case "DATA":
			c.serveData(param)
			c.reset()
		case "NOOP":
			write(c.tpc, StatusOk, Success, "OK")
		case "QUIT":
			write(c.tpc, StatusClose, Success, "Bye, see you soon")
			c.server.closeConn(c.conn)
			return
		case "STARTTLS":
			c.serveStartTls(param)
		default:
			log.Debugf("unknown smtp command: %v", cmd)
			write(c.tpc, StatusCommandNotImplemented, Undefined, fmt.Sprintf("Command %v not implemented", cmd))
		}
	}
}

func (c *conn) ehlo() (map[string]string, error) {
	if c.tpc == nil {
		c.tpc = textproto.NewConn(c.conn)
	}
	id, err := c.tpc.Cmd("EHLO mokapi")
	if err != nil {
		return nil, err
	}
	c.tpc.StartResponse(id)
	defer c.tpc.EndResponse(id)
	_, msg, err := c.tpc.ReadResponse(250)
	if err != nil {
		return nil, err
	}

	ext := map[string]string{}
	split := strings.Split(msg, "\n")
	if len(split) > 1 {
		for _, line := range split[1:] {
			k, v, _ := strings.Cut(line, " ")
			ext[k] = v
		}
	}

	return ext, nil
}

func (c *conn) mail(from string) error {
	_, err := c.tpc.Cmd("MAIL FROM:<%s>", from)
	if err != nil {
		return err
	}
	_, _, err = c.tpc.ReadResponse(250)
	return err
}

func (c *conn) rcpt(to string) error {
	_, err := c.tpc.Cmd("RCPT TO:<%s>", to)
	if err != nil {
		return err
	}
	_, _, err = c.tpc.ReadResponse(250)
	return err
}

func (c *conn) data() (io.WriteCloser, error) {
	_, err := c.tpc.Cmd("DATA")
	if err != nil {
		return nil, err
	}
	_, _, err = c.tpc.ReadResponse(354)
	if err != nil {
		return nil, err
	}
	return c.tpc.DotWriter(), nil
}

func (c *conn) quit() error {
	_, err := c.tpc.Cmd("QUIT")
	if err != nil {
		return err
	}
	_, _, err = c.tpc.ReadResponse(221)
	return err
}

func (c *conn) serveMail(param string) error {
	if len(param) < 6 || !strings.HasSuffix(strings.ToUpper(param[:5]), "FROM:") {
		return write(c.tpc, StatusSyntaxError, SyntaxError, "expected from address")
	}

	args := strings.Split(param[5:], " ")
	from := args[0]
	if !strings.HasPrefix(from, "<") || !strings.HasSuffix(from, ">") {
		return write(c.tpc, 501, SyntaxError, "Expected MAIL syntax of FROM:<address>")
	}
	from = strings.Trim(from, "<>")

	c.server.Handler.ServeSMTP(&response{conn: c.tpc}, NewMailRequest(from, c.ctx))
	return nil
}

func (c *conn) serveRcpt(param string) error {
	ctx := ClientFromContext(c.ctx)
	if len(ctx.From) == 0 {
		return write(c.tpc, BadSequenceOfCommands, InvalidCommand, "Missing MAIL command.")
	}

	if len(param) < 3 || !strings.HasSuffix(strings.ToUpper(param[:3]), "TO:") {
		return write(c.tpc, StatusSyntaxError, SyntaxError, "expected to address")
	}

	args := strings.Split(param[3:], " ")
	to := args[0]
	if !strings.HasPrefix(to, "<") || !strings.HasSuffix(to, ">") {
		return write(c.tpc, 501, SyntaxError, "Expected RCPT syntax of TO:<address>")
	}
	to = strings.Trim(to, "<>")

	c.server.Handler.ServeSMTP(&response{conn: c.tpc}, NewRcptRequest(to, c.ctx))
	return nil
}

func (c *conn) serveData(param string) error {
	ctx := ClientFromContext(c.ctx)
	if ctx.From == "" {
		return write(c.tpc, BadSequenceOfCommands, InvalidCommand, "Missing MAIL command.")
	}
	if len(ctx.To) == 0 {
		return write(c.tpc, BadSequenceOfCommands, InvalidCommand, "Missing RCPT command.")
	}

	err := write(c.tpc, StatusStartMailInput, Success, "Send message, ending in CRLF.CRLF")
	if err != nil {
		return err
	}
	msg := &Message{
		Server: c.conn.LocalAddr().String(),
	}
	err = msg.readFrom(c.tpc.Reader)
	if clientDisconnected(err) {
		return err
	} else if err != nil {
		log.Infof("smtp: %v", err)
		return write(c.tpc, StatusSyntaxError, SyntaxError, err.Error())
	}
	c.server.Handler.ServeSMTP(&response{conn: c.tpc}, NewDataRequest(msg, c.ctx))
	return nil
}

func (c *conn) serveAuth(param string) error {
	parts := strings.Fields(param)
	switch strings.ToUpper(parts[0]) {
	case "PLAIN":
		if len(parts) != 2 {
			return write(c.tpc, 501, SyntaxError, "Expected plain message")
		}
		return c.servePlainAuth(parts[1])
	case "LOGIN":
		msg := ""
		if len(parts) == 2 {
			msg = parts[1]
		}
		return c.serveLoginAuth(msg)
	default:
		return write(c.tpc, 504, SyntaxError, fmt.Sprintf("Command parameter %v is not supported", parts[0]))
	}
}

func (c *conn) servePlainAuth(message string) error {
	b, err := base64.StdEncoding.DecodeString(message)
	if err != nil {
		return write(c.tpc, 501, SyntaxError, "Expected plain credentials encoded base64")
	}
	data := strings.Split(string(b), "\x00")
	if len(data) != 3 {
		return write(c.tpc, 501, SyntaxError, "invalid plain auth message format")
	}

	r := &LoginRequest{
		ctx:      c.ctx,
		Username: data[1],
		Password: data[2],
	}

	c.server.Handler.ServeSMTP(&response{
		conn: c.tpc,
	}, r)
	return nil
}

func (c *conn) serveLoginAuth(message string) error {
	var err error
	var username, password string

	if message == "" {
		err = write(c.tpc, StatusAuthMethodAccepted, Undefined, "VXNlcm5hbWU6") // base64(Username:)
		if err != nil {
			return err
		}
		username, err = c.tpc.ReadLine()
		if err != nil {
			return err
		}
	} else {
		username = message
	}

	err = write(c.tpc, StatusAuthMethodAccepted, Undefined, "UGFzc3dvcmQ6") // base64(Password:)
	if err != nil {
		return err
	}

	password, err = c.tpc.ReadLine()
	if err != nil {
		return err
	}

	r := &LoginRequest{ctx: c.ctx}
	var b []byte
	b, err = base64.StdEncoding.DecodeString(username)
	if err != nil {
		return write(c.tpc, 501, SyntaxError, "Expected username encoded base64")
	} else {
		r.Username = string(b)
	}
	b, err = base64.StdEncoding.DecodeString(password)
	if err != nil {
		return write(c.tpc, 501, SyntaxError, "Expected password encoded base64")
	} else {
		r.Password = string(b)
	}

	c.server.Handler.ServeSMTP(&response{
		conn: c.tpc,
	}, r)

	return nil
}

func (c *conn) startTls(config *tls.Config) error {
	_, err := c.tpc.Cmd("STARTTLS")
	if err != nil {
		return err
	}
	_, _, err = c.tpc.ReadResponse(220)
	if err != nil {
		return err
	}

	c.conn = tls.Client(c.conn, config)
	c.tpc = textproto.NewConn(c.conn)
	return nil
}

func (c *conn) serveStartTls(param string) {
	write(c.tpc, 220, EnhancedStatusCode{2, 0, 0}, "Starting TLS")

	tlsConn := tls.Server(c.conn, c.server.TLSConfig)

	if err := tlsConn.Handshake(); err != nil {
		write(c.tpc, 550, EnhancedStatusCode{5, 0, 0}, "Handshake error")
		return
	}
	c.conn = tlsConn
	c.tpc = textproto.NewConn(c.conn)
	c.reset()
}

func (c *conn) reset() {
	ctx := ClientFromContext(c.ctx)
	ctx.Reset()
}

func (c *conn) close() {
	c.conn.Close()
}

func clientDisconnected(err error) bool {
	return err == io.EOF || errors.Is(err, net.ErrClosed) || errors.Is(err, syscall.ECONNRESET)
}
