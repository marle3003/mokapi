package generator

import (
	"fmt"
	"mokapi/schema/json/schema"
	"strconv"
	"strings"

	"github.com/brianvoe/gofakeit/v6"
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
				{
					Name: "floor",
					Fake: fakeFloor,
					Children: []*Node{
						{
							Name: "door",
							Fake: fakeRoom,
						},
					},
				},
				{
					Name:       "room",
					Attributes: []string{"door", "apartment", "suite", "flat", "room"},
					Fake:       fakeRoom,
				},
			},
				personal[0].Children...,
			),
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
			Name:       "city",
			Attributes: []string{"city", "locality"},
			Fake:       fakeCity,
		},
		{
			Name:       "zip",
			Attributes: []string{"zip", "postcode", "postal"},
			Fake:       fakePostcode,
			Children: []*Node{
				{
					Name: "code",
					Fake: fakePostcode,
				},
			},
		},
		{
			Name:       "house",
			Attributes: []string{"house", "building"},
			Fake:       fakeHouseNumber,
		},
	}
}

func fakeStreet(r *Request) (any, error) {
	v := gofakeit.Street()
	r.Context.Values["street"] = v
	return v, nil
}

func fakeCity(r *Request) (any, error) {
	var v interface{}
	var err error
	s := r.Schema
	if s.IsAny() || s.IsString() {
		v = gofakeit.City()
	} else if s.IsInteger() {
		v, err = newPostCode(s)
	} else {
		return nil, NotSupported
	}
	if err != nil {
		return nil, err
	}
	r.Context.Values["city"] = v
	return v, nil
}

func fakePostcode(r *Request) (any, error) {
	return newPostCode(r.Schema)
}

func newPostCode(s *schema.Schema) (any, error) {
	if s == nil || s.IsAny() {
		s = &schema.Schema{Type: []string{"string"}}
	}
	if s.Pattern != "" {
		return nil, NotSupported
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
		return int(codeN), nil
	} else if s.IsNumber() {
		codeN, _ := strconv.ParseInt(code, 10, 32)
		return float64(codeN), nil
	}
	return code, nil
}

func fakeAddress(_ *Request) (interface{}, error) {
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

func fakeFloor(_ *Request) (any, error) {
	index := gofakeit.Number(0, len(floor)-1)
	return floor[index], nil
}

func fakeRoom(_ *Request) (any, error) {
	index := gofakeit.Number(0, len(room)-1)
	return room[index], nil
}

var (
	floor = []string{
		"1", "2", "3", "G", "LG", "UG", "B1", "M", "10", "PH", "R",
	}
	room = []string{
		"12A",
		"3B",
		"101",
		"7-14",
		"B2",
		"G5", // Ground floor unit 5
		"2F", // 2nd Floor
		"Unit 8",
		"Apt. 305",
		"Suite 12",
		"Flat 4",
		"PH1", // Penthouse 1
		"12B North",
		"Block C, 402",
	}
)
