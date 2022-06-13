package ldap

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/ldap"
	"mokapi/runtime/monitor"
)

// Filters: https://ldapwiki.com/wiki/LDAP%20Filter%20Choices
// Search: https://ldapwiki.com/wiki/SearchRequest
// https://datatracker.ietf.org/doc/html/rfc4511

const (
	ApplicationBindRequest       = 0
	ApplicationBindResponse      = 1
	ApplicationUnbindRequest     = 2
	ApplicationSearchRequest     = 3
	ApplicationSearchResultEntry = 4
	ApplicationSearchResultDone  = 5
	ApplicationAbandonRequest    = 16

	FilterAnd            = 0
	FilterOr             = 1
	FilterNot            = 2
	FilterEqualityMatch  = 3
	FilterSubstrings     = 4
	FilterGreaterOrEqual = 5
	FilterLessOrEqual    = 6
	FilterPresent        = 7
	FilterApproxMatch    = 8

	FilterSubstringsStartWith = 0
	FilterSubstringsAny       = 1
	FilterSubstringsEndWith   = 2

	// ScopeBaseObject examines only the level specified by the base DN (and none of its child entries)
	ScopeBaseObject = 0
	// ScopeSingleLevel examines only the level immediately below the base DN
	ScopeSingleLevel = 1
	// ScopeWholeSubtree examines the subtree below the base DN and includes the base DN level
	ScopeWholeSubtree = 2

	ResultSuccess          int64 = 0
	ApplicationError       int64 = 1
	ProtocolError          int64 = 2
	AuthMethodNotSupported int64 = 7
)

type Directory struct {
	config  *ldap.Config
	server  *Server
	monitor *monitor.Ldap
}

func NewDirectory(config *ldap.Config, m *monitor.Ldap) *Directory {
	d := &Directory{
		config:  config,
		monitor: m,
		server: &Server{
			Addr: config.Address,
		},
	}
	d.server.Handler = d
	return d
}

func (d *Directory) Start() {
	go func() {
		log.Infof("starting ldap %v on %v", d.config.Info.Name, d.config.Address)
		err := d.server.ListenAndServe()
		if err != ErrServerClosed {
			log.Errorf("unable to start ldap server on %v: %v", d.server.Addr, err)
		}
	}()
}

func (d *Directory) Close() {
	d.server.Close()
}

func (d *Directory) Update(config *ldap.Config) {
	d.config = config
}

func (d *Directory) Serve(rw ResponseWriter, r *Request) {
	r.Context = monitor.NewLdapContext(r.Context, d.monitor)

	switch r.Body.Tag {
	case ApplicationBindRequest:
		if err := d.bind(rw, r); err != nil {
			log.Errorf("unable to process bind request: %v", err)
		}
	case ApplicationUnbindRequest:
		log.Debugf("received unbind request with messageId %v", r.MessageId)
		// just close connection
		return
	case ApplicationSearchRequest:
		if err := d.search(rw, r); err != nil {
			log.Errorf("unable to process search request with messageId %v: %v", r.MessageId, err)
		}
	case ApplicationAbandonRequest:
		log.Debugf("received abandon request with messageId %v", r.MessageId)
		// todo stop any searches on this messageid
		// The abandon operation does not have a response
	default:
		log.Errorf("unsupported tag %v", r.Body.Tag)
	}
}

func (d *Directory) getEntry(dn string) (ldap.Entry, error) {
	if entry, ok := d.config.Entries[dn]; ok {
		return entry, nil
	}
	return ldap.Entry{}, errors.Errorf("entry with dn %q not found", dn)
}
