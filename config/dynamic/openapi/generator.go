package openapi

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"math"
	"math/rand"
	"strings"
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
		return g.newArray(schema.Value)
	} else {
		if len(schema.Value.Faker) > 0 {
			if strings.HasPrefix(schema.Value.Faker, "{") {
				return gofakeit.Generate(schema.Value.Faker)
			}
			return gofakeit.Generate(fmt.Sprintf("{%s}", schema.Value.Faker))
		} else {
			switch schema.Value.Type {
			case "boolean":
				return gofakeit.Bool()
			case "integer", "number":
				return getNumber(schema.Value)
			case "string":
				if len(schema.Value.Format) > 0 {
					return getByFormat(schema.Value.Format)
				} else if len(schema.Value.Pattern) > 0 {
					return gofakeit.Generate(fmt.Sprintf("{regex:%v}", schema.Value.Pattern))
				}
				return gofakeit.Lexify("???????????????")
			}
		}
	}
	return nil
}

func getNumber(s *Schema) string {
	if s.Type == "number" {
		if s.Format == "float" {
			if s.Minimum == nil && s.Maximum == nil {
				return fmt.Sprintf("%v", gofakeit.Float32())
			}
			max := float32(math.MaxFloat32)
			min := max * -1
			if s.Minimum != nil {
				min = float32(*s.Minimum)
			}
			if s.Maximum != nil {
				max = float32(*s.Maximum)
			}
			return fmt.Sprintf("%v", gofakeit.Float32Range(min, max))
		} else {
			if s.Minimum == nil && s.Maximum == nil {
				return fmt.Sprintf("%v", gofakeit.Float64())
			}
			max := math.MaxFloat64
			min := max * -1
			if s.Minimum != nil {
				min = *s.Minimum
			}
			if s.Maximum != nil {
				max = *s.Maximum
			}
			return fmt.Sprintf("%v", gofakeit.Float64Range(min, max))
		}

	} else if s.Type == "integer" {
		if s.Minimum == nil && s.Maximum == nil {
			if s.Format == "int32" {
				return fmt.Sprintf("%v", gofakeit.Int32())
			} else {
				return fmt.Sprintf("%v", gofakeit.Int64())
			}
		}
		max := math.MaxInt64
		min := math.MinInt64
		if s.Minimum != nil {
			min = int(*s.Minimum)
		}
		if s.Maximum != nil {
			max = int(*s.Maximum)
		}
		return fmt.Sprintf("%v", gofakeit.Number(min, max))
	}

	return "0"
}

func (g *Generator) newArray(s *Schema) (r []interface{}) {
	maxItems := 5
	if s.MaxItems != nil {
		maxItems = *s.MaxItems + 1
	}
	minItems := 0
	if s.MinItems != nil {
		minItems = *s.MinItems
	}

	var f func(i int) interface{}

	if s.UniqueItems && s.Items.Value != nil && len(s.Items.Value.Enum) > 0 {
		if maxItems > len(s.Items.Value.Enum) {
			maxItems = len(s.Items.Value.Enum)
		}
		f = func(i int) interface{} {
			return s.Items.Value.Enum[i]
		}
		defer func() {
			rand.Shuffle(len(r), func(i, j int) { r[i], r[j] = r[j], r[i] })
		}()
	} else {
		f = func(i int) interface{} {
			return g.New(s.Items)
		}
	}

	length := rand.Intn(maxItems-minItems) + minItems
	r = make([]interface{}, length)

	for i := range r {
		r[i] = f(i)
	}
	return r
}

func getByFormat(format string) string {
	switch format {
	case "date":
		return gofakeit.Generate("{year}-{month}-{day}")
	case "date-time":
		return gofakeit.Generate("{date}")
	case "password":
		return gofakeit.Generate("{password}")
	case "email":
		return gofakeit.Generate("{email}")
	case "uuid":
		return gofakeit.Generate("{uuid}")
	case "uri":
		return gofakeit.Generate("{url}")
	case "hostname":
		return gofakeit.Generate("{domainname}")
	case "ipv4":
		return gofakeit.Generate("{ipv4address}")
	case "ipv6":
		return gofakeit.Generate("{ipv6address}")
	}

	return ""
}
