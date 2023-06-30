package imap

import (
	"crypto/tls"
	"net/textproto"
)

func (c *conn) canStartTLS() bool {
	_, isTLS := c.conn.(*tls.Conn)
	return !isTLS && c.tlsConfig != nil && c.state == NotAuthenticated
}

func (c *conn) handleStartTLS(tag string) error {
	if !c.canStartTLS() {
		return c.writeResponse(tag, &response{
			status: bad,
			text:   "STARTTLS not available",
		})
	}
	err := c.writeResponse(tag, &response{
		status: ok,
		text:   "Begin TLS negotiation now",
	})
	if err != nil {
		return err
	}

	tlsConn := tls.Server(c.conn, c.tlsConfig)
	if err := tlsConn.Handshake(); err != nil {
		return c.writeResponse(tag, &response{
			status: bad,
			text:   "Handshake error: " + err.Error(),
		})
	}
	c.conn = tlsConn
	c.tpc = textproto.NewConn(tlsConn)
	return nil
}
