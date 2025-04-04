package v2

import (
	"github.com/brianvoe/gofakeit/v6"
	"strings"
)

func newNumberNode() *Node {
	return &Node{Name: "number", Fake: fakeNumber}
}

func fakeNumber(r *Request) (interface{}, error) {
	s := r.Schema
	minLength := 11
	maxLength := 11
	if s != nil && s.MaxLength != nil {
		maxLength = *s.MaxLength
	}
	if s != nil && s.MinLength != nil {
		minLength = *s.MinLength
	} else if s != nil && s.MaxLength != nil {
		minLength = 0
	}
	var n int
	if minLength == maxLength {
		n = minLength
	} else {
		n = gofakeit.Number(minLength, maxLength)
	}
	return gofakeit.Numerify(strings.Repeat("#", n)), nil
}
