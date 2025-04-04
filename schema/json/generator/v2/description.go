package v2

import (
	"github.com/brianvoe/gofakeit/v6"
	"math"
)

func newDescriptionNode() *Node {
	return &Node{Name: "description", Fake: fakeDescription}
}

func fakeDescription(r *Request) (interface{}, error) {
	wordCount := 15
	if r.Schema != nil {
		avgWordLength := 5
		minLength := 0
		maxLength := 150
		if r.Schema.MinLength != nil {
			minLength = *r.Schema.MinLength
		}
		if r.Schema.MaxLength != nil {
			maxLength = *r.Schema.MaxLength
		}

		maxWords := int(math.Floor(float64(maxLength) / float64(avgWordLength)))

		for ; maxWords > 0; maxWords-- {
			v := gofakeit.Sentence(maxWords)
			if len(v) >= minLength && len(v) <= maxLength {
				return v, nil
			}
		}

		return nil, NotSupported
	}
	return gofakeit.Sentence(wordCount), nil
}
