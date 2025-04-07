package generator

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"math/rand"
	"mokapi/schema/json/schema"
	"time"
)

var types = []any{"string", "number", "integer", "boolean", "array", "object", "null"}
var weightTypes = []float32{1, 1, 1, 1, 0.2, 0.2, 0.05}

type generator struct {
	rand *rand.Rand

	root *Node
}

var g = &generator{
	rand: rand.New(rand.NewSource(time.Now().Unix())),
	root: buildTree(),
}

func New(r *Request) (interface{}, error) {
	r.g = g
	if r.ctx == nil {
		r.ctx = newContext()
	}
	f, err := resolve(r, true)
	if err != nil {
		return nil, err
	}
	return f.fake()
}

func Seed(seed int64) {
	gofakeit.Seed(seed)
	g.rand.Seed(seed)
}

func fakeBySchema(r *Request) (interface{}, error) {
	if fake, ok := applyConstraints(r); ok {
		return fake()
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

func removeNull(slice schema.Types) schema.Types {
	for i, v := range slice {
		if v == "null" {
			slice = append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}
