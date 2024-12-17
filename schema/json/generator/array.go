package generator

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/jinzhu/inflection"
	"github.com/pkg/errors"
	"reflect"
)

type ArrayOptions struct {
	MinItems    *int
	MaxItems    *int
	Shuffle     bool
	UniqueItems bool
	Nullable    bool
}

func Array() *Tree {
	return &Tree{
		Name: "Array",
		Test: func(r *Request) bool {
			return r.LastSchema().IsArray()
		},
		Fake: func(r *Request) (interface{}, error) {
			s := r.LastSchema()
			maxItems := 5
			if s.MaxItems != nil {
				maxItems = *s.MaxItems
			}
			minItems := 0
			if s.MinItems != nil {
				minItems = *s.MinItems
			}
			length := minItems
			if maxItems-minItems > 0 {
				length = gofakeit.Number(minItems, maxItems)
			}

			name := RefName(s.Items)
			if name == "" {
				name = inflection.Singular(r.LastName())
			}

			elem := r.With(UsePathElement(name, s.Items))
			arr := make([]interface{}, 0, length)
			for i := 0; i < length; i++ {
				var v interface{}
				var err error
				if s.UniqueItems {
					v, err = nextUnique(arr, elem)
				} else {
					v, err = r.g.tree.Resolve(elem)
				}
				if err != nil {
					var rec *RecursionGuard
					if errors.As(err, &rec) {
						continue
					}
					return nil, fmt.Errorf("%v: %v", err, s)
				}
				arr = append(arr, v)
			}

			if s.ShuffleItems {
				r.g.rand.Shuffle(len(arr), func(i, j int) { arr[i], arr[j] = arr[j], arr[i] })
			}

			return arr, nil
		},
	}
}

func nextUnique(arr []interface{}, r *Request) (interface{}, error) {
	s := r.LastSchema()
	if s.Enum != nil {
		n := gofakeit.Number(0, len(s.Enum)-1)
		for i := 0; i < len(s.Enum); i++ {
			index := (n + i) % len(s.Enum)
			v := s.Enum[index]
			if !contains(arr, v) {
				return v, nil
			}
		}
	} else {
		for i := 0; i < 10; i++ {
			v, err := r.g.tree.Resolve(r)
			if err != nil {
				return nil, err
			}
			if !contains(arr, v) {
				return v, nil
			}
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
