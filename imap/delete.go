package imap

func (c *conn) handleDelete(tag string, d *Decoder) error {
	mailbox, err := d.String()
	if err != nil {
		return err
	}

	err = c.handler.Delete(mailbox, c.ctx)
	if err != nil {
		return c.writeResponse(tag, &response{
			status: no,
			text:   err.Error(),
		})
	}

	return c.writeResponse(tag, &response{
		status: ok,
		text:   "DELETE completed",
	})
}
