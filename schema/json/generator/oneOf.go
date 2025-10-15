package generator

import (
	"fmt"
	"mokapi/schema/json/parser"
	"mokapi/schema/json/schema"

	"github.com/brianvoe/gofakeit/v6"
)

const maxOneOfTries = 20

func (r *resolver) oneOf(req *Request) (*faker, error) {
	s := req.Schema
	p := parser.Parser{}
	f := func() (any, error) {
		index := gofakeit.Number(0, len(s.OneOf)-1)
		var err error
		for i := 0; i < len(s.OneOf); i++ {

			selected := selectIndexAndSubtractOthers(index, s.OneOf...)
			var data any
			err = fakeWithRetries(maxOneOfTries, func() error {
				fake, err := r.resolve(req.WithSchema(selected), true)
				if err != nil {
					return err
				}
				data, err = fake.fake()
				if err != nil {
					return err
				}
				for i, one := range s.OneOf {
					if i == index {
						continue
					}
					_, err = p.ParseWith(data, one)
					if err == nil {
						return fmt.Errorf("data is valid against more of the given oneOf subschemas")
					}
				}
				return nil
			})
			if err == nil {
				return data, nil
			}
			index = (index + 1) % len(s.OneOf)
		}
		return nil, fmt.Errorf("cannot create data with one of the subschemas in 'oneOf': %w", err)
	}
	return newFaker(f), nil
}

func selectIndexAndSubtractOthers(index int, schemas ...*schema.Schema) *schema.Schema {
	selected := schemas[index]
	for i, other := range schemas {
		if i == index {
			continue
		}
		selected = subtractSchema(selected, other)
	}
	return selected
}

func subtractSchema(target, subtract *schema.Schema) *schema.Schema {
	result := target.Clone()

	subtractTypes(result, subtract)

	return result
}

func subtractTypes(target, subtract *schema.Schema) {
	for _, v := range subtract.Type {
		if schemaNoConstraintsForType(subtract, v) {
			var out schema.Types
			for _, t := range target.Type {
				if t != v {
					out = append(out, t)
				}
				if v == "integer" && target.IsNumber() {
					if target.Not != nil {
						target.Not = &schema.Schema{
							AnyOf: []*schema.Schema{
								target.Not, {Type: schema.Types{"integer"}},
							},
						}
					} else {
						target.Not = &schema.Schema{Type: schema.Types{"integer"}}
					}
				}
			}
			target.Type = out
		}
	}
}

func schemaNoConstraintsForType(s *schema.Schema, typeName string) bool {
	switch typeName {
	case "string":
		if s.MinLength != nil ||
			s.MaxLength != nil ||
			s.Pattern != "" ||
			s.Format != "" {
			return false
		}
	case "number", "integer":
		if s.Minimum != nil ||
			s.Maximum != nil ||
			s.MultipleOf != nil ||
			s.ExclusiveMinimum != nil ||
			s.ExclusiveMaximum != nil {
			return false
		}
	case "array":
		if s.Items != nil ||
			s.PrefixItems != nil ||
			s.Contains != nil ||
			s.MinItems != nil ||
			s.MaxItems != nil ||
			s.UniqueItems != nil ||
			s.MinContains != nil ||
			s.MaxContains != nil {
			return false
		}
	case "object":
		if s.Properties.Len() > 0 ||
			len(s.Required) > 0 ||
			s.MinProperties != nil ||
			s.MaxProperties != nil ||
			len(s.PatternProperties) > 0 ||
			len(s.DependentSchemas) > 0 {
			return false
		}
	}

	if s.Enum != nil || s.Const != nil || len(s.AllOf) > 0 || len(s.AnyOf) > 0 || len(s.OneOf) > 0 || s.Not != nil {
		return false
	}

	return true
}
