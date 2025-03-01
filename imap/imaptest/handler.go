package imaptest

import (
	"context"
	"mokapi/imap"
)

type Handler struct {
	session      map[string]interface{}
	LoginFunc    func(username, password string, session map[string]interface{}) error
	SelectFunc   func(mailbox string, session map[string]interface{}) (*imap.Selected, error)
	UnselectFunc func(session map[string]interface{}) error
	ListFunc     func(ref, pattern string, flags []imap.MailboxFlags, session map[string]interface{}) ([]imap.ListEntry, error)
	FetchFunc    func(request *imap.FetchRequest, response imap.FetchResponse, session map[string]interface{}) error
}

func (h *Handler) Login(username, password string, _ context.Context) error {
	if h.LoginFunc != nil {
		h.ensureSession()
		return h.LoginFunc(username, password, h.session)
	}
	return nil
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

func (h *Handler) List(ref, pattern string, flags []imap.MailboxFlags, _ context.Context) ([]imap.ListEntry, error) {
	if h.ListFunc != nil {
		h.ensureSession()
		return h.ListFunc(ref, pattern, flags, h.session)
	}
	panic("list not implemented")
}

func (h *Handler) Fetch(request *imap.FetchRequest, response imap.FetchResponse, _ context.Context) error {
	if h.FetchFunc != nil {
		h.ensureSession()
		return h.FetchFunc(request, response, h.session)
	}
	panic("fetch not implemented")
}

func (h *Handler) ensureSession() {
	if h.session == nil {
		h.session = map[string]interface{}{}
	}
}
