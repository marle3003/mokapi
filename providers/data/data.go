package data

import "mokapi/models"

type Provider interface {
	Provide(parameters map[string]string, schema *models.Schema) (interface{}, error)
	Close()
}
