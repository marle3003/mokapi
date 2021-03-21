package openapi

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"math/rand"
	"time"
)

type Generator struct {
	randomNumber *rand.Rand
}

func NewGenerator() *Generator {
	return &Generator{randomNumber: rand.New(rand.NewSource(time.Now().Unix()))}
}

func (g *Generator) New(schema *SchemaRef) interface{} {
	if schema == nil || schema.Value == nil {
		return nil
	}

	if schema.Value.Example != nil {
		return schema.Value.Example
	} else if schema.Value.Enum != nil && len(schema.Value.Enum) > 0 {
		return schema.Value.Enum[g.randomNumber.Intn(len(schema.Value.Enum))]
	}

	if schema.Value.Type == "object" {
		obj := make(map[string]interface{})
		for name, propSchema := range schema.Value.Properties.Value {
			value := g.New(propSchema)
			obj[name] = value
		}
		return obj
	} else if schema.Value.Type == "array" {
		length := rand.Intn(5)
		obj := make([]interface{}, length)
		for i := range obj {
			obj[i] = g.New(schema.Value.Items)
		}
		return obj
	} else {
		if len(schema.Value.Faker) > 0 {
			switch schema.Value.Faker {
			case "numbers.uint32":
				return gofakeit.Uint32()
			default:
				return gofakeit.Generate(fmt.Sprintf("{%s}", schema.Value.Faker))
			}
		} else if schema.Value.Type == "integer" {
			return gofakeit.Int32()
		} else if schema.Value.Type == "string" {
			return gofakeit.Lexify("???????????????")
		}
	}
	return nil
}
