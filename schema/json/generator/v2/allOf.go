package v2

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"mokapi/schema/json/parser"
	"mokapi/schema/json/schema"
	"slices"
)

func (r *resolver) allOf(req *Request) (*faker, error) {
	s := req.Schema
	p := parser.Parser{}
	validate := func(data interface{}, schemas []*schema.Schema) error {
		for _, s := range schemas {
			if _, err := p.ParseWith(data, s); err != nil {
				return err
			}
		}
		return nil
	}

	sharedType, err := getShareTypes(s.AllOf)
	if err != nil {
		return nil, fmt.Errorf("generate random data for schema failed: %v: %v", s.String(), err)
	}

	f := func() (any, error) {
		var result interface{}
		var worked []*schema.Schema
		index := gofakeit.Number(0, len(sharedType)-1)
		selectedType := schema.Types{sharedType[index]}

		for _, one := range s.AllOf {
			if one == nil {
				continue
			}

			copySchema := *one
			copySchema.Type = selectedType

			// We could skip this schema if current result is valid for it but
			// for an object all properties are expected even if the existing result is already valid
			obj, isObject := result.(map[string]interface{})
			if result != nil && !isObject {
				if _, err = p.ParseWith(result, &copySchema); err == nil {
					continue
				}
			}

			var f *faker
			var next interface{}
			f, err = r.resolve(req.WithSchema(&copySchema), true)
			if err != nil {
				return nil, fmt.Errorf("generate random data for schema failed: %v: %v", copySchema.String(), err)
			}
			next, err = f.fake()
			if err != nil {
				return nil, fmt.Errorf("generate random data for schema failed: %v: %v", copySchema.String(), err)
			}

			if isObject {
				var nextObj map[string]interface{}
				nextObj, isObject = next.(map[string]interface{})
				if !isObject {
					return nil, fmt.Errorf("generate random data for schema failed: types in allOf does not match")
				}

				for it := copySchema.Properties.Iter(); it.Next(); {
					if v, found := obj[it.Key()]; found {
						if _, err = p.ParseWith(v, it.Value()); err != nil {
							obj[it.Key()] = nextObj[it.Key()]
						}
					} else {
						obj[it.Key()] = nextObj[it.Key()]
					}
				}
			} else {
				result = next
			}

			if err = validate(result, worked); err != nil {
				return nil, err
			}

			worked = append(worked, &copySchema)
		}
		return result, nil
	}
	return newFaker(f), nil
}

func getShareTypes(sets []*schema.Schema) (schema.Types, error) {
	m := map[string]struct{}{}

	countNoSchemaDefined := 0
	for _, set := range sets {
		if set == nil {
			countNoSchemaDefined++
			continue
		}
		if set.Type == nil {
			// JSON schema does not require a type
			continue
		}

		if len(m) == 0 {
			for _, t := range set.Type {
				m[t] = struct{}{}
			}
		} else {
			for k := range m {
				if !set.Type.Includes(k) {
					delete(m, k)
				}
			}
		}
	}

	if len(sets) == countNoSchemaDefined {
		return nil, nil
	}
	if len(m) == 0 {
		return nil, fmt.Errorf("no shared types found")
	}

	var result schema.Types
	for k := range m {
		result = append(result, k)
	}
	// sort to get predictable test results
	slices.Sort(result)
	return result, nil
}
