package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"mokapi/schema/json/schema"
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

func fakeString(r *Request) (interface{}, error) {
	s := r.Schema
	if s.Pattern != "" {
		return fakePattern(r)
	}
	if s.MinLength == nil && s.MaxLength == nil {
		if s.Format != "" {
			return fakeFormat(s)
		}
	}

	minLength := 0
	maxLength := 15

	if s.MinLength != nil {
		minLength = *s.MinLength
		if s.MaxLength == nil && minLength > maxLength {
			maxLength = minLength + maxLength
		}
	}
	if s.MaxLength != nil {
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
}

func fakeFormat(s *schema.Schema) (interface{}, error) {
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
