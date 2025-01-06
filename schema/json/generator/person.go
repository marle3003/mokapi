package generator

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"strings"
)

func Personal() *Tree {
	return &Tree{
		Name: "Personal",
		Nodes: []*Tree{
			Person(),
			Phone(),
			Contact(),
			FirstName(),
			LastName(),
			Gender(),
			Language(),
			PersonAny(),
		},
	}
}

func Person() *Tree {
	return &Tree{
		Name: "Person",
		Test: func(r *Request) bool {
			return r.Path.MatchLast(NameIgnoreCase("person", "persons"), Any())

		},
		Nodes: []*Tree{
			PersonName(),
		},
	}
}

func PersonName() *Tree {
	return &Tree{
		Name: "PersonName",
		Test: func(r *Request) bool {
			name := strings.ToLower(r.LastName())
			return name == "name" || name == "fullname"
		},
		Fake: func(r *Request) (interface{}, error) {
			return gofakeit.Name(), nil
		},
	}
}

func FirstName() *Tree {
	return &Tree{
		Name: "FirstName",
		Test: func(r *Request) bool {
			return strings.ToLower(r.LastName()) == "firstname"
		},
		Fake: func(r *Request) (interface{}, error) {
			return gofakeit.FirstName(), nil
		},
	}
}

func LastName() *Tree {
	return &Tree{
		Name: "LastName",
		Test: func(r *Request) bool {
			return strings.ToLower(r.LastName()) == "lastname"
		},
		Fake: func(r *Request) (interface{}, error) {
			return gofakeit.LastName(), nil
		},
	}
}

func Gender() *Tree {
	return &Tree{
		Name: "PersonGender",
		Test: func(r *Request) bool {
			name := r.LastName()
			return name == "gender" || name == "sex"
		},
		Fake: func(r *Request) (interface{}, error) {
			return gofakeit.Gender(), nil
		},
	}
}

func Phone() *Tree {
	return &Tree{
		Name: "Phone",
		Test: func(r *Request) bool {
			last := r.Last()
			if last == nil {
				return false
			}
			name := strings.ToLower(last.Name)
			return (name == "phone" || name == "phonenumber" || strings.HasSuffix(last.Name, "Phone")) &&
				(last.Schema.IsString() || last.Schema.IsAny())
		},
		Fake: func(r *Request) (interface{}, error) {
			countryCode := gofakeit.IntRange(1, 999)
			countryCodeLen := len(fmt.Sprintf("%v", countryCode))
			max := 15 - countryCodeLen
			min := 4
			s := r.LastSchema()
			if s != nil && s.MinLength != nil {
				min = *s.MinLength - countryCodeLen - 1
			}
			if s != nil && s.MaxLength != nil {
				max = *s.MaxLength - countryCodeLen - 1
			}
			nationalCodeLen := gofakeit.IntRange(min, max)
			return fmt.Sprintf("+%v%v", countryCode, gofakeit.Numerify(strings.Repeat("#", nationalCodeLen))), nil
		},
	}
}

func Contact() *Tree {
	return &Tree{
		Name: "Contact",
		Test: func(r *Request) bool {
			last := r.Last()
			if last == nil {
				return false
			}
			if !last.Schema.IsObject() && !last.Schema.IsAny() {
				return false
			}
			return last.Name == "contact"
		},
		Fake: func(r *Request) (interface{}, error) {
			phone, err := r.g.tree.Resolve(r.With(UsePathElement("phone", nil)))
			if err != nil {
				return nil, err
			}
			email, err := r.g.tree.Resolve(r.With(UsePathElement("email", nil)))
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"phone": phone,
				"email": email,
			}, nil
		},
	}
}

func PersonAny() *Tree {
	return &Tree{
		Name: "AnyPerson",
		Test: func(r *Request) bool {
			last := r.Last()
			if last == nil {
				return false
			}
			return last.Name == "person" && last.Schema.IsAny()
		},
		Fake: func(r *Request) (interface{}, error) {
			return map[string]interface{}{
				"firstname": gofakeit.FirstName(),
				"lastname":  gofakeit.LastName(),
				"gender":    gofakeit.Gender(),
				"email":     gofakeit.Email(),
			}, nil
		},
	}
}
