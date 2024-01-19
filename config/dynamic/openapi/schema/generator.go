package schema

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"math"
	"math/rand"
	"reflect"
	"time"
)

const (
	lowerChars   = "abcdefghijklmnopqrstuvwxyz"
	upperChars   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numericChars = "0123456789"
	specialChars = "!@#$%&*+-_=?:;,.|(){}<>"
	spaceChar    = " "
	allChars     = lowerChars + upperChars + numericChars + specialChars + spaceChar
)

type Generator struct {
	r *rand.Rand
}

type builder struct {
	requests []*Schema
	r        *rand.Rand
}

func NewGenerator() *Generator {
	return &Generator{r: rand.New(rand.NewSource(time.Now().Unix()))}
}

func NewGeneratorWithSeed(seed int64) *Generator {
	return &Generator{r: rand.New(rand.NewSource(seed))}
}

func (g *Generator) New(ref *Ref) (interface{}, error) {
	i, err := newBuilder(g.r).create(ref)
	if err != nil {
		return nil, fmt.Errorf("generating data failed: %w", err)
	}
	return i, nil
}

func newBuilder(r *rand.Rand) *builder {
	return &builder{r: r}
}

func (b *builder) create(ref *Ref) (interface{}, error) {
	if ref == nil || ref.Value == nil {
		return nil, nil
	}
	schema := ref.Value

	switch {
	case schema.Example != nil:
		return schema.Example, nil
	case schema.Enum != nil && len(schema.Enum) > 0:
		return schema.Enum[gofakeit.Number(0, len(schema.Enum)-1)], nil
	default:
		switch schema.Type {
		case "object":
			return b.createObject(schema)
		case "array":
			return b.createArray(schema)
		case "boolean":
			return gofakeit.Bool(), nil
		case "integer", "number":
			return b.createNumber(schema)
		case "string":
			return b.createString(schema), nil
		default:
			if len(schema.AnyOf) > 0 {
				i := gofakeit.Number(0, len(schema.AnyOf)-1)
				return b.create(schema.AnyOf[i])
			}
			if len(schema.AllOf) > 0 {
				result := map[string]interface{}{}
				for _, all := range schema.AllOf {
					if all == nil || all.Value == nil {
						continue
					}
					if all.Value.Type != "object" {
						return nil, fmt.Errorf("allOf expects type of object but got %v", all.Value.Type)
					}
					o, err := b.create(all)
					if err != nil {
						return nil, fmt.Errorf("allOf expects to be valid against all of subschemas: %w", err)
					}
					m := o.(map[string]interface{})
					for key, value := range m {
						result[key] = value
					}

				}
				return result, nil
			}
			return nil, nil
		}
	}
}

func (b *builder) createObject(s *Schema) (interface{}, error) {
	if s.Nullable {
		n := gofakeit.Float32Range(0, 1)
		if n < 0.05 {
			return nil, nil
		}
	}

	// recursion guard. Currently, we use a fixed depth: 1
	n := len(b.requests)
	numRequestsSameAsThisOne := 0
	for _, r := range b.requests {
		if s == r {
			numRequestsSameAsThisOne++
		}
	}
	if numRequestsSameAsThisOne > 1 {
		return nil, nil
	}
	b.requests = append(b.requests, s)
	// remove schemas from guard
	defer func() { b.requests = b.requests[:n] }()

	//m := sortedmap.NewLinkedHashMap()
	m := map[string]interface{}{}

	if s.IsDictionary() {
		length := gofakeit.Number(1, 10)
		for i := 0; i < length; i++ {
			v, err := b.create(s.AdditionalProperties.Ref)
			if err != nil {
				return nil, err
			}
			//m.Set(gofakeit.Word(), v)
			m[gofakeit.Word()] = v
		}
		return m, nil
	}

	if s.Properties == nil {
		return m, nil
	}

	for it := s.Properties.Iter(); it.Next(); {
		key := it.Key()
		propSchema := it.Value()
		value, err := b.create(propSchema)
		if err != nil {
			return nil, err
		}
		//m.Set(key, value)
		m[key] = value
	}

	return m, nil
}

func (b *builder) createString(s *Schema) interface{} {
	if s.Nullable {
		n := gofakeit.Float32Range(0, 1)
		if n < 0.05 {
			return nil
		}
	}

	if len(s.Format) > 0 {
		return b.createStringByFormat(s.Format)
	} else if len(s.Pattern) > 0 {
		return gofakeit.Generate(fmt.Sprintf("{regex:%v}", s.Pattern))
	}

	minLength := 0
	maxLength := 15

	if s.MinLength != nil {
		minLength = *s.MinLength
	}
	if s.MaxLength != nil {
		maxLength = *s.MaxLength
	} else if minLength > maxLength {
		maxLength += minLength
	}

	categories := []interface{}{0, 1, 2, 3}
	weights := []float32{5, 0.5, 0.3, 0.1}
	letters := lowerChars + upperChars

	length := gofakeit.IntRange(minLength, maxLength)
	result := make([]rune, length)
	for i := 0; i < length; i++ {
		c, _ := gofakeit.Weighted(categories, weights)

		switch c {
		case 0:
			n := gofakeit.IntRange(0, len(letters)-1)
			result[i] = rune(letters[n])
		case 1:
			n := gofakeit.IntRange(0, len(numericChars)-1)
			result[i] = rune(numericChars[n])
		case 2:
			result[i] = ' '
		case 3:
			n := gofakeit.IntRange(0, len(specialChars)-1)
			result[i] = rune(specialChars[n])
		}
	}
	return string(result)
}

func (b *builder) createNumber(s *Schema) (interface{}, error) {
	if s.Nullable {
		n := gofakeit.Float32Range(0, 1)
		if n < 0.05 {
			return nil, nil
		}
	}

	f := false
	if s.ExclusiveMinimum != nil && (*s.ExclusiveMinimum) {
		ss := *s
		ss.ExclusiveMinimum = &f
		if ss.Minimum != nil {
			ss.Minimum = toPointerF(*ss.Minimum + 1)
		}
		return b.createNumber(&ss)
	}
	if s.ExclusiveMaximum != nil && (*s.ExclusiveMaximum) {
		ss := *s
		ss.ExclusiveMaximum = &f
		if ss.Maximum != nil {
			ss.Maximum = toPointerF(*ss.Maximum - 1)
		}
		return b.createNumber(&ss)
	}

	if s.Minimum != nil && s.Maximum != nil &&
		(*s.Minimum) > (*s.Maximum) {
		return nil, fmt.Errorf("invalid minimum '%v' and maximum '%v' in %v", *s.Minimum, *s.Maximum, s)
	}

	if s.Type == "number" {
		if s.Format == "float" {
			if s.Minimum == nil && s.Maximum == nil {
				return gofakeit.Float32(), nil
			}
			max := float32(math.MaxFloat32)
			min := max * -1
			if s.Minimum != nil {
				min = float32(*s.Minimum)
			}
			if s.Maximum != nil {
				max = float32(*s.Maximum)
			}
			return gofakeit.Float32Range(min, max), nil
		} else {
			if s.Minimum == nil && s.Maximum == nil {
				return gofakeit.Float64(), nil
			}
			max := math.MaxFloat64
			min := max * -1
			if s.Minimum != nil {
				min = *s.Minimum
			}
			if s.Maximum != nil {
				max = *s.Maximum
			}
			return gofakeit.Float64Range(min, max), nil
		}

	} else if s.Type == "integer" {
		if s.Minimum == nil && s.Maximum == nil {
			if s.Format == "int32" {
				return gofakeit.Int32(), nil
			} else {
				return gofakeit.Int64(), nil
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
			return int32(math.Round(float64(gofakeit.Float32Range(float32(min), float32(max))))), nil
		} else {
			max := math.MaxInt64
			min := math.MinInt64
			if s.Minimum != nil {
				min = int(*s.Minimum)
			}
			if s.Maximum != nil {
				max = int(*s.Maximum)
			}

			return int64(math.Round(gofakeit.Float64Range(float64(min), float64(max)))), nil
		}
	}

	return 0, nil
}

func (b *builder) createArray(s *Schema) (r []interface{}, err error) {
	if s.Nullable {
		n := gofakeit.Float32Range(0, 1)
		if n < 0.05 {
			return nil, nil
		}
	}

	maxItems := 5
	if s.MaxItems != nil {
		maxItems = *s.MaxItems
	}
	minItems := 0
	if s.MinItems != nil {
		minItems = *s.MinItems
	}

	if s.ShuffleItems {
		defer func() {
			b.r.Shuffle(len(r), func(i, j int) { r[i], r[j] = r[j], r[i] })
		}()
	}

	var f func() (interface{}, error)

	if s.UniqueItems && s.Items.Value != nil && len(s.Items.Value.Enum) > 0 {
		if maxItems > len(s.Items.Value.Enum) {
			maxItems = len(s.Items.Value.Enum)
		}
		f = func() (interface{}, error) {
			i := gofakeit.Number(0, len(s.Items.Value.Enum)-1)
			return s.Items.Value.Enum[i], nil
		}
	} else {
		f = func() (interface{}, error) {
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
			r[i], err = b.getUnique(r, f)
			if err != nil {
				return nil, fmt.Errorf("%w for %v", err, s)
			}
		} else {
			r[i], err = f()
			if err != nil {
				return
			}
		}
	}
	return r, nil
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

func (b *builder) getUnique(s []interface{}, gen func() (interface{}, error)) (interface{}, error) {
	for i := 0; i < 10; i++ {
		v, err := gen()
		if err != nil {
			return nil, err
		}
		if !contains(s, v) {
			return v, nil
		}
	}
	return nil, fmt.Errorf("can not fill array with unique items")
}

func contains(s []interface{}, v interface{}) bool {
	for _, i := range s {
		if reflect.DeepEqual(i, v) {
			return true
		}
	}
	return false
}

func toPointerF(f float64) *float64 {
	return &f
}