package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"strings"
)

func Colors() *Tree {
	return &Tree{
		Name: "Colors",
		Nodes: []*Tree{
			HexColor(),
			RGBColor(),
			ColorName(),
		},
	}
}

func ColorName() *Tree {
	return &Tree{
		Name: "ColorName",
		Test: func(r *Request) bool {
			last := r.Last()
			return strings.ToLower(last.Name) == "color" && last.Schema.IsAnyString()
		},
		Fake: func(r *Request) (interface{}, error) {
			return gofakeit.Color(), nil
		},
	}
}

func HexColor() *Tree {
	return &Tree{
		Name: "HEX-Color",
		Test: func(r *Request) bool {
			last := r.Last()
			if strings.ToLower(last.Name) != "color" || !last.Schema.IsString() {
				return false
			}
			s := r.LastSchema()
			if s.MaxLength != nil && *s.MaxLength == 7 {
				return true
			}
			return false
		},
		Fake: func(r *Request) (interface{}, error) {
			return gofakeit.HexColor(), nil
		},
	}
}

func RGBColor() *Tree {
	return &Tree{
		Name: "RGB-Color",
		Test: func(r *Request) bool {
			last := r.Last()
			if strings.ToLower(last.Name) != "color" || !last.Schema.IsArray() {
				return false
			}
			s := r.LastSchema()
			if s.Items.IsAny() || (s.Items.IsInteger() && s.Items.Value.MaxItems == nil || *s.Items.Value.MaxItems == 3) {
				return true
			}
			return false
		},
		Fake: func(r *Request) (interface{}, error) {
			return gofakeit.RGBColor(), nil
		},
	}
}
