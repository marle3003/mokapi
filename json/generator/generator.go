package generator

import (
	"math/rand"
	"time"
)

var g = &generator{
	rand: rand.New(rand.NewSource(time.Now().Unix())),
	tree: NewTree(),
}

type generator struct {
	rand *rand.Rand

	tree *Tree
}

func Seed(seed int64) {
	g.rand.Seed(seed)
}

func New(r *Request) (interface{}, error) {
	r.g = g
	r.context = map[string]interface{}{}
	return r.g.tree.Resolve(r)
}
