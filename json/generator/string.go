package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"mokapi/json/schema"
	"strings"
)

var (
	lowerChars   = "abcdefghijklmnopqrstuvwxyz"
	upperChars   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numericChars = "0123456789"
	specialChars = "!@#$%&*+-_=?:;,.|(){}<>"
	spaceChar    = " "
	allStr       = lowerChars + upperChars + numericChars + specialChars + spaceChar

	categories = []interface{}{0, 1, 2, 3, 4}
	weights    = []float32{5, 3, 0.5, 0.1, 0.1}
)

func Strings() *Tree {
	return &Tree{
		Name: "Strings",
		Nodes: []*Tree{
			StringFormat(),
			StringPattern(),
			Name(),
			StringNumber(),
			StringKey(),

			StringDescription(),
			String(),
		},
	}
}

func StringFormat() *Tree {
	return &Tree{
		Name: "Format",
		Test: func(r *Request) bool {
			s := r.LastSchema()
			return s.IsString() && len(s.Format) > 0
		},
		Fake: func(r *Request) (interface{}, error) {
			s := r.LastSchema()
			switch s.Format {
			case "date":
				return gofakeit.Date().Format("2006-01-02"), nil
			case "date-time":
				return gofakeit.Generate("{date}"), nil
			case "password":
				return gofakeit.Generate("{password}"), nil
			case "email":
				return gofakeit.Generate("{email}"), nil
			case "uuid":
				return gofakeit.Generate("{uuid}"), nil
			case "uri":
				return gofakeit.Generate("{url}"), nil
			case "hostname":
				return gofakeit.Generate("{domainname}"), nil
			case "ipv4":
				return gofakeit.Generate("{ipv4address}"), nil
			case "ipv6":
				return gofakeit.Generate("{ipv6address}"), nil
			default:
				return gofakeit.Generate(s.Format), nil
			}
		},
	}
}

func StringId() *Tree {
	return &Tree{
		Name: "StringId",
		Test: func(r *Request) bool {
			return r.Path.MatchLast(ComparerFunc(func(p *PathElement) bool {
				return (strings.ToLower(p.Name) == "id" || strings.HasSuffix(p.Name, "Id")) &&
					p.Schema.IsString() && !hasPattern(p.Schema) && !hasFormat(p.Schema)
			}))
		},
		Fake: func(r *Request) (interface{}, error) {
			return newId(r.LastSchema())
		},
	}
}

func StringNumber() *Tree {
	return &Tree{
		Name: "StringNumber",
		Test: func(r *Request) bool {
			return r.Path.MatchLast(ComparerFunc(func(p *PathElement) bool {
				return strings.HasSuffix(p.Name, "Number") &&
					p.Schema.IsString() && !hasPattern(p.Schema) && !hasFormat(p.Schema)
			}))
		},
		Fake: func(r *Request) (interface{}, error) {
			s := r.LastSchema()
			min := 11
			max := 11
			if s.MaxLength != nil {
				max = *s.MaxLength
			}
			if s.MinLength != nil {
				min = *s.MinLength
			} else if s.MaxLength != nil {
				min = 0
			}
			var n int
			if min == max {
				n = min
			} else {
				n = gofakeit.Number(min, max)
			}
			return gofakeit.Numerify(strings.Repeat("#", n)), nil
		},
	}
}

func StringKey() *Tree {
	return &Tree{
		Name: "StringKey",
		Test: func(r *Request) bool {
			last := r.Last()
			return (strings.ToLower(last.Name) == "key" || strings.HasSuffix(last.Name, "Key")) &&
				last.Schema.IsString() && !hasPattern(last.Schema) && !hasFormat(last.Schema)
		},
		Fake: func(r *Request) (interface{}, error) {
			return newId(r.LastSchema())
		},
	}
}

func StringEmail() *Tree {
	return &Tree{
		Name: "Email",
		Test: func(r *Request) bool {
			last := r.Last()
			return strings.ToLower(last.Name) == "email" &&
				(last.Schema.IsAnyString() || last.Schema.IsAny())
		},
		Fake: func(r *Request) (interface{}, error) {
			return gofakeit.Email(), nil
		},
	}
}

func StringDescription() *Tree {
	return &Tree{
		Name: "StringDescription",
		Test: func(r *Request) bool {
			last := r.Last()
			return (strings.ToLower(last.Name) == "description" || strings.HasSuffix(last.Name, "Description")) &&
				last.Schema.IsAnyString()
		},
		Fake: func(r *Request) (interface{}, error) {
			return gofakeit.Sentence(15), nil
		},
	}
}

func String() *Tree {
	return &Tree{
		Name: "String",
		Test: func(r *Request) bool {
			return r.LastSchema().IsString()
		},
		Fake: func(r *Request) (interface{}, error) {
			s := r.LastSchema()
			minLength := 0
			maxLength := 15

			if s != nil && s.MinLength != nil {
				minLength = *s.MinLength
			}
			if s != nil && s.MaxLength != nil {
				maxLength = *s.MaxLength
			}

			length := gofakeit.IntRange(minLength, maxLength)
			result := make([]rune, length)
			for i := 0; i < length; i++ {
				c, _ := gofakeit.Weighted(categories, weights)

				switch c {
				case 0:
					n := gofakeit.IntRange(0, len(lowerChars)-1)
					result[i] = rune(lowerChars[n])
				case 1:
					n := gofakeit.IntRange(0, len(upperChars)-1)
					result[i] = rune(upperChars[n])
				case 2:
					n := gofakeit.IntRange(0, len(numericChars)-1)
					result[i] = rune(numericChars[n])
				case 3:
					result[i] = ' '
				case 4:
					n := gofakeit.IntRange(0, len(specialChars)-1)
					result[i] = rune(specialChars[n])
				}
			}
			return string(result), nil
		},
	}
}

func hasPattern(r *schema.Ref) bool {
	return r != nil && r.Value != nil && r.Value.Pattern != ""
}

func hasFormat(r *schema.Ref) bool {
	return r != nil && r.Value != nil && r.Value.Format != ""
}

func newId(s *schema.Schema) (string, error) {
	min := 37
	max := 37
	if s.MaxLength != nil {
		max = *s.MaxLength
	}
	if s.MinLength != nil {
		min = *s.MinLength
	} else if s.MaxLength != nil {
		min = max
	}

	if min <= 37 && max >= 37 {
		return gofakeit.UUID(), nil
	}
	n := gofakeit.Number(min, max)
	return gofakeit.Numerify(strings.Repeat("#", n)), nil
}
