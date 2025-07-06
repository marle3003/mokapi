package imaptest

import (
	"context"
	"mokapi/imap"
)

type Handler struct {
	session         map[string]interface{}
	LoginFunc       func(username, password string, session map[string]interface{}) error
	SelectFunc      func(mailbox string, readonly bool, session map[string]interface{}) (*imap.Selected, error)
	UnselectFunc    func(session map[string]interface{}) error
	ListFunc        func(ref, pattern string, flags []imap.MailboxFlags, session map[string]interface{}) ([]imap.ListEntry, error)
	FetchFunc       func(request *imap.FetchRequest, response imap.FetchResponse, session map[string]interface{}) error
	StoreFunc       func(request *imap.StoreRequest, response imap.FetchResponse, session map[string]interface{}) error
	ExpungeFunc     func(set *imap.IdSet, w imap.ExpungeWriter, session map[string]interface{}) error
	CreateFunc      func(name string, opt *imap.CreateOptions, session map[string]interface{}) error
	DeleteFunc      func(mailbox string, session map[string]interface{}) error
	RenameFunc      func(existingName, newName string, session map[string]interface{}) error
	CopyFunc        func(set *imap.IdSet, dest string, w imap.CopyWriter, session map[string]interface{}) error
	MoveFunc        func(set *imap.IdSet, dest string, w imap.MoveWriter, session map[string]interface{}) error
	StatusFunc      func(req *imap.StatusRequest, session map[string]interface{}) (imap.StatusResult, error)
	SubscribeFunc   func(mailbox string, session map[string]interface{}) error
	UnsubscribeFunc func(mailbox string, session map[string]interface{}) error
}

func (h *Handler) Login(username, password string, _ context.Context) error {
	if h.LoginFunc != nil {
		h.ensureSession()
		return h.LoginFunc(username, password, h.session)
	}
	return nil
}

func (h *Handler) Select(mailbox string, readonly bool, _ context.Context) (*imap.Selected, error) {
	if h.SelectFunc != nil {
		h.ensureSession()
		return h.SelectFunc(mailbox, readonly, h.session)
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

func (h *Handler) Store(request *imap.StoreRequest, response imap.FetchResponse, _ context.Context) error {
	if h.StoreFunc != nil {
		h.ensureSession()
		return h.StoreFunc(request, response, h.session)
	}
	panic("STORE not implemented")
}

func (h *Handler) Expunge(set *imap.IdSet, w imap.ExpungeWriter, _ context.Context) error {
	if h.ExpungeFunc != nil {
		h.ensureSession()
		return h.ExpungeFunc(set, w, h.session)
	}
	panic("EXPUNGE not implemented")
}

func (h *Handler) Create(name string, opt *imap.CreateOptions, _ context.Context) error {
	if h.CreateFunc != nil {
		h.ensureSession()
		return h.CreateFunc(name, opt, h.session)
	}
	panic("CREATE not implemented")
}

func (h *Handler) Delete(name string, _ context.Context) error {
	if h.DeleteFunc != nil {
		h.ensureSession()
		return h.DeleteFunc(name, h.session)
	}
	panic("DELETE not implemented")
}

func (h *Handler) Rename(existingName, newName string, _ context.Context) error {
	if h.RenameFunc != nil {
		h.ensureSession()
		return h.RenameFunc(existingName, newName, h.session)
	}
	panic("DELETE not implemented")
}

func (h *Handler) Copy(set *imap.IdSet, dest string, w imap.CopyWriter, _ context.Context) error {
	if h.CopyFunc != nil {
		h.ensureSession()
		return h.CopyFunc(set, dest, w, h.session)
	}
	panic("COPY not implemented")
}

func (h *Handler) Move(set *imap.IdSet, dest string, w imap.MoveWriter, _ context.Context) error {
	if h.MoveFunc != nil {
		h.ensureSession()
		return h.MoveFunc(set, dest, w, h.session)
	}
	panic("MOVE not implemented")
}

func (h *Handler) Status(req *imap.StatusRequest, _ context.Context) (imap.StatusResult, error) {
	if h.StatusFunc != nil {
		h.ensureSession()
		return h.StatusFunc(req, h.session)
	}
	panic("STATUS not implemented")
}

func (h *Handler) Subscribe(mailbox string, _ context.Context) error {
	if h.SubscribeFunc != nil {
		h.ensureSession()
		return h.SubscribeFunc(mailbox, h.session)
	}
	panic("SUBSCRIBE not implemented")
}

func (h *Handler) Unsubscribe(mailbox string, _ context.Context) error {
	if h.UnsubscribeFunc != nil {
		h.ensureSession()
		return h.UnsubscribeFunc(mailbox, h.session)
	}
	panic("UNSUBSCRIBE not implemented")
}

func (h *Handler) ensureSession() {
	if h.session == nil {
		h.session = map[string]interface{}{}
	}
}
