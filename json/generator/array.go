package generator

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"reflect"
)

type ArrayOptions struct {
	MinItems    *int
	MaxItems    *int
	Shuffle     bool
	UniqueItems bool
	Nullable    bool
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
