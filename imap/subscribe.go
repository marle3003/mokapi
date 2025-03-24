package imap

func (c *conn) handleSubscribe(tag string, d *Decoder) error {
	mailbox, err := d.String()
	if err != nil {
		return err
	}

	err = c.handler.Subscribe(mailbox, c.ctx)
	if err != nil {
		return err
	}

	return c.writeResponse(tag, &response{
		status: ok,
		text:   "SUBSCRIBE completed",
	})
}

func (c *conn) handleUnsubscribe(tag string, d *Decoder) error {
	mailbox, err := d.String()
	if err != nil {
		return err
	}

	err = c.handler.Unsubscribe(mailbox, c.ctx)
	if err != nil {
		return err
	}

	return c.writeResponse(tag, &response{
		status: ok,
		text:   "UNSUBSCRIBE completed",
	})
}
