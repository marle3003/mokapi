package generator

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"mokapi/json/schema"
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
		Name:  "Array",
		nodes: nil,
		Test: func(r *Request) bool {
			return r.Schema.IsArray()
		},
		Fake: func(r *Request) (interface{}, error) {
			maxItems := 5
			if r.Schema.MaxItems != nil {
				maxItems = *r.Schema.MaxItems
			}
			minItems := 0
			if r.Schema.MinItems != nil {
				minItems = *r.Schema.MinItems
			}
			length := minItems
			if maxItems-minItems > 0 {
				length = gofakeit.Number(minItems, maxItems)
			}

			var items *schema.Schema
			if r.Schema.Items != nil {
				items = r.Schema.Items.Value
			}

			elem := r.With(Schema(items))
			arr := make([]interface{}, length)
			for i := range arr {
				var v interface{}
				var err error
				if r.Schema.UniqueItems {
					v, err = nextUnique(arr, elem)
				} else {
					v, err = r.g.tree.Resolve(elem)
				}
				if err != nil {
					return nil, fmt.Errorf("%v: %v", err, r.Schema)
				}
				arr[i] = v
			}

			if r.Schema.ShuffleItems {
				r.g.rand.Shuffle(len(arr), func(i, j int) { arr[i], arr[j] = arr[j], arr[i] })
			}

			return arr, nil
		},
	}
}

func nextUnique(arr []interface{}, r *Request) (interface{}, error) {
	if r.Schema.Enum != nil {
		n := gofakeit.Number(0, len(r.Schema.Enum)-1)
		for i := 0; i < len(r.Schema.Enum); i++ {
			index := (n + i) % len(r.Schema.Enum)
			v := r.Schema.Enum[index]
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

func NewArray(opt ArrayOptions, gen func() (interface{}, error)) (r []interface{}, err error) {
	if opt.Nullable {
		n := gofakeit.Float32Range(0, 1)
		if n < 0.05 {
			return nil, nil
		}
	}

	maxItems := 5
	if opt.MaxItems != nil {
		maxItems = *opt.MaxItems
	}
	minItems := 0
	if opt.MinItems != nil {
		minItems = *opt.MinItems
	}

	if opt.Shuffle {
		defer func() {
			g.rand.Shuffle(len(r), func(i, j int) { r[i], r[j] = r[j], r[i] })
		}()
	}

	length := minItems
	if maxItems-minItems > 0 {
		length = gofakeit.Number(minItems, maxItems)
	}
	r = make([]interface{}, length)

	for i := range r {
		if opt.UniqueItems {
			r[i], err = getUnique(r, gen)
			if err != nil {
				return nil, err
			}
		} else {
			r[i], err = gen()
			if err != nil {
				return
			}
		}
	}
	return r, nil
}

func getUnique(s []interface{}, gen func() (interface{}, error)) (interface{}, error) {
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
