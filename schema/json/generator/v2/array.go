package v2

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/jinzhu/inflection"
	"mokapi/schema/json/schema"
	"reflect"
)

func (r *resolver) resolveArray(req *Request) (*faker, error) {
	s := req.Schema
	last := req.Path[len(req.Path)-1]
	path := req.Path
	singular := inflection.Singular(last)
	if singular != last {
		path = append(req.Path, singular)
	}
	if s.Items != nil && s.Items.Ref != "" {
		path = append(path, getPathFromRef(s.Items.Ref))
	}
	item, err := r.resolve(req.With(path, s.Items), true)
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
