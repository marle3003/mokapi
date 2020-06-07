package models

import (
	"mokapi/providers/parser"

	log "github.com/sirupsen/logrus"
)

type Application struct {
	Services map[string]*ServiceInfo
	Metrics  *Metrics
}

type ServiceInfo struct {
	Service *Service
	Status  string
	Errors  []string
}

type Filter struct {
	Raw   string
	Expr  *parser.Expression
	Error string
}

func (a Application) AddOrUpdateService(s *Service, errors []string) {
	if si, ok := a.Services[s.Name]; ok {
		si.Service = s
		si.Errors = errors
	} else {
		a.Services[s.Name] = NewServiceInfo(s, errors)
	}
}

func NewApplication() *Application {
	return &Application{Services: make(map[string]*ServiceInfo), Metrics: NewMetrics()}
}

func NewServiceInfo(s *Service, errors []string) *ServiceInfo {
	return &ServiceInfo{Service: s, Errors: errors}
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
