package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"strings"
)

func Location() *Tree {
	return &Tree{
		Name: "Location",
		Nodes: []*Tree{
			City(),
			Country(),
			CountryName(),
			Longitude(),
			Latitude(),
			Address(),
		},
	}
}

func Country() *Tree {
	return &Tree{
		Name: "Country",
		Test: func(r *Request) bool {
			last := r.Last()
			if last.Name == "country" &&
				(last.Schema.IsAny() || last.Schema.IsString()) {
				if hasPattern(last.Schema) {
					p := last.Schema.Value.Pattern
					return p == "[A-Z]{2}" || p == "[a-z]{2}"
				}
				if hasFormat(last.Schema) {
					return false
				}
				if last.Schema.IsString() {
					return last.Schema.Value.MaxLength == nil || *last.Schema.Value.MaxLength >= 56 &&
						last.Schema.Value.MinLength == nil || *last.Schema.Value.MinLength <= 4
				}
			}
			return false
		},
		Fake: func(r *Request) (interface{}, error) {
			s := r.LastSchema()
			max := 2
			if s != nil && s.MaxLength != nil {
				max = *s.MaxLength
			}

			if max == 2 {
				country := gofakeit.CountryAbr()
				if s != nil && s.Pattern == "[a-z]{2}" {
					return strings.ToLower(country), nil
				}
				return country, nil
			}

			return gofakeit.Country(), nil
		},
	}
}

func CountryName() *Tree {
	return &Tree{
		Name: "CountryName",
		Test: func(r *Request) bool {
			last := r.Last()
			return last.Name == "countryName" &&
				(last.Schema.IsAny() || last.Schema.IsAnyString())
		},
		Fake: func(r *Request) (interface{}, error) {
			return gofakeit.Country(), nil
		},
	}
}

func Longitude() *Tree {
	return &Tree{
		Name: "Longitude",
		Test: func(r *Request) bool {
			last := r.Last()
			name := strings.ToLower(last.Name)
			return name == "longitude" && (last.Schema.IsNumber() || last.Schema.IsAny())
		},
		Fake: func(r *Request) (interface{}, error) {
			return gofakeit.Longitude(), nil
		},
	}
}

func Latitude() *Tree {
	return &Tree{
		Name: "Latitude",
		Test: func(r *Request) bool {
			last := r.Last()
			name := strings.ToLower(last.Name)
			return name == "latitude" && (last.Schema.IsNumber() || last.Schema.IsAny())
		},
		Fake: func(r *Request) (interface{}, error) {
			return gofakeit.Latitude(), nil
		},
	}
}
