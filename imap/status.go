package imap

import (
	"fmt"
	"strings"
)

type StatusRequest struct {
	Mailbox     string
	Messages    bool
	Recent      bool
	UIDNext     bool
	UIDValidity bool
	Unseen      bool
}

type StatusResult struct {
	Messages    uint32
	Recent      uint32
	UIDNext     uint32
	UIDValidity uint32
	Unseen      uint32
}

func (c *conn) handleStatus(tag string, d *Decoder) error {
	mailbox, err := d.String()
	if err != nil {
		return err
	}

	req := &StatusRequest{
		Mailbox: mailbox,
	}
	err = d.SP().List(func() error {
		var opt string
		opt, err = d.String()
		if err != nil {
			return err
		}
		switch strings.ToUpper(opt) {
		case "MESSAGES":
			req.Messages = true
		case "RECENT":
			req.Recent = true
		case "UIDNEXT":
			req.UIDNext = true
		case "UIDVALIDITY":
			req.UIDValidity = true
		case "UNSEEN":
			req.Unseen = true
		default:
			return fmt.Errorf("unknown status option: %s", opt)
		}
		return nil
	})

	var res StatusResult
	res, err = c.handler.Status(req, c.ctx)
	if err != nil {
		return err
	}

	e := Encoder{}
	e.Atom("STATUS")
	e.SP().Atom(req.Mailbox)
	e.SP().BeginList()
	if req.Messages {
		e.ListItem("MESSAGES").SP().Number(res.Messages)
	}
	if req.Recent {
		e.ListItem("RECENT").SP().Number(res.Recent)
	}
	if req.UIDNext {
		e.ListItem("UIDNEXT").SP().Number(res.UIDNext)
	}
	if req.UIDValidity {
		e.ListItem("UIDVALIDITY").SP().Number(res.UIDValidity)
	}
	if req.Unseen {
		e.ListItem("UNSEEN").SP().Number(res.Unseen)
	}
	e.EndList()

	err = c.writeResponse(untagged, &response{
		text: e.String(),
	})
	if err != nil {
		return err
	}

	return c.writeResponse(tag, &response{
		status: ok,
		text:   "STATUS completed",
	})
}
