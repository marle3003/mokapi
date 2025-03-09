package imap

import "strings"

type capability string

const (
	imap4rev1Cap capability = "IMAP4rev1"
	startTLSCap  capability = "STARTTLS"
	authPlainCap capability = "AUTH=PLAIN"
	saslIrCap    capability = "SASL-IR"
	logoutCap    capability = "LOGOUT"

	selectCap capability = "SELECT"
	listCap   capability = "LIST"
	lSubCap   capability = "LSUB"
	fetchCap  capability = "FETCH"
	closeCap  capability = "CLOSE"
	uidCap    capability = "UID"
)

type capabilities []capability

func (c capabilities) String() string {
	var sb strings.Builder
	for i, cap := range c {
		if i > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(string(cap))
	}
	return sb.String()
}

func (c *conn) writeStatusCapability(tag string, status responseStatus, message string) error {
	caps := c.getCapabilities()
	return c.tpc.PrintfLine("%s %s [CAPABILITY %s] %s", tag, status, caps, message)
}

func (c *conn) handleCapability() *response {
	caps := c.getCapabilities()
	err := c.tpc.PrintfLine("* CAPABILITY %s", caps)
	if err != nil {
		return &response{
			status: bad,
			text:   err.Error(),
		}
	}
	return &response{
		status: ok,
		text:   "CAPABILITY completed",
	}
}

func (c *conn) getCapabilities() capabilities {
	caps := capabilities{imap4rev1Cap, saslIrCap}
	if c.canStartTLS() {
		caps = append(caps, startTLSCap)
	}
	if c.canAuth() {
		caps = append(caps, authPlainCap)
	}

	if c.state == AuthenticatedState || c.state == SelectedState {
		caps = append(caps, selectCap, listCap, fetchCap, closeCap, logoutCap, lSubCap, uidCap)
	}

	return caps
}
