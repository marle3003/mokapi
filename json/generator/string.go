package generator

import (
	"crypto/sha1"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"mokapi/json/schema"
	"strings"
)

const (
	defaultMaxStringLength = 15

	lowerChars   = "abcdefghijklmnopqrstuvwxyz"
	upperChars   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numericChars = "0123456789"
	specialChars = "!@#$%&*+-_=?:;,.|(){}<>"
	spaceChar    = " "
)

func StringTree() *Tree {
	return &Tree{
		Name: "String",
		nodes: []*Tree{
			StringFormat(),
			StringNumber(),
			StringKey(),
			StringEmail(),
			Uri(),
			Uris(),
			Language(),
			Error(),
			StringHash(),
			String(),
		},
	}
}

func StringNumber() *Tree {
	return &Tree{
		Name: "StringNumber",
		compare: func(r *Request) bool {
			return strings.HasSuffix(r.LastName(), "Number") &&
				r.Schema.IsString() && r.Schema.Pattern == "" && r.Schema.Format == ""
		},
		resolve: func(r *Request) (interface{}, error) {
			min := 11
			max := 11
			if r.Schema.MinLength != nil {
				min = *r.Schema.MinLength
			}
			if r.Schema.MaxLength != nil {
				max = *r.Schema.MaxLength
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
		compare: func(r *Request) bool {
			last := r.LastName()
			return (strings.ToLower(last) == "key" || strings.HasSuffix(last, "Key")) &&
				r.Schema.IsString() && r.Schema.Pattern == "" && r.Schema.Format == ""
		},
		resolve: func(r *Request) (interface{}, error) {
			return gofakeit.UUID(), nil
		},
	}
}

func StringHash() *Tree {
	hash := sha1.New()
	return &Tree{
		Name: "StringKey",
		compare: func(r *Request) bool {
			last := r.LastName()
			return (strings.ToLower(last) == "hash" || strings.HasSuffix(last, "Hash")) &&
				r.Schema.IsString() && r.Schema.Pattern == "" && r.Schema.Format == ""
		},
		resolve: func(r *Request) (interface{}, error) {
			s := gofakeit.SentenceSimple()
			b := hash.Sum([]byte(s))
			return fmt.Sprintf("%x", b), nil
		},
	}
}

func StringEmail() *Tree {
	return &Tree{
		Name: "StringKey",
		compare: func(r *Request) bool {
			last := r.LastName()
			return strings.ToLower(last) == "email" &&
				r.Schema.IsString() && r.Schema.Pattern == "" && r.Schema.Format == ""
		},
		resolve: func(r *Request) (interface{}, error) {
			return gofakeit.Email(), nil
		},
	}
}

func String() *Tree {
	return &Tree{
		Name: "String",
		compare: func(r *Request) bool {
			return r.Schema.IsString()
		},
		resolve: func(r *Request) (interface{}, error) {
			opt := StringOptions{
				MaxLength: r.Schema.MaxLength,
				Format:    r.Schema.Format,
				Pattern:   r.Schema.Pattern,
			}
			if r.Schema.MinLength != nil {
				opt.MinLength = *r.Schema.MinLength
			}
			return NewString(opt), nil
		},
	}
}

func StringFormat() *Tree {
	return &Tree{
		Name: "StringFormat",
		compare: func(r *Request) bool {
			return r.Schema.IsString() && len(r.Schema.Format) > 0
		},
		resolve: func(r *Request) (interface{}, error) {
			switch r.Schema.Format {
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
				return gofakeit.Generate(r.Schema.Format), nil
			}
		},
	}
}

func Uri() *Tree {
	return &Tree{
		Name: "URI",
		compare: func(r *Request) bool {
			if len(r.Names) == 0 || (!r.Schema.IsString() && !r.Schema.IsAny()) {
				return false
			}
			name := strings.ToLower(r.LastName())
			return len(r.Names) > 0 &&
				(name == "uri" || name == "url" ||
					strings.HasSuffix(name, "url") || strings.HasSuffix(name, "uri"))
		},
		resolve: func(r *Request) (interface{}, error) {
			return gofakeit.URL(), nil
		},
	}
}

func Uris() *Tree {
	return &Tree{
		Name: "URIs",
		compare: func(r *Request) bool {
			if len(r.Names) == 0 {
				return false
			}
			if !r.Schema.IsArray() && !r.Schema.IsAny() {
				return false
			}
			var items *schema.Ref
			if r.Schema != nil {
				items = r.Schema.Items
			}
			if !items.IsString() && !items.IsAny() {
				return false
			}

			name := strings.ToLower(r.LastName())
			return len(r.Names) > 0 &&
				(name == "uris" || name == "urls" ||
					strings.HasSuffix(name, "urls") || strings.HasSuffix(name, "uris"))
		},
		resolve: func(r *Request) (interface{}, error) {
			return r.g.tree.Resolve(r.With(Name("url")))
		},
	}
}

func Language() *Tree {
	return &Tree{
		Name: "Language",
		nodes: []*Tree{
			{
				Name: "LanguageString",
				compare: func(r *Request) bool {
					last := strings.ToLower(r.LastName())
					return (last == "language" || last == "lang") && (r.Schema.IsString() || r.Schema.IsAny())
				},
				resolve: func(r *Request) (interface{}, error) {
					return gofakeit.LanguageBCP(), nil
				},
			}, {
				Name: "Languages",
				compare: func(r *Request) bool {
					last := strings.ToLower(r.LastName())
					return (last == "languages" || last == "langs") && (r.Schema.IsArray() || r.Schema.IsAny())
				},
				resolve: func(r *Request) (interface{}, error) {
					next := r.With(Name("language"))
					if r.Schema.IsAny() {
						next = next.With(Schema(&schema.Schema{Type: []string{"array"}}))
					}
					return r.g.tree.Resolve(next)
				},
			},
		},
	}
}

func Error() *Tree {
	return &Tree{
		Name: "Error",
		compare: func(r *Request) bool {
			return strings.ToLower(r.LastName()) == "error" && (r.Schema.IsString() || r.Schema.IsAny())
		},
		resolve: func(r *Request) (interface{}, error) {
			return fmt.Sprintf("%v", gofakeit.ErrorHTTP()), nil
		},
	}
}

type StringOptions struct {
	MinLength int
	MaxLength *int
	Format    string
	Pattern   string
	Nullable  bool
}

func NewString(opt StringOptions) interface{} {
	if opt.Nullable {
		n := gofakeit.Float32Range(0, 1)
		if n < 0.05 {
			return nil
		}
	}

	if len(opt.Format) > 0 {
		return newStringByFormat(opt.Format)
	} else if len(opt.Pattern) > 0 {
		return gofakeit.Generate(fmt.Sprintf("{regex:%v}", opt.Pattern))
	}

	minLength := opt.MinLength
	maxLength := defaultMaxStringLength

	if opt.MaxLength != nil {
		maxLength = *opt.MaxLength
	} else if minLength > maxLength {
		maxLength += minLength
	}

	categories := []interface{}{0, 1, 2, 3}
	weights := []float32{5, 0.5, 0.3, 0.1}
	letters := lowerChars + upperChars

	length := gofakeit.IntRange(minLength, maxLength)
	result := make([]rune, length)
	for i := 0; i < length; i++ {
		c, _ := gofakeit.Weighted(categories, weights)

		switch c {
		case 0:
			n := gofakeit.IntRange(0, len(letters)-1)
			result[i] = rune(letters[n])
		case 1:
			n := gofakeit.IntRange(0, len(numericChars)-1)
			result[i] = rune(numericChars[n])
		case 2:
			result[i] = ' '
		case 3:
			n := gofakeit.IntRange(0, len(specialChars)-1)
			result[i] = rune(specialChars[n])
		}
	}
	return string(result)
}

func newStringByFormat(format string) string {
	switch format {
	case "date":
		return gofakeit.Date().Format("2006-01-02")
	case "date-time":
		return gofakeit.Generate("{date}")
	case "password":
		return gofakeit.Generate("{password}")
	case "email":
		return gofakeit.Generate("{email}")
	case "uuid":
		return gofakeit.Generate("{uuid}")
	case "uri":
		return gofakeit.Generate("{url}")
	case "hostname":
		return gofakeit.Generate("{domainname}")
	case "ipv4":
		return gofakeit.Generate("{ipv4address}")
	case "ipv6":
		return gofakeit.Generate("{ipv6address}")
	default:
		return gofakeit.Generate(format)
	}
}
