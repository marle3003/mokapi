package imap

import "strings"

type capability string

const (
	imap4rev1 capability = "IMAP4rev1"
	startTLS  capability = "STARTTLS"
	auth      capability = "AUTH=PLAIN"
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
	caps := capabilities{imap4rev1}
	if c.canStartTLS() {
		caps = append(caps, startTLS)
	}
	if c.canAuth() {
		caps = append(caps, auth)
	}
	return caps
}
