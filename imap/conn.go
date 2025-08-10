package imap

import (
	"context"
	"crypto/tls"
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
	d := Decoder{msg: param}

	// missing feature
	// - ID
	// - namespace
	// - sort
	// - thread

	var res *response
	switch cmd {
	case "AUTHENTICATE":
		res = c.handleAuth(tag, param)
	case "LOGIN":
		err = c.handleLogin(tag, param)
	case "CAPABILITY":
		res = c.handleCapability()
	case "STARTTLS":
		err = c.handleStartTLS(tag)
	case "SELECT", "EXAMINE":
		err = c.handleSelect(tag, param, cmd == "EXAMINE")
	case "STATUS":
		err = c.handleStatus(tag, &d)
	case "LIST":
		err = c.handleList(tag, &d)
	case "LSUB":
		err = c.handleLSub(tag, &d)
	case "SUBSCRIBE":
		err = c.handleSubscribe(tag, &d)
	case "UNSUBSCRIBE":
		err = c.handleUnsubscribe(tag, &d)
	case "FETCH":
		err = c.handleFetch(tag, param)
	case "CLOSE", "UNSELECT":
		err = c.handleUnselect(tag, cmd == "CLOSE")
	case "EXPUNGE":
		err = c.handleExpunge(tag, &d, false)
	case "UID":
		err = c.handleUid(tag, param)
	case "LOGOUT":
		c.tpc.PrintfLine("BYE logout")
		return nil
	case "STORE":
		err = c.handleStore(tag, param)
	case "CREATE":
		err = c.handleCreate(tag, &d)
	case "DELETE":
		err = c.handleDelete(tag, &d)
	case "RENAME":
		err = c.handleRename(tag, &d)
	case "COPY":
		err = c.handleCopy(tag, &d, false)
	case "MOVE":
		err = c.handleMove(tag, &d, false)
	case "SEARCH":
		err = c.handleSearch(tag, &d, false)
	case "APPEND":
		err = c.handleAppend(tag, &d)
	case "NOOP":
		res = &response{
			status: ok,
		}
	case "IDLE":
		err = c.handleIdle(tag)
	default:
		log.Errorf("imap: unknown command: %v", line)
		res = &response{
			status: bad,
			text:   fmt.Sprintf("Unknown command %v", cmd),
		}
	}
	if err != nil {
		res = &response{
			status: bad,
			text:   fmt.Sprintf("error %v", err.Error()),
		}
		return fmt.Errorf("parse command failed '%v': %w", line, err)
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
