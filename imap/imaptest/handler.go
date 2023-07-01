package imaptest

import "mokapi/imap"

type Handler struct {
	SelectFunc func(mailbox string) *imap.Selected
}

func (h *Handler) Select(mailbox string) *imap.Selected {
	if h.SelectFunc != nil {
		return h.SelectFunc(mailbox)
	}
	panic("not implemented")
}
