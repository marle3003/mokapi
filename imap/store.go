package imap

import (
	"fmt"
	"strings"
)

type StoreAction string

type StoreRequest struct {
	Sequence IdSet
	Action   string
	Flags    []Flag
	Silent   bool
}

func (c *conn) handleStore(tag, param string) error {
	if c.state != AuthenticatedState && c.state != SelectedState {
		return c.writeResponse(tag, &response{
			status: bad,
			text:   "Command is only valid in authenticated state",
		})
	}

	d := Decoder{msg: param}
	req, err := parseStoreRequest(&d)

	res := fetchResponse{}
	if err = c.handler.Store(&req, &res, c.ctx); err != nil {
		return err
	}

	if err = c.writeFetchResponse(&res); err != nil {
		return err
	}

	return c.writeResponse(tag, &response{
		status: ok,
		text:   "STORE completed",
	})
}

func parseStoreRequest(d *Decoder) (StoreRequest, error) {
	req := StoreRequest{}
	var err error

	req.Sequence, err = d.Sequence()
	if err != nil {
		return req, err
	}

	var action string
	if action, err = d.SP().String(); err != nil {
		return req, err
	}

	action = strings.ToUpper(action)

	if strings.HasSuffix(action, ".SILENT") {
		req.Silent = true
		action = strings.TrimSuffix(action, ".SILENT")
	}

	switch action {
	case "+FLAGS":
		req.Action = "add"
	case "-FLAGS":
		req.Action = "remove"
	case "FLAGS":
		req.Action = "replace"
	default:
		return req, fmt.Errorf("store action '%s' not supported", action)
	}

	err = d.SP().List(func() error {
		var flag string
		if flag, err = d.ReadFlag(); err != nil {
			return err
		}
		req.Flags = append(req.Flags, Flag(flag))
		return nil
	})

	return req, err
}
