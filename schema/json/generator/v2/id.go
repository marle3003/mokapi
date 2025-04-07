package v2

import (
	"github.com/brianvoe/gofakeit/v6"
	"strings"
)

func newIdNode() *Node {
	return &Node{Name: "id", Fake: fakeId}
}

func fakeId(r *Request) (interface{}, error) {
	s := r.Schema
	if s.IsString() {
		minLength := 37
		maxLength := 37
		if s.MaxLength != nil {
			maxLength = *s.MaxLength
		}
		if s.MinLength != nil {
			minLength = *s.MinLength
		} else if s.MaxLength != nil {
			minLength = maxLength
		}

		if minLength <= 37 && maxLength >= 37 {
			return gofakeit.UUID(), nil
		}
		n := gofakeit.Number(minLength, maxLength)
		return gofakeit.Numerify(strings.Repeat("#", n)), nil
	} else if s.IsInteger() || s.IsAny() {
		return fakeIntegerWithRange(s, 1, 100000)
	}

	return nil, NotSupported
}
