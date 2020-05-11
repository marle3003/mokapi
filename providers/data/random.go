package data

import (
	"fmt"
	"math/rand"
	"mokapi/models"

	"github.com/brianvoe/gofakeit/v4"
)

type RandomDataProvider struct {
}

func NewRandomDataProvider() *RandomDataProvider {
	return &RandomDataProvider{}
}

func (provider *RandomDataProvider) Provide(parameters map[string]string, schema *models.Schema) (interface{}, error) {
	return provider.getRandomObject(schema), nil
}

func (provider *RandomDataProvider) Close() {}

func (provider *RandomDataProvider) getRandomObject(schema *models.Schema) interface{} {
	if schema.Type == "object" {
		obj := make(map[string]interface{})
		for name, propSchema := range schema.Properties {
			value := provider.getRandomObject(propSchema)
			obj[name] = value
		}
		return obj
	} else if schema.Type == "array" {
		length := rand.Intn(5)
		obj := make([]interface{}, length)
		for i := range obj {
			obj[i] = provider.getRandomObject(schema.Items)
		}
		return obj
	} else {
		if len(schema.Faker) > 0 {
			switch schema.Faker {
			case "numbers.uint32":
				return gofakeit.Uint32()
			default:
				return gofakeit.Generate(fmt.Sprintf("{%s}", schema.Faker))
			}
		} else if schema.Type == "integer" {
			return gofakeit.Int32()
		} else if schema.Type == "string" {
			return gofakeit.Lexify("???????????????")
		}
	}
	return nil
}
