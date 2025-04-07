package v2

import (
	"github.com/brianvoe/gofakeit/v6"
	"mokapi/schema/json/parser"
	"strings"
)

func locations() []*Node {
	return []*Node{
		{
			Name: "country",
			Fake: fakeCountry,
			Children: []*Node{
				{
					Name: "name",
					Fake: fakeCountry,
				},
			},
		},
		{
			Name: "longitude",
			Fake: fakeLongitude,
		},
		{
			Name: "latitude",
			Fake: fakeLatitude,
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

	r.ctx.store["country"] = v
	return v, nil
}

func fakeLongitude(r *Request) (any, error) {
	v := gofakeit.Longitude()
	r.ctx.store["longitude"] = v
	return v, nil
}

func fakeLatitude(r *Request) (any, error) {
	v := gofakeit.Latitude()
	r.ctx.store["latitude"] = v
	return v, nil
}
