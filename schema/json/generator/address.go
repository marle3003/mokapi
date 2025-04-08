package generator

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"mokapi/schema/json/schema"
	"strconv"
	"strings"
)

func addresses() []*Node {
	return []*Node{
		{
			Name: "address",
			Fake: fakeAddress,
			Children: append([]*Node{
				{
					Name: "co",
					Fake: fakePersonName,
				},
				{
					Name: "line1",
					Fake: fakePersonName,
				},
				{
					Name: "line2",
					Fake: fakeStreet,
				},
				{
					Name: "line3",
					Fake: func(r *Request) (interface{}, error) {
						return fmt.Sprintf("%v %v %v", gofakeit.City(), gofakeit.StateAbr(), gofakeit.Zip()), nil
					},
				},
			}, personal[0].Children...),
		},
		{
			Name: "co",
			Children: []*Node{
				{
					Name: "address",
					Fake: fakePersonName,
				},
			},
		},
		{
			Name: "street",
			Fake: fakeStreet,
		},
		{
			Name: "city",
			Fake: fakeCity,
		},
		{
			Name: "postcode",
			Fake: fakePostcode,
		},
		{
			Name: "zip",
			Fake: fakePostcode,
			Children: []*Node{
				{
					Name: "code",
					Fake: fakePostcode,
				},
			},
		},
		{
			Name: "house",
			Fake: fakeHouseNumber,
			Children: []*Node{
				{
					Name: "number",
					Fake: fakeHouseNumber,
				},
			},
		},
	}
}

func fakeStreet(r *Request) (any, error) {
	v := gofakeit.Street()
	r.ctx.store["street"] = v
	return v, nil
}

func fakeCity(r *Request) (any, error) {
	var v interface{}
	s := r.Schema
	if s.IsAny() || s.IsString() {
		v = gofakeit.City()
	} else if s.IsInteger() {
		v = newPostCode(s)
	} else {
		return nil, NotSupported
	}
	r.ctx.store["city"] = v
	return v, nil
}

func fakePostcode(r *Request) (any, error) {
	return newPostCode(r.Schema), nil
}

func newPostCode(s *schema.Schema) any {
	if s == nil || s.IsAny() {
		s = &schema.Schema{Type: []string{"string"}}
	}
	minLength := 4
	maxLength := 6
	if s.IsInteger() {
		if s.Minimum != nil {
			minLength = len(fmt.Sprintf("%v", *s.Minimum))
		}
		if s.Maximum != nil {
			maxLength = len(fmt.Sprintf("%v", *s.Maximum))
		}
	} else if s.IsString() {
		if s.MinLength != nil {
			minLength = *s.MinLength
		}
		if s.MaxLength != nil {
			maxLength = *s.MaxLength
		}
	}

	var n int
	if minLength == maxLength {
		n = minLength
	} else {
		n = gofakeit.Number(minLength, maxLength)
	}

	code := gofakeit.Numerify(strings.Repeat("#", n))
	if s.IsInteger() {
		codeN, _ := strconv.ParseInt(code, 10, 32)
		return int(codeN)
	} else if s.IsNumber() {
		codeN, _ := strconv.ParseInt(code, 10, 32)
		return float64(codeN)
	}
	return code
}

func fakeAddress(r *Request) (interface{}, error) {
	addr := gofakeit.Address()
	return map[string]interface{}{
		"address":   addr.Address,
		"street":    addr.Street,
		"city":      addr.City,
		"state":     addr.State,
		"zip":       addr.Zip,
		"country":   addr.Country,
		"latitude":  addr.Latitude,
		"longitude": addr.Longitude,
	}, nil
}

func fakeHouseNumber(r *Request) (any, error) {
	if r.Schema.IsNumber() || r.Schema.IsString() {
		return fakeIntegerWithRange(r.Schema, 1, 100)
	}
	return gofakeit.StreetNumber(), nil
}
