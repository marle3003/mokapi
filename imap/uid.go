package imap

import (
	"fmt"
	"strings"
)

func (c *conn) handleUid(tag, param string) error {
	d := Decoder{msg: param}

	cmd, err := d.String()
	if err != nil {
		return err
	}

	switch strings.ToUpper(cmd) {
	case "FETCH":
		return c.handleUidFetch(tag, d.SP())
	default:
		return fmt.Errorf("UID command %s is not supported", cmd)
	}
}

func (c *conn) handleUidFetch(tag string, d *Decoder) error {
	req, err := parseFetch(d)
	if err != nil {
		return err
	}
	req.Options.UID = true

	res := fetchResponse{}
	if err = c.handler.Fetch(req, &res, c.ctx); err != nil {
		return err
	}

	if err := c.writeFetchResponse(&res); err != nil {
		return err
	}

	return c.writeResponse(tag, &response{
		status: ok,
		text:   "FETCH completed",
	})
}
