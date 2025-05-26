package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"math/rand"
	"time"
)

var types = []any{"string", "number", "integer", "boolean", "array", "object", "null"}
var weightTypes = []float32{1, 1, 1, 1, 0.2, 0.2, 0.05}

type generator struct {
	rand *rand.Rand

	root *Node
}

var g *generator

func init() {
	g = &generator{
		rand: rand.New(rand.NewSource(time.Now().Unix())),
		root: buildTree(),
	}
}

func New(r *Request) (interface{}, error) {
	r.g = g
	if r.Context == nil {
		r.Context = newContext()
	}
	r.examples = examplesFromRequest(r)
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
