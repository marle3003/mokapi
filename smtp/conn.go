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

	tpc := textproto.NewConn(c.conn)
	client := ClientFromContext(ctx)

	tpc.PrintfLine("220 localhost ESMTP Service Ready")

	for {
		line, err := tpc.ReadLine()
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

			exts := []string{"AUTH LOGIN"}
			if c.server.TLSConfig != nil {
				exts = append(exts, "STARTTLS")
			}
			args := []string{"Hello " + param}
			args = append(args, exts...)
			write(tpc, StatusOk, Undefined, args...)
		case "AUTH":
			c.serveAuth(tpc, param)
		case "MAIL":
			c.reset()
			c.serveMail(tpc, param)
		case "RCPT":
			c.serveRcpt(tpc, param)
		case "DATA":
			c.serveData(tpc, param)
			c.reset()
		case "NOOP":
			write(tpc, StatusOk, Success, "OK")
		case "QUIT":
			write(tpc, StatusClose, Success, "Bye, see you soon")
			c.server.closeConn(c.conn)
			return
		case "STARTTLS":
			c.serveStartTls(tpc, param)
			tpc = textproto.NewConn(c.conn)
		default:
			log.Debugf("unknown smtp command: %v", cmd)
			write(tpc, StatusCommandNotImplemented, Undefined, fmt.Sprintf("Command %v not implemented", cmd))
		}
	}
}

func (c *conn) serveMail(conn *textproto.Conn, param string) error {
	if len(param) < 6 || !strings.HasSuffix(strings.ToUpper(param[:5]), "FROM:") {
		return write(conn, StatusSyntaxError, SyntaxError, "expected from address")
	}

	args := strings.Split(param[5:], " ")
	from := args[0]
	if !strings.HasPrefix(from, "<") || !strings.HasSuffix(from, ">") {
		return write(conn, 501, SyntaxError, "Expected MAIL syntax of FROM:<address>")
	}
	from = strings.Trim(from, "<>")

	c.server.Handler.ServeSMTP(&response{conn: conn}, NewMailRequest(from, c.ctx))
	return nil
}

func (c *conn) serveRcpt(conn *textproto.Conn, param string) error {
	ctx := ClientFromContext(c.ctx)
	if len(ctx.From) == 0 {
		return write(conn, BadSequenceOfCommands, InvalidCommand, "Missing MAIL command.")
	}

	if len(param) < 3 || !strings.HasSuffix(strings.ToUpper(param[:3]), "TO:") {
		return write(conn, StatusSyntaxError, SyntaxError, "expected to address")
	}

	args := strings.Split(param[3:], " ")
	to := args[0]
	if !strings.HasPrefix(to, "<") || !strings.HasSuffix(to, ">") {
		return write(conn, 501, SyntaxError, "Expected RCPT syntax of TO:<address>")
	}
	to = strings.Trim(to, "<>")

	c.server.Handler.ServeSMTP(&response{conn: conn}, NewRcptRequest(to, c.ctx))
	return nil
}

func (c *conn) serveData(conn *textproto.Conn, param string) error {
	ctx := ClientFromContext(c.ctx)
	if ctx.From == "" || len(ctx.To) == 0 {
		return write(conn, BadSequenceOfCommands, InvalidCommand, "Missing MAIL/RCPT command.")
	}

	err := write(conn, StatusStartMailInput, Success, "Send message, ending in CRLF.CRLF")
	if err != nil {
		return err
	}
	msg := &Message{
		Server: c.conn.LocalAddr().String(),
	}
	err = msg.readFrom(conn.Reader)
	if clientDisconnected(err) {
		return err
	} else if err != nil {
		log.Infof("smtp: %v", err)
		return write(conn, StatusSyntaxError, SyntaxError, err.Error())
	}
	c.server.Handler.ServeSMTP(&response{conn: conn}, NewDataRequest(msg, c.ctx))
	return nil
}

func (c *conn) serveAuth(conn *textproto.Conn, param string) error {
	parts := strings.Fields(param)
	switch strings.ToUpper(parts[0]) {
	case "PLAIN":
		if len(parts) != 2 {
			return write(conn, 501, SyntaxError, "Expected plain message")
		}
		return c.servePlainAuth(conn, parts[1])
	case "LOGIN":
		msg := ""
		if len(parts) == 2 {
			msg = parts[1]
		}
		return c.serveLoginAuth(conn, msg)
	default:
		return write(conn, 504, SyntaxError, fmt.Sprintf("Command parameter %v is not supported", parts[0]))
	}
}

func (c *conn) servePlainAuth(conn *textproto.Conn, message string) error {
	b, err := base64.StdEncoding.DecodeString(message)
	if err != nil {
		return write(conn, 501, SyntaxError, "Expected plain credentials encoded base64")
	}
	data := strings.Split(string(b), "\x00")
	if len(data) != 3 {
		return write(conn, 501, SyntaxError, "invalid plain auth message format")
	}

	r := &LoginRequest{
		ctx:      c.ctx,
		Username: data[1],
		Password: data[2],
	}

	c.server.Handler.ServeSMTP(&response{
		conn: conn,
	}, r)
	return nil
}

func (c *conn) serveLoginAuth(conn *textproto.Conn, message string) error {
	var err error
	var username, password string

	if message == "" {
		err = write(conn, StatusAuthMethodAccepted, Undefined, "VXNlcm5hbWU6") // base64(Username:)
		if err != nil {
			return err
		}
		username, err = conn.ReadLine()
		if err != nil {
			return err
		}
	} else {
		username = message
	}

	err = write(conn, StatusAuthMethodAccepted, Undefined, "UGFzc3dvcmQ6") // base64(Password:)
	if err != nil {
		return err
	}

	password, err = conn.ReadLine()
	if err != nil {
		return err
	}

	r := &LoginRequest{ctx: c.ctx}
	var b []byte
	b, err = base64.StdEncoding.DecodeString(username)
	if err != nil {
		return write(conn, 501, SyntaxError, "Expected username encoded base64")
	} else {
		r.Username = string(b)
	}
	b, err = base64.StdEncoding.DecodeString(password)
	if err != nil {
		return write(conn, 501, SyntaxError, "Expected password encoded base64")
	} else {
		r.Password = string(b)
	}

	c.server.Handler.ServeSMTP(&response{
		conn: conn,
	}, r)

	return nil
}

func (c *conn) serveStartTls(conn *textproto.Conn, param string) {
	write(conn, 220, EnhancedStatusCode{2, 0, 0}, "Starting TLS")

	tlsConn := tls.Server(c.conn, c.server.TLSConfig)

	if err := tlsConn.Handshake(); err != nil {
		write(conn, 550, EnhancedStatusCode{5, 0, 0}, "Handshake error")
		return
	}
	c.conn = tlsConn
	c.reset()
}

func (c *conn) reset() {
	ctx := ClientFromContext(c.ctx)
	ctx.Reset()
}

func clientDisconnected(err error) bool {
	return err == io.EOF || errors.Is(err, net.ErrClosed) || errors.Is(err, syscall.ECONNRESET)
}
