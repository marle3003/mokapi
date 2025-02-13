package directory

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	engine "mokapi/engine/common"
	"mokapi/ldap"
	"mokapi/runtime/monitor"
)

type Directory struct {
	config  *Config
	emitter engine.EventEmitter
}

func NewHandler(config *Config, emitter engine.EventEmitter) ldap.Handler {
	return &Directory{config: config, emitter: emitter}
}

func (d *Directory) ServeLDAP(res ldap.ResponseWriter, r *ldap.Request) {
	switch r.Message.(type) {
	case *ldap.BindRequest:
		d.serveBind(res, r)
	case *ldap.SearchRequest:
		d.serveSearch(res, r)
	}

}

func (d *Directory) serveBind(rw ldap.ResponseWriter, r *ldap.Request) {
	msg := r.Message.(*ldap.BindRequest)
	// we do only support ldap v3
	if msg.Version != 3 {
		log.Debugf("unsupported LDAP version: %d", msg.Version)
		rw.Write(&ldap.BindResponse{
			Result:  ldap.ProtocolError,
			Message: "server supports only ldap version 3",
		})
		return
	}

	switch msg.Auth {
	case ldap.Simple:
		log.Debugf("received bind request with messageId %v, version %v. auth: %v", r.MessageId, msg.Version, msg.Name)
		if m, ok := monitor.LdapFromContext(r.Context); ok {
			m.Bind.WithLabel(d.config.Info.Name).Add(1)
		}
		rw.Write(&ldap.BindResponse{
			Result: ldap.Success,
		})
	default:
		rw.Write(&ldap.BindResponse{
			Result:  ldap.AuthMethodNotSupported,
			Message: "server supports only simple auth method",
		})
	}
}

func (d *Directory) skip(e *Entry, baseDN string) bool {
	if baseDN != "" && e.Dn == "" {
		return true
	}

	root, _ := d.config.Entries.Get("")
	name, ok := root.Attributes["subschemaSubentry"]
	if !ok || len(name) == 0 {
		return false
	}
	return e.Dn == name[0] && baseDN != fmt.Sprintf("cn=%s", name[0])
}
