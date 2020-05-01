package data

import "mokapi/service"

type Provider interface {
	Provide(parameters map[string]string, schema *service.Schema) (interface{}, error)
	Close()
}
