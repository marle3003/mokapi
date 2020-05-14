package data

import "mokapi/models"

type Provider interface {
	Provide(name string, schema *models.Schema) (interface{}, error)
	Close()
}
