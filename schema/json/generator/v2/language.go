package v2

import (
	"github.com/brianvoe/gofakeit/v6"
)

func languages() []*Node {
	return []*Node{
		{
			Name: "language",
			Fake: fakeLanguage,
		},
	}
}

func fakeLanguage(r *Request) (any, error) {
	s := r.Schema
	if s.MaxLength != nil {
		if *s.MaxLength == 2 {
			// ISO-639-1
			return gofakeit.LanguageAbbreviation(), nil
		}
		if *s.MaxLength == 5 {
			// BCP-47
			return gofakeit.LanguageBCP(), nil
		}
		if *s.MaxLength > 5 {
			// BCP-47
			return gofakeit.Language(), nil
		}
	}
	return gofakeit.LanguageAbbreviation(), nil
}
