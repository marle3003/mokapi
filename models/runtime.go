package models

import (
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/ldap"
	"mokapi/config/dynamic/openapi"
	"mokapi/config/dynamic/smtp"
	"time"
)

type Runtime struct {
	OpenApi  map[string]*openapi.Config
	Ldap     map[string]*ldap.Config
	AsyncApi map[string]*asyncApi.Config
	Smtp     map[string]*smtp.Config
	Metrics  *Metrics
}

func NewRuntime() *Runtime {
	return &Runtime{
		OpenApi:  make(map[string]*openapi.Config),
		Ldap:     make(map[string]*ldap.Config),
		AsyncApi: make(map[string]*asyncApi.Config),
		Smtp:     make(map[string]*smtp.Config),
		Metrics:  newMetrics(),
	}
}

type WorkflowLog struct {
	Name     string
	Logs     []string
	Duration time.Duration
}
