package imap

import (
	"encoding/base64"
	"fmt"
	"mokapi/sasl"
	"strings"
)

func (c *conn) canAuth() bool {
	return c.state == NotAuthenticatedState
}

func (c *conn) handleAuth(_, param string) *response {
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

	c.state = AuthenticatedState

	return &response{
		status: ok,
		text:   "Authenticated",
	}
}
