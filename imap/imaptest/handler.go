package imaptest

import (
	"context"
	"mokapi/imap"
)

type Handler struct {
	session      map[string]interface{}
	SelectFunc   func(mailbox string, session map[string]interface{}) (*imap.Selected, error)
	UnselectFunc func(session map[string]interface{}) error
}

func (h *Handler) Select(mailbox string, _ context.Context) (*imap.Selected, error) {
	if h.SelectFunc != nil {
		h.ensureSession()
		return h.SelectFunc(mailbox, h.session)
	}
	panic("select not implemented")
}

func (h *Handler) Unselect(_ context.Context) error {
	if h.UnselectFunc != nil {
		h.ensureSession()
		return h.UnselectFunc(h.session)
	}
	panic("unselect not implemented")
}

func (h *Handler) ensureSession() {
	if h.session == nil {
		h.session = map[string]interface{}{}
	}
}
