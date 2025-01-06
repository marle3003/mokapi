package generator

import (
	"fmt"
	"mokapi/schema/json/parser"
	"mokapi/schema/json/schema"
)

func AllOf() *Tree {
	p := parser.Parser{}
	validate := func(data interface{}, schemas []*schema.Ref) error {
		for _, s := range schemas {
			if _, err := p.ParseWith(data, s); err != nil {
				return err
			}
		}
		return nil
	}

	return &Tree{
		Name: "AllOf",
		Test: func(r *Request) bool {
			s := r.LastSchema()
			return s != nil && s.AllOf != nil && len(s.AllOf) > 0
		},
		Fake: func(r *Request) (interface{}, error) {
			s := r.LastSchema()

			var result interface{}
			var worked []*schema.Ref
			sharedType, err := getShareTypes(s.AllOf)
			if err != nil {
				return nil, fmt.Errorf("generate random data for schema failed: %v: %v", s.String(), err)
			}

			for _, one := range s.AllOf {
				if one == nil || one.Value == nil {
					continue
				}

				var err error
				copySchema := *one
				copySchema.Value.Type = sharedType

				// We can skip this schema if current result is valid for it but
				// for an object all properties are expected even if the existing result is already valid
				obj, isObject := result.(map[string]interface{})
				if result != nil && !isObject {
					if _, err = p.ParseWith(result, &copySchema); err == nil {
						continue
					}
				}

				var next interface{}
				next, err = r.g.tree.Resolve(r.With(UsePathElement("", &copySchema)))
				if err != nil {
					return nil, fmt.Errorf("generate random data for schema failed: %v: %v", copySchema.String(), err)
				}

				if isObject {
					var nextObj map[string]interface{}
					nextObj, isObject = next.(map[string]interface{})
					if !isObject {
						return nil, fmt.Errorf("generate random data for schema failed: types in allOf does not match")
					}

					for it := copySchema.Value.Properties.Iter(); it.Next(); {
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
		},
	}
}

func getShareTypes(sets []*schema.Ref) (schema.Types, error) {
	m := map[string]struct{}{}

	countNoSchemaDefined := 0
	for _, set := range sets {
		if set == nil || set.Value == nil {
			countNoSchemaDefined++
			continue
		}
		if set.Value.Type == nil {
			// JSON schema does not require a type
			continue
		}

		if len(m) == 0 {
			for _, t := range set.Value.Type {
				m[t] = struct{}{}
			}
		} else {
			for k := range m {
				if !set.Value.Type.Includes(k) {
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
	return result, nil
}
