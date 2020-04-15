package data

import "mokapi/config"

type DataProvider interface {
	Provide(parameters map[string]string, schema *config.Schema) (interface{}, error)
}
