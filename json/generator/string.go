package generator

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
)

const (
	defaultMinStringLength = 0
	defaultMaxStringLength = 15

	lowerChars   = "abcdefghijklmnopqrstuvwxyz"
	upperChars   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numericChars = "0123456789"
	specialChars = "!@#$%&*+-_=?:;,.|(){}<>"
	spaceChar    = " "
)

type StringOptions struct {
	MinLength *int
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

	minLength := defaultMinStringLength
	maxLength := defaultMaxStringLength

	if opt.MinLength != nil {
		minLength = *opt.MinLength
	}
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
