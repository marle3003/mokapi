package models

import (
	"mokapi/config/dynamic"
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

func (a Application) Apply(config *dynamic.Configuration) {
	a.ApplyWebService(config.OpenApi)
	a.ApplyLdap(config.Ldap)
}

func NewApplication() *Application {
	return &Application{
		WebServices:  make(map[string]*WebServiceInfo),
		LdapServices: make(map[string]*LdapServiceInfo),
		Metrics:      newMetrics(),
	}
}

func NewServiceInfo() *WebServiceInfo {
	webService := &WebService{Servers: make([]Server, 0), Endpoint: make(map[string]*Endpoint), Models: make(map[string]*Schema)}

	return &WebServiceInfo{Data: webService, Errors: make([]string, 0)}
}

func NewLdapServiceInfo() *LdapServiceInfo {
	ldapServer := &LdapServer{Root: &Entry{Attributes: make(map[string][]string)}}

	return &LdapServiceInfo{Data: ldapServer, Errors: make([]string, 0)}
}
