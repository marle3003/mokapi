package generator

import (
	"math/rand"
	"time"
)

var g = &generator{rand: rand.New(rand.NewSource(time.Now().Unix()))}

type generator struct {
	rand *rand.Rand
}

func Seed(seed int64) {
	g.rand.Seed(seed)
}
