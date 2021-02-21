package models

import (
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/asyncApi"
	event "mokapi/models/eventService"
)

type Application struct {
	WebServices   map[string]*WebServiceInfo
	LdapServices  map[string]*LdapServiceInfo
	EventServices map[string]*event.Service
	Metrics       *Metrics
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
	a.ApplyEventService(config.AsyncApi)
}

func NewApplication() *Application {
	return &Application{
		WebServices:   make(map[string]*WebServiceInfo),
		LdapServices:  make(map[string]*LdapServiceInfo),
		EventServices: make(map[string]*event.Service),
		Metrics:       newMetrics(),
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

func (a *Application) ApplyEventService(config map[string]*asyncApi.Config) {
	for filePath, item := range config {
		key := filePath
		if len(item.Info.Name) > 0 {
			key = item.Info.Name
		}
		s, found := a.EventServices[key]
		if !found {
			s = event.NewService()
			a.EventServices[key] = s
		}
		s.Apply(item, filePath)
	}
}
