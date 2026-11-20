package server

import (
	"context"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	engine "mokapi/engine/common"
	"mokapi/ldap"
	"mokapi/providers/directory"
	"mokapi/runtime"
	"mokapi/server/cert"
	"net"
	"sync"
	"sync/atomic"
)

type LdapDirectoryManager struct {
	servers map[string]*ldapHost

	eventEmitter engine.EventEmitter
	certStore    *cert.Store
	app          *runtime.App
	m            sync.Mutex
}

// ldapHost keeps the listener next to the server because ldap.Server only learns about it once
// Serve runs, so Close alone can leave the port open. A reload swaps the handler of the running
// server, because ldap.Server reads its Handler for every request.
type ldapHost struct {
	*ldap.Server
	handler *swappableLdapHandler
	ln      net.Listener
}

func (h *ldapHost) close() {
	h.Close()
	if h.ln != nil {
		_ = h.ln.Close()
	}
}

// swappableLdapHandler delegates to the handler of the current config snapshot.
type swappableLdapHandler struct {
	current atomic.Pointer[ldap.Handler]
}

func newSwappableLdapHandler(h ldap.Handler) *swappableLdapHandler {
	s := &swappableLdapHandler{}
	s.swap(h)
	return s
}

func (s *swappableLdapHandler) swap(h ldap.Handler) {
	s.current.Store(&h)
}

func (s *swappableLdapHandler) ServeLDAP(rw ldap.ResponseWriter, r *ldap.Request) {
	(*s.current.Load()).ServeLDAP(rw, r)
}

func (s *swappableLdapHandler) Unbind(ctx context.Context) {
	(*s.current.Load()).Unbind(ctx)
}

func NewLdapDirectoryManager(emitter engine.EventEmitter, store *cert.Store, app *runtime.App) *LdapDirectoryManager {
	return &LdapDirectoryManager{
		eventEmitter: emitter,
		certStore:    store,
		app:          app,
		servers:      map[string]*ldapHost{},
	}
}

func (m *LdapDirectoryManager) UpdateConfig(e dynamic.ConfigEvent) {
	cfg, ok := runtime.IsLdapConfig(e.Config)
	if !ok {
		return
	}

	m.m.Lock()
	defer m.m.Unlock()

	name := cfg.Info.Name
	info := m.app.Ldap.Get(name)
	var old *directory.Config
	if info != nil {
		old = info.Config()
	}

	if e.Event == dynamic.Delete {
		if info == nil {
			return
		}
		m.app.Ldap.Remove(e.Config)
	} else {
		info = m.app.Ldap.Add(e.Config, m.eventEmitter)
	}

	current := info.Config()
	if current == nil {
		m.removeHost(name)
		return
	}

	host, running := m.servers[name]
	switch {
	case !running:
		m.start(info)
	case old != nil && old.Address != current.Address:
		m.removeHost(name)
		m.start(info)
	default:
		host.handler.swap(info.Handler(m.app.Monitor.Ldap, m.app.Events))
	}
}

// removeHost expects the caller to hold m.m.
func (m *LdapDirectoryManager) removeHost(name string) {
	h, ok := m.servers[name]
	if !ok {
		return
	}
	log.Infof("removing LDAP host '%v' on binding %v", name, h.Addr)
	h.close()
	delete(m.servers, name)
}

func (m *LdapDirectoryManager) start(info *runtime.LdapInfo) {
	cfg := info.Config()
	handler := newSwappableLdapHandler(info.Handler(m.app.Monitor.Ldap, m.app.Events))
	s := &ldapHost{Server: &ldap.Server{Addr: cfg.Address, Handler: handler}, handler: handler}
	m.servers[cfg.Info.Name] = s
	log.Infof("adding LDAP host '%v' on binding %v", cfg.Info.Name, s.Addr)

	logFailure := func(err error) {
		log.Errorf("adding LDAP host '%v' on binding %v failed: %v", cfg.Info.Name, s.Addr, err)
	}
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		logFailure(err)
		return
	}
	s.ln = ln
	go func() {
		if err := s.Serve(ln); !errors.Is(err, ldap.ErrServerClosed) {
			logFailure(err)
		}
	}()
}

func (m *LdapDirectoryManager) Stop() {
	m.m.Lock()
	defer m.m.Unlock()

	for _, s := range m.servers {
		s.close()
	}
}
