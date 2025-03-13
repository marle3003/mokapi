package imap

type CreateOptions struct {
	Flags []MailboxFlags
}

func (c *conn) handleCreate(tag string, d *Decoder) error {
	if c.state != AuthenticatedState && c.state != SelectedState {
		return c.writeResponse(tag, &response{
			status: bad,
			text:   "Command is only valid in authenticated state",
		})
	}

	name, err := d.String()
	if err != nil {
		return err
	}

	opt := CreateOptions{}
	if d.IsSP() {
		_ = d.ExpectSP()
		if err = d.expect("("); err != nil {
			return err
		}
		if err = d.expect("USE"); err != nil {
			return err
		}
		err = d.SP().List(func() error {
			var flag string
			flag, err = d.ReadFlag()
			opt.Flags = append(opt.Flags, MailboxFlags(flag))
			return err
		})
	}

	err = c.handler.Create(name, &opt, c.ctx)
	if err != nil {
		return err
	}

	return c.writeResponse(tag, &response{
		status: ok,
		text:   "CREATE completed",
	})
}
