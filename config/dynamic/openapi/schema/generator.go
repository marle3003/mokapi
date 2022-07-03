package schema

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	log "github.com/sirupsen/logrus"
	"math"
	"math/rand"
	"mokapi/sortedmap"
	"reflect"
	"time"
)

type Generator struct {
	randomNumber *rand.Rand
}

type builder struct {
	requests []*Schema
}

func NewGenerator() *Generator {
	return &Generator{randomNumber: rand.New(rand.NewSource(time.Now().Unix()))}
}

func (g *Generator) New(ref *Ref) interface{} {
	return newBuilder().create(ref)
}

func newBuilder() *builder {
	return &builder{}
}

func (b *builder) create(ref *Ref) interface{} {
	if ref == nil || ref.Value == nil {
		return nil
	}
	schema := ref.Value

	switch {
	case schema.Example != nil:
		return schema.Example
	case schema.Enum != nil && len(schema.Enum) > 0:
		return schema.Enum[gofakeit.Number(0, len(schema.Enum)-1)]
	default:
		switch schema.Type {
		case "object":
			return b.createObject(schema)
		case "array":
			return b.createArray(schema)
		case "boolean":
			return gofakeit.Bool()
		case "integer", "number":
			return b.createNumber(schema)
		case "string":
			return b.createString(schema)
		default:
			if len(schema.AnyOf) > 0 {
				i := gofakeit.Number(0, len(schema.AnyOf)-1)
				return b.create(schema.AnyOf[i])
			}
			if len(schema.AllOf) > 0 {
				m := sortedmap.NewLinkedHashMap()
				for _, all := range schema.AllOf {
					if all.Value == nil {
						continue
					}
					if all.Value.Type != "object" {
						log.Info("allOf only supports type of object")
						continue
					}
					o := b.create(all).(*sortedmap.LinkedHashMap)
					m.Merge(o)

				}
				return m
			}
			return nil
		}
	}
}

func (b *builder) createObject(s *Schema) interface{} {
	// recursion guard. Currently, we use a fixed depth: 1
	n := len(b.requests)
	numRequestsSameAsThisOne := 0
	for _, r := range b.requests {
		if s == r {
			numRequestsSameAsThisOne++
		}
	}
	if numRequestsSameAsThisOne > 1 {
		return nil
	}
	b.requests = append(b.requests, s)
	// remove schemas from guard
	defer func() { b.requests = b.requests[:n] }()

	m := sortedmap.NewLinkedHashMap()

	if s.IsDictionary() {
		length := gofakeit.Number(1, 10)
		for i := 0; i < length; i++ {
			m.Set(gofakeit.Word(), b.create(s.AdditionalProperties.Ref))
		}
		return m
	}

	if s.Properties == nil || s.Properties.Value == nil {
		return m
	}

	for it := s.Properties.Value.Iter(); it.Next(); {
		key := it.Key().(string)
		propSchema := it.Value().(*Ref)
		value := b.create(propSchema)
		m.Set(key, value)
	}

	return m
}

func getType(s *Schema) reflect.Type {
	switch s.Type {
	case "integer":
		if s.Format == "int32" {
			return reflect.TypeOf(int32(0))
		}
		return reflect.TypeOf(int64(0))
	case "number":
		if s.Format == "float32" {
			return reflect.TypeOf(float32(0))
		}
		return reflect.TypeOf(float64(0))
	case "string":
		return reflect.TypeOf("")
	case "boolean":
		return reflect.TypeOf(false)
	case "array":
		return reflect.SliceOf(getType(s.Items.Value))
	}

	panic(fmt.Sprintf("type %v not implemented", s.Type))
}

func (b *builder) createString(s *Schema) string {
	if len(s.Format) > 0 {
		return b.createStringByFormat(s.Format)
	} else if len(s.Pattern) > 0 {
		return gofakeit.Generate(fmt.Sprintf("{regex:%v}", s.Pattern))
	}
	return gofakeit.Lexify("???????????????")
}

func (b *builder) createNumberExclusive(s *Schema) interface{} {
	for i := 0; i < 10; i++ {
		n := b.createNumber(s)
		if *s.ExclusiveMinimum && s.Minimum == n {
			continue
		}
		if *s.ExclusiveMaximum && s.Maximum == n {
			continue
		}
		return n
	}
	log.Errorf("unable to find a valid number with exclusive")
	return nil
}

func (b *builder) createNumber(s *Schema) interface{} {
	if s.ExclusiveMinimum != nil && (*s.ExclusiveMinimum) ||
		s.ExclusiveMaximum != nil && (*s.ExclusiveMaximum) {
		return b.createNumberExclusive(s)
	}

	if s.Type == "number" {
		if s.Format == "float" {
			if s.Minimum == nil && s.Maximum == nil {
				return gofakeit.Float32()
			}
			max := float32(math.MaxFloat32)
			min := max * -1
			if s.Minimum != nil {
				min = float32(*s.Minimum)
			}
			if s.Maximum != nil {
				max = float32(*s.Maximum)
			}
			return gofakeit.Float32Range(min, max)
		} else {
			if s.Minimum == nil && s.Maximum == nil {
				return gofakeit.Float64()
			}
			max := math.MaxFloat64
			min := max * -1
			if s.Minimum != nil {
				min = *s.Minimum
			}
			if s.Maximum != nil {
				max = *s.Maximum
			}
			return gofakeit.Float64Range(min, max)
		}

	} else if s.Type == "integer" {
		if s.Minimum == nil && s.Maximum == nil {
			if s.Format == "int32" {
				return gofakeit.Int32()
			} else {
				return gofakeit.Int64()
			}
		}
		if s.Format == "int32" {
			max := math.MaxInt32
			min := math.MinInt32
			if s.Minimum != nil {
				min = int(*s.Minimum)
			}
			if s.Maximum != nil {
				max = int(*s.Maximum)
			}

			// gofakeit uses Intn function which panics if number is <= 0
			return int32(math.Round(float64(gofakeit.Float32Range(float32(min), float32(max)))))
		} else {
			max := math.MaxInt64
			min := math.MinInt64
			if s.Minimum != nil {
				min = int(*s.Minimum)
			}
			if s.Maximum != nil {
				max = int(*s.Maximum)
			}

			return int64(math.Round(gofakeit.Float64Range(float64(min), float64(max))))
		}
	}

	return 0
}

func (b *builder) createArray(s *Schema) (r []interface{}) {
	maxItems := 5
	if s.MaxItems != nil {
		maxItems = *s.MaxItems + 1
	}
	minItems := 0
	if s.MinItems != nil {
		minItems = *s.MinItems
	}

	var f func() interface{}

	if s.UniqueItems && s.Items.Value != nil && len(s.Items.Value.Enum) > 0 {
		if maxItems > len(s.Items.Value.Enum) {
			maxItems = len(s.Items.Value.Enum)
		}
		f = func() interface{} {
			i := gofakeit.Number(0, len(s.Items.Value.Enum)-1)
			return s.Items.Value.Enum[i]
		}
		if s.ShuffleItems {
			defer func() {
				rand.Shuffle(len(r), func(i, j int) { r[i], r[j] = r[j], r[i] })
			}()
		}
	} else {
		f = func() interface{} {
			return b.create(s.Items)
		}
	}

	length := minItems
	if maxItems-minItems > 0 {
		length = gofakeit.Number(minItems, maxItems)
	}
	r = make([]interface{}, length)

	for i := range r {
		if s.UniqueItems {
			r[i] = b.getUnique(r, f)
		} else {
			r[i] = f()
		}
	}
	return r
}

func (b *builder) createStringByFormat(format string) string {
	switch format {
	case "date":
		return gofakeit.Date().Format("2006-01-02")
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
	default:
		return gofakeit.Generate(format)
	}
}

func (b *builder) getUnique(s []interface{}, gen func() interface{}) interface{} {
	for i := 0; i < 10; i++ {
		v := gen()
		if !contains(s, v) {
			return v
		}
	}
	panic("can not fill array with unique items")
}

func contains(s []interface{}, v interface{}) bool {
	for _, i := range s {
		if reflect.DeepEqual(i, v) {
			return true
		}
	}
	return false
}
