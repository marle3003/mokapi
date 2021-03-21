package models

import (
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/ldap"
	"mokapi/config/dynamic/openapi"
)

type Runtime struct {
	OpenApi  map[string]*openapi.Config
	Ldap     map[string]*ldap.Config
	AsyncApi map[string]*asyncApi.Config
	Metrics  *Metrics
}

func NewRuntime() *Runtime {
	return &Runtime{
		OpenApi:  make(map[string]*openapi.Config),
		Ldap:     make(map[string]*ldap.Config),
		AsyncApi: make(map[string]*asyncApi.Config),
		Metrics:  newMetrics(),
	}
}
