package smtp

import (
	"context"
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
			write(tpc, StatusOk, Undefined, "Hello "+param, "AUTH LOGIN")
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
	msg := &Message{}
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
	r := &LoginRequest{ctx: c.ctx}
	var err error

	parts := strings.Fields(param)
	if len(parts) == 1 {
		err = write(conn, StatusAuthMethodAccepted, Undefined, "VXNlcm5hbWU6") // base64(Username:)
		if err != nil {
			return err
		}
		r.Username, err = conn.ReadLine()
		if err != nil {
			return err
		}
	} else {
		r.Username = parts[1]
	}

	err = write(conn, StatusAuthMethodAccepted, Undefined, "UGFzc3dvcmQ6") // base64(Password:)
	if err != nil {
		return err
	}

	r.Password, err = conn.ReadLine()
	if err != nil {
		return err
	}

	c.server.Handler.ServeSMTP(&response{
		conn: conn,
	}, r)

	return nil
}

func (c *conn) reset() {
	ctx := ClientFromContext(c.ctx)
	ctx.Reset()
}

func clientDisconnected(err error) bool {
	return err == io.EOF || errors.Is(err, net.ErrClosed) || errors.Is(err, syscall.ECONNRESET)
}
