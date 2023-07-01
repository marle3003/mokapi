package imaptest

import "mokapi/imap"

type Handler struct {
	SelectFunc func(mailbox string) (*imap.Selected, error)
}

func (h *Handler) Select(mailbox string) (*imap.Selected, error) {
	if h.SelectFunc != nil {
		return h.SelectFunc(mailbox)
	}
	panic("not implemented")
}
