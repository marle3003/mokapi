package v2

import (
	"github.com/brianvoe/gofakeit/v6"
	"strings"
)

func locations() []*Node {
	return []*Node{
		{
			Name: "country",
			Fake: fakeCountry,
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
	if s != nil {
		max := -1
		if s.MaxLength != nil {
			max = *s.MaxLength
		}

		if max == 2 {
			country := gofakeit.CountryAbr()
			if s.Pattern == "[a-z]{2}" {
				return strings.ToLower(country), nil
			}
			return country, nil
		}
	}

	return gofakeit.Country(), nil
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
