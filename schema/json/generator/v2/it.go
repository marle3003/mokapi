package v2

import (
	"crypto/sha1"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
)

func newItNodes() []*Node {
	return []*Node{
		newErrorNode(),
		newHashNode(),
	}
}

func newErrorNode() *Node {
	return &Node{Name: "error", Fake: fakeError}
}

func fakeError(r *Request) (interface{}, error) {
	return gofakeit.Error().Error(), nil
}

func newHashNode() *Node {
	return &Node{Name: "hash", Fake: fakeHash}
}

func fakeHash(_ *Request) (interface{}, error) {
	hash := sha1.New()
	s := gofakeit.SentenceSimple()
	b := hash.Sum([]byte(s))
	return fmt.Sprintf("%x", b), nil
}
