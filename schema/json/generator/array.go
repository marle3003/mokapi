package generator

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/jinzhu/inflection"
	"mokapi/schema/json/schema"
	"reflect"
)

func (r *resolver) resolveArray(req *Request) (*faker, error) {
	s := req.Schema
	path := req.Path
	if len(path) > 0 {
		last := req.Path[len(req.Path)-1]
		singular := inflection.Singular(last)
		if singular != last {
			path = append(req.Path, singular)
		}
	}
	var item *faker
	var err error
	if s.Items != nil {
		if s.Items.Ref != "" {
			path = append(path, getPathFromRef(s.Items.Ref))
		}
		if len(s.Items.Enum) > 0 && s.UniqueItems {
			index := gofakeit.Number(0, len(s.Items.Enum)-1)
			item = newFaker(func() (any, error) {
				index = (index + 1) % len(s.Items.Enum)
				return s.Items.Enum[index], nil
			})
		}
	}
	if item == nil {
		req.examples = examplesFromRequest(req)
		item, err = r.resolve(req.With(path, s.Items, itemsFromExample(req)), true)
	}
	if err != nil {
		return nil, err
	}
	return newFaker(func() (interface{}, error) {
		return fakeArray(req, item)
	}), nil
}

func fakeArray(r *Request, fakeItem *faker) (interface{}, error) {
	s := r.Schema
	if s == nil {
		s = &schema.Schema{}
	}

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

	arr := make([]interface{}, 0, length)
	for i := 0; i < length; i++ {
		r.ctx.Snapshot()
		var v interface{}
		var err error
		if s.UniqueItems {
			v, err = nextUnique(arr, fakeItem.fake)
		} else {
			v, err = fakeItem.fake()
		}
		if err != nil {
			return nil, fmt.Errorf("%v: %v", err, s)
		}
		arr = append(arr, v)
		r.ctx.Restore()
	}

	if s.ShuffleItems {
		r.g.rand.Shuffle(len(arr), func(i, j int) { arr[i], arr[j] = arr[j], arr[i] })
	}

	return arr, nil
}

func nextUnique(arr []interface{}, fakeItem func() (interface{}, error)) (interface{}, error) {
	for i := 0; i < 10; i++ {
		v, err := fakeItem()
		if err != nil {
			return nil, err
		}
		if !contains(arr, v) {
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

func itemsFromExample(r *Request) []any {
	var result []any
	for _, e := range r.examples {
		if arr, ok := e.([]any); ok {
			result = append(result, arr...)
		}
	}
	return result
}

func itemExamplesFromRequest(r *Request, item *schema.Schema) []any {
	var result []any
	if r.examples != nil {
		for _, e := range r.examples {
			if arr, ok := e.([]any); ok {
				for _, i := range arr {
					result = append(result, i)
				}
			}
		}
	}

	for _, v := range examples(r.Schema) {
		if arr, isArray := v.([]any); isArray {
			for _, i := range arr {
				result = append(result, i)
			}
		}
	}

	result = append(result, examples(item)...)

	return result
}
