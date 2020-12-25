package models

import (
	"mokapi/config/dynamic"
	"mokapi/providers/parser"

	log "github.com/sirupsen/logrus"
)

type Application struct {
	WebServices  map[string]*WebServiceInfo
	LdapServices map[string]*LdapServiceInfo
	Metrics      *Metrics
}

type WebServiceInfo struct {
	Data   *WebService
	Status string
	Errors []string
}

type LdapServiceInfo struct {
	Data   *LdapServer
	Status string
	Errors []string
}

type Filter struct {
	Raw   string
	Expr  *parser.Expression
	Error string
}

func (a Application) Apply(config *dynamic.Configuration) {
	a.ApplyWebService(config.OpenApi)
	a.ApplyLdap(config.Ldap)
}

func NewApplication() *Application {
	return &Application{
		WebServices:  make(map[string]*WebServiceInfo),
		LdapServices: make(map[string]*LdapServiceInfo),
		Metrics:      NewMetrics(),
	}
}

func NewServiceInfo() *WebServiceInfo {
	webService := &WebService{Servers: make([]Server, 0), Endpoint: make(map[string]*Endpoint)}

	return &WebServiceInfo{Data: webService, Errors: make([]string, 0)}
}

func NewLdapServiceInfo() *LdapServiceInfo {
	ldapServer := &LdapServer{Root: &Entry{Attributes: make(map[string][]string)}}

	return &LdapServiceInfo{Data: ldapServer, Errors: make([]string, 0)}
}

func NewFilter(s string) *Filter {
	f := &Filter{Raw: s}
	expr, error := parser.ParseExpression(s)
	if error != nil {
		log.Error(error)
		f.Error = error.Error()
	} else {
		f.Expr = expr
	}

	return f
}

func (f *Filter) IsValid() bool {
	return f.Expr != nil && len(f.Error) == 0
}
