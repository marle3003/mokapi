package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"mokapi/schema/json/schema"
	"strings"
	"time"
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
				return date().Format("2006-01-02"), nil
			case "date-time":
				return date().Format(time.RFC3339), nil
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
			minLength := 11
			maxLength := 11
			if s.MaxLength != nil {
				maxLength = *s.MaxLength
			}
			if s.MinLength != nil {
				minLength = *s.MinLength
			} else if s.MaxLength != nil {
				minLength = 0
			}
			var n int
			if minLength == maxLength {
				n = minLength
			} else {
				n = gofakeit.Number(minLength, maxLength)
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
			if last == nil {
				return false
			}
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
			if last == nil {
				return false
			}
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
			if last == nil {
				return false
			}
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

func hasPattern(s *schema.Schema) bool {
	return s != nil && s.Pattern != ""
}

func hasFormat(s *schema.Schema) bool {
	return s != nil && s.Format != ""
}

func newId(s *schema.Schema) (string, error) {
	minLength := 37
	maxLength := 37
	if s.MaxLength != nil {
		maxLength = *s.MaxLength
	}
	if s.MinLength != nil {
		minLength = *s.MinLength
	} else if s.MaxLength != nil {
		minLength = maxLength
	}

	if minLength <= 37 && maxLength >= 37 {
		return gofakeit.UUID(), nil
	}
	n := gofakeit.Number(minLength, maxLength)
	return gofakeit.Numerify(strings.Repeat("#", n)), nil
}

var maxDayInMonth = []int{
	31, // january
	28, // february
	31, // march
	30, // april
	31, // may
	30, // june
	31, // july
	31, // august
	30, // september
	31, // october
	30, // november
	31, // december
}

// gofakeit uses year range between 1900 and now
func date() time.Time {
	year := gofakeit.Number(1970, 2040)
	month := gofakeit.Number(1, 12)
	day := gofakeit.Number(1, maxDayInMonth[month-1])
	hour := gofakeit.Number(0, 23)
	minute := gofakeit.Number(0, 59)
	second := gofakeit.Number(0, 59)
	nanosecond := gofakeit.Number(0, 999999999)
	return time.Date(year, time.Month(month), day, hour, minute, second, nanosecond, time.UTC)
}
