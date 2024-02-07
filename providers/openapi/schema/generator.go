package schema

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"mokapi/json/generator"
)

const (
	lowerChars   = "abcdefghijklmnopqrstuvwxyz"
	upperChars   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numericChars = "0123456789"
	specialChars = "!@#$%&*+-_=?:;,.|(){}<>"
	spaceChar    = " "
	allChars     = lowerChars + upperChars + numericChars + specialChars + spaceChar
)

type builder struct {
	requests []*Schema
}

func CreateValue(ref *Ref) (interface{}, error) {
	i, err := newBuilder().create(ref)
	if err != nil {
		return nil, fmt.Errorf("generating data failed: %w", err)
	}
	return i, nil
}

func newBuilder() *builder {
	return &builder{}
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

	m := map[string]interface{}{}

	if s.IsDictionary() {
		length := gofakeit.Number(1, 10)
		for i := 0; i < length; i++ {
			v, err := b.create(s.AdditionalProperties.Ref)
			if err != nil {
				return nil, err
			}
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
		m[key] = value
	}

	return m, nil
}

func (b *builder) createString(s *Schema) interface{} {
	opt := generator.StringOptions{
		MinLength: s.MinLength,
		MaxLength: s.MaxLength,
		Format:    s.Format,
		Pattern:   s.Pattern,
		Nullable:  s.Nullable,
	}

	return generator.NewString(opt)
}

func (b *builder) createNumber(s *Schema) (interface{}, error) {
	opt := generator.NumberOptions{
		Minimum:  s.Minimum,
		Maximum:  s.Maximum,
		Nullable: s.Nullable,
	}

	if s.Type == "number" {
		if s.Format == "float" {
			opt.Format = generator.Float32
		} else {
			opt.Format = generator.Float64
		}
	} else if s.Type == "integer" {
		if s.Format == "int32" {
			opt.Format = generator.Int32
		} else {
			opt.Format = generator.Int64
		}
	}

	if s.Minimum != nil && s.ExclusiveMinimum != nil {
		min := *s.Minimum
		opt.Minimum = toPointerF(min + 1e-15)
	}
	if s.Maximum != nil && s.ExclusiveMaximum != nil {
		opt.Maximum = toPointerF(*s.Maximum - 1e-15)
	}

	n, err := generator.NewNumber(opt)
	if err != nil {
		return nil, fmt.Errorf("%w in %s", err, s)
	}
	return n, nil
}

func (b *builder) createArray(s *Schema) ([]interface{}, error) {
	opt := generator.ArrayOptions{
		MinItems:    s.MinItems,
		MaxItems:    s.MaxItems,
		Shuffle:     s.ShuffleItems,
		UniqueItems: s.UniqueItems,
		Nullable:    s.Nullable,
	}

	gen := func() (interface{}, error) {
		return b.create(s.Items)
	}
	if s.Items.Value != nil && len(s.Items.Value.Enum) > 0 {
		if opt.MaxItems != nil && *opt.MaxItems > len(s.Items.Value.Enum) {
			maxItems := len(s.Items.Value.Enum)
			opt.MaxItems = &maxItems
		}
		gen = func() (interface{}, error) {
			i := gofakeit.Number(0, len(s.Items.Value.Enum)-1)
			return s.Items.Value.Enum[i], nil
		}
	}

	r, err := generator.NewArray(opt, gen)
	if err != nil {
		return nil, fmt.Errorf("%w for %s", err, s)
	}
	return r, nil
}

func toPointerF(f float64) *float64 {
	return &f
}
