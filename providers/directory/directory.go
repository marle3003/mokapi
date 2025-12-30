package directory

import (
	"context"
	"errors"
	engine "mokapi/engine/common"
	"mokapi/ldap"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"slices"
	"time"

	log "github.com/sirupsen/logrus"
)

type Directory struct {
	config  *Config
	emitter engine.EventEmitter
	eh      events.Handler
}

func NewHandler(config *Config, emitter engine.EventEmitter, eh events.Handler) ldap.Handler {
	return &Directory{config: config, emitter: emitter, eh: eh}
}

func (d *Directory) ServeLDAP(res ldap.ResponseWriter, r *ldap.Request) {
	switch m := r.Message.(type) {
	case *ldap.BindRequest:
		d.serveBind(res, r)
	case *ldap.SearchRequest:
		d.serveSearch(res, r)
	case *ldap.ModifyRequest:
		d.serveModify(res, m, r.Context)
	case *ldap.AddRequest:
		d.serveAdd(res, m, r.Context)
	case *ldap.DeleteRequest:
		d.serveDelete(res, m, r.Context)
	case *ldap.ModifyDNRequest:
		d.serveModifyDn(res, m, r.Context)
	case *ldap.CompareRequest:
		d.serveCompare(res, m, r.Context)
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

	var res *ldap.BindResponse
	switch msg.Auth {
	case ldap.Simple:
		log.Debugf("received bind request with messageId %v, version %v. auth: %v", r.MessageId, msg.Version, msg.Name)

		if msg.Name != "" {
			e := d.getEntry(msg.Name)
			if e == nil {
				res = &ldap.BindResponse{
					Result: ldap.InvalidCredentials,
				}
			} else {
				pw, ok := e.Attributes["userPassword"]
				if !ok {
					res = &ldap.BindResponse{
						Result: ldap.Success,
					}
				} else if pw[0] == msg.Password {
					res = &ldap.BindResponse{
						Result: ldap.Success,
					}
				} else {
					res = &ldap.BindResponse{
						Result: ldap.InvalidCredentials,
					}
				}
			}
		} else {
			res = &ldap.BindResponse{
				Result: ldap.Success,
			}
		}
	default:
		res = &ldap.BindResponse{
			Result:  ldap.AuthMethodNotSupported,
			Message: "server supports only simple auth method",
		}
	}

	m, doMonitor := monitor.LdapFromContext(r.Context)
	if doMonitor {
		l := NewBindLogEvent(msg, res, d.eh, events.NewTraits().WithName(d.config.Info.Name))
		defer func() {
			i := r.Context.Value("time")
			if i != nil {
				t := i.(time.Time)
				l.Duration = time.Now().Sub(t).Milliseconds()
			}
		}()
		m.RequestCounter.WithLabel(d.config.Info.Name, "modify").Add(1)
		m.LastRequest.WithLabel(d.config.Info.Name).Set(float64(time.Now().Unix()))
	}

	rw.Write(res)
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
	return e.Dn == name[0] && baseDN != name[0]
}

func (d *Directory) getEntry(dn string) *Entry {
	for it := d.config.Entries.Iter(); it.Next(); {
		e := it.Value()
		if e.Dn == dn {
			return &e
		}
	}
	return nil
}

func (d *Directory) serveModify(rw ldap.ResponseWriter, r *ldap.ModifyRequest, ctx context.Context) {
	modify := &ModifyRecord{Dn: r.Dn}
	for _, m := range r.Items {
		var t string
		switch m.Operation {
		case ldap.DeleteOperation:
			t = "delete"
		case ldap.AddOperation:
			t = "add"
		case ldap.ReplaceOperation:
			t = "replace"
		}
		a := ModifyAction{
			Type:       t,
			Name:       m.Modification.Type,
			Attributes: map[string][]string{m.Modification.Type: m.Modification.Values},
		}
		modify.Actions = append(modify.Actions, &a)
	}

	err := modify.Apply(d.config.Entries, d.config.Schema)
	var res *ldap.ModifyResponse
	if err != nil {
		var ee *EntryError
		errors.As(err, &ee)
		res = &ldap.ModifyResponse{ResultCode: ee.Code, MatchedDn: r.Dn, Message: err.Error()}
	} else {
		res = &ldap.ModifyResponse{ResultCode: ldap.Success, MatchedDn: r.Dn}
		go d.config.rebuildMemberOf()
	}

	m, doMonitor := monitor.LdapFromContext(ctx)
	if doMonitor {
		l := NewModifyLogEvent(r, res, d.eh, events.NewTraits().WithName(d.config.Info.Name))
		defer func() {
			i := ctx.Value("time")
			if i != nil {
				t := i.(time.Time)
				l.Duration = time.Now().Sub(t).Milliseconds()
			}
		}()
		m.RequestCounter.WithLabel(d.config.Info.Name, "modify").Add(1)
		m.LastRequest.WithLabel(d.config.Info.Name).Set(float64(time.Now().Unix()))
	}

	rw.Write(res)
}

func (d *Directory) Unbind(ctx context.Context) {
	m, doMonitor := monitor.LdapFromContext(ctx)
	if doMonitor {
		l := NewUnbindEvent(d.eh, events.NewTraits().WithName(d.config.Info.Name))
		i := ctx.Value("time")
		if i != nil {
			t := i.(time.Time)
			l.Duration = time.Now().Sub(t).Milliseconds()
		}
		m.RequestCounter.WithLabel(d.config.Info.Name, "unbind").Add(1)
		m.LastRequest.WithLabel(d.config.Info.Name).Set(float64(time.Now().Unix()))
	}
}

func (d *Directory) serveAdd(rw ldap.ResponseWriter, r *ldap.AddRequest, ctx context.Context) {
	add := &AddRecord{
		Dn:         r.Dn,
		Attributes: map[string][]string{},
	}
	for _, attr := range r.Attributes {
		add.Attributes[attr.Type] = attr.Values
	}

	err := add.Apply(d.config.Entries, d.config.Schema)
	var res *ldap.AddResponse
	if err != nil {
		var ee *EntryError
		errors.As(err, &ee)
		res = &ldap.AddResponse{ResultCode: ee.Code, Message: err.Error()}
	} else {
		res = &ldap.AddResponse{ResultCode: ldap.Success, MatchedDn: r.Dn}
		go d.config.rebuildMemberOf()
	}

	m, doMonitor := monitor.LdapFromContext(ctx)
	if doMonitor {
		l := NewAddLogEvent(r, res, d.eh, events.NewTraits().WithName(d.config.Info.Name))
		defer func() {
			i := ctx.Value("time")
			if i != nil {
				t := i.(time.Time)
				l.Duration = time.Now().Sub(t).Milliseconds()
			}
		}()
		m.RequestCounter.WithLabel(d.config.Info.Name, "add").Add(1)
		m.LastRequest.WithLabel(d.config.Info.Name).Set(float64(time.Now().Unix()))
	}

	rw.Write(res)
}

func (d *Directory) serveDelete(rw ldap.ResponseWriter, r *ldap.DeleteRequest, ctx context.Context) {
	del := &DeleteRecord{
		Dn: r.Dn,
	}
	err := del.Apply(d.config.Entries, d.config.Schema)
	var res *ldap.DeleteResponse
	if err != nil {
		var ee *EntryError
		errors.As(err, &ee)
		res = &ldap.DeleteResponse{ResultCode: ee.Code, MatchedDn: del.Dn, Message: err.Error()}
	} else {
		res = &ldap.DeleteResponse{ResultCode: ldap.Success, MatchedDn: del.Dn}
		go d.config.rebuildMemberOf()
	}

	m, doMonitor := monitor.LdapFromContext(ctx)
	if doMonitor {
		l := NewDeleteLogEvent(r, res, d.eh, events.NewTraits().WithName(d.config.Info.Name))
		defer func() {
			i := ctx.Value("time")
			if i != nil {
				t := i.(time.Time)
				l.Duration = time.Now().Sub(t).Milliseconds()
			}
		}()
		m.RequestCounter.WithLabel(d.config.Info.Name, "delete").Add(1)
		m.LastRequest.WithLabel(d.config.Info.Name).Set(float64(time.Now().Unix()))
	}

	rw.Write(res)
}

func (d *Directory) serveModifyDn(rw ldap.ResponseWriter, r *ldap.ModifyDNRequest, ctx context.Context) {
	del := &ModifyDnRecord{
		Dn:            r.Dn,
		NewRdn:        r.NewRdn,
		NewSuperiorDn: r.NewSuperiorDn,
		DeleteOldDn:   r.DeleteOldDn,
	}
	err := del.Apply(d.config.Entries, d.config.Schema)
	var res *ldap.ModifyDNResponse
	if err != nil {
		var ee *EntryError
		errors.As(err, &ee)
		res = &ldap.ModifyDNResponse{ResultCode: ee.Code, MatchedDn: r.Dn, Message: err.Error()}
	} else {
		res = &ldap.ModifyDNResponse{ResultCode: ldap.Success, MatchedDn: r.Dn}
		go d.config.rebuildMemberOf()
	}

	m, doMonitor := monitor.LdapFromContext(ctx)
	if doMonitor {
		l := NewModifyDNLogEvent(r, res, d.eh, events.NewTraits().WithName(d.config.Info.Name))
		defer func() {
			i := ctx.Value("time")
			if i != nil {
				t := i.(time.Time)
				l.Duration = time.Now().Sub(t).Milliseconds()
			}
		}()
		m.RequestCounter.WithLabel(d.config.Info.Name, "modifyDN").Add(1)
		m.LastRequest.WithLabel(d.config.Info.Name).Set(float64(time.Now().Unix()))
	}

	rw.Write(res)
}

func (d *Directory) serveCompare(rw ldap.ResponseWriter, r *ldap.CompareRequest, ctx context.Context) {
	e := d.getEntry(r.Dn)
	var res *ldap.CompareResponse
	if e != nil {
		if a, ok := e.Attributes[r.Attribute]; ok {
			if slices.Contains(a, r.Value) {
				res = &ldap.CompareResponse{ResultCode: ldap.CompareTrue}
			}
		}
	}
	if res == nil {
		res = &ldap.CompareResponse{ResultCode: ldap.CompareFalse}
	}

	m, doMonitor := monitor.LdapFromContext(ctx)
	if doMonitor {
		l := NewCompareLogEvent(r, res, d.eh, events.NewTraits().WithName(d.config.Info.Name))
		defer func() {
			i := ctx.Value("time")
			if i != nil {
				t := i.(time.Time)
				l.Duration = time.Now().Sub(t).Milliseconds()
			}
		}()
		m.RequestCounter.WithLabel(d.config.Info.Name, "compare").Add(1)
		m.LastRequest.WithLabel(d.config.Info.Name).Set(float64(time.Now().Unix()))
	}

	rw.Write(res)
}
