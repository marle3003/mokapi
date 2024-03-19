package generator

import "github.com/brianvoe/gofakeit/v6"

func Language() *Tree {
	return &Tree{
		Name: "Language",
		Test: func(r *Request) bool {
			last := r.Last()
			return (last.Name == "language" || last.Name == "lang") &&
				((last.Schema.IsString() && !hasFormat(last.Schema) && !hasPattern(last.Schema)) || last.Schema.IsAny())
		},
		Nodes: []*Tree{
			LanguageIso639_1(),
			LanguageBCP47(),
			LanguageLong(),
		},
	}
}

func LanguageIso639_1() *Tree {
	return &Tree{
		Name: "ISO-639-1",
		Test: func(r *Request) bool {
			s := r.LastSchema()
			return s.MaxLength == nil || *s.MaxLength == 2
		},
		Fake: func(r *Request) (interface{}, error) {
			return gofakeit.LanguageAbbreviation(), nil
		},
	}
}

func LanguageBCP47() *Tree {
	return &Tree{
		Name: "BCP-47",
		Test: func(r *Request) bool {
			s := r.LastSchema()
			return s.MaxLength != nil && *s.MaxLength == 5
		},
		Fake: func(r *Request) (interface{}, error) {
			return gofakeit.LanguageBCP(), nil
		},
	}
}

func LanguageLong() *Tree {
	return &Tree{
		Name: "BCP-47",
		Test: func(r *Request) bool {
			s := r.LastSchema()
			return s.MaxLength != nil && *s.MaxLength > 5
		},
		Fake: func(r *Request) (interface{}, error) {
			return gofakeit.Language(), nil
		},
	}
}
