package generator

import (
	"mokapi/schema/json/parser"
	"strings"

	"github.com/brianvoe/gofakeit/v7"
)

func locations() []*Node {
	return []*Node{
		{
			Name:       "country",
			Attributes: []string{"country", "countryName"},
			Fake:       fakeCountry,
		},
		{
			Name:       "longitude",
			Attributes: []string{"longitude"},
			Fake:       fakeLongitude,
		},
		{
			Name:       "latitude",
			Attributes: []string{"latitude"},
			Fake:       fakeLatitude,
		},
	}
}

func fakeCountry(r *Request) (any, error) {
	s := r.Schema
	var v string

	if s != nil {
		max := -1
		if s.MaxLength != nil {
			max = *s.MaxLength
		}

		if max == 2 {
			country := gofakeit.CountryAbr()
			v = country
		} else if s.Pattern != "" {
			country := gofakeit.CountryAbr()
			p := parser.Parser{Schema: s}
			_, err := p.Parse(country)
			if err == nil {
				v = country
			} else {
				// try lower case
				country = strings.ToLower(country)
				_, err = p.Parse(country)
				if err == nil {
					v = country
				}
			}
		}
	}
	if v == "" {
		v = gofakeit.Country()
	}

	r.Context.Values["country"] = v
	return v, nil
}

func fakeLongitude(r *Request) (any, error) {
	v := gofakeit.Longitude()
	r.Context.Values["longitude"] = v
	return v, nil
}

func fakeLatitude(r *Request) (any, error) {
	v := gofakeit.Latitude()
	r.Context.Values["latitude"] = v
	return v, nil
}
