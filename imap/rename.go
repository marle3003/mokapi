package imap

func (c *conn) handleRename(tag string, d *Decoder) error {
	existingName, err := d.String()
	if err != nil {
		return err
	}
	newName, err := d.String()
	if err != nil {
		return err
	}

	err = c.handler.Rename(existingName, newName, c.ctx)
	if err != nil {
		return c.writeResponse(tag, &response{
			status: no,
			text:   err.Error(),
		})
	}

	return c.writeResponse(tag, &response{
		status: ok,
		text:   "RENAME completed",
	})
}
