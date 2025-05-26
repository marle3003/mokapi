package generator

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"mokapi/schema/json/schema"
	"mokapi/sortedmap"
)

func fakeBySchemaNode() *Node {
	return &Node{
		Name:       "fakeBySchema",
		Attributes: []string{""},
		Weight:     0,
		Fake: func(r *Request) (any, error) {
			if len(r.Path) == 0 {
				return fakeBySchema(r)
			}
			return nil, NotSupported
		},
	}
}

func fakeBySchema(r *Request) (interface{}, error) {
	if fake, ok := applyConstraints(r); ok {
		return fake()
	}
	if r.examples != nil {
		if v, err := selectExample(r); err == nil {
			return v, nil
		}
	}

	s := r.Schema
	var t schema.Types
	if s != nil {
		t = s.Type
	}

	if s != nil && len(s.Type) > 1 {
		t = s.Type
		if s.IsNullable() {
			n := gofakeit.Float32Range(0, 1)
			if n > 0.05 {
				t = removeNull(s.Type)
			}
		}

		index := gofakeit.Number(0, len(t)-1)
		t = schema.Types{t[index]}
		c := *s
		c.Type = t
		s = &c
		r.Schema = s
	}

	switch {
	case t.IsString():
		return fakeString(r)
	case t.IsObject():
		return fakeObject(r)
	case t.IsArray():
		items := func() (interface{}, error) {
			return fakeBySchema(r.WithSchema(s.Items))
		}
		return fakeArray(r, newFaker(items))
	case t.IsBool():
		return gofakeit.Bool(), nil
	case t.IsNumber():
		return fakeNumber(r)
	case t.IsInteger():
		return fakeInteger(r.Schema)
	case t.IsNullable():
		return nil, nil
	case t.IsNullable():
		return nil, nil
	case s != nil && len(s.Type) > 0:
		return nil, fmt.Errorf("unsupported schema: %s", s)
	}

	i, _ := gofakeit.Weighted(types, weightTypes)
	s = &schema.Schema{Type: schema.Types{i.(string)}}
	return fakeBySchema(r.WithSchema(s))
}

func fakeObject(r *Request) (interface{}, error) {
	s := r.Schema
	if s.Properties == nil {
		s.Properties = &schema.Schemas{LinkedHashMap: sortedmap.LinkedHashMap[string, *schema.Schema]{}}
		length := numProperties(0, 10, s)

		if length == 0 {
			return map[string]interface{}{}, nil
		}

		for i := 0; i < length; i++ {
			name := fakeDictionaryKey()
			s.Properties.Set(name, nil)
		}
	}

	m := map[string]any{}
	for it := s.Properties.Iter(); it.Next(); {
		v, err := New(r.With([]string{it.Key()}, it.Value(), nil))
		if err != nil {
			return nil, err
		}
		m[it.Key()] = v
	}
	return m, nil
}

func removeNull(slice schema.Types) schema.Types {
	for i, v := range slice {
		if v == "null" {
			slice = append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

func selectExample(r *Request) (any, error) {
	items := r.examples
	start := gofakeit.Number(0, len(items)-1)
	for i := 0; i < len(items); i++ {
		index := (start + i) % len(items)
		item := items[index]

		if v, err := validate(item, r); err == nil {
			return v, nil
		}
	}
	return nil, NoMatchFound
}
