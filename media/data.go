package media

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/brianvoe/gofakeit/v6/data"
)

var faker = gofakeit.New(0)
var types map[string][]ContentType

func GetRandom(accept string) ContentType {
	a := ParseContentType(accept)
	if a.Type == "*" {
		m := gofakeit.FileMimeType()
		return ParseContentType(m)
	}

	subtypes := types[a.Type]
	i := faker.Rand.Intn(len(subtypes))
	return subtypes[i]
}

func SetFaker(seed int64) {
	faker = gofakeit.New(seed)
}

func init() {
	types = make(map[string][]ContentType)
	for _, t := range data.Files["mime_type"] {
		ct := ParseContentType(t)
		types[ct.Type] = append(types[ct.Type], ct)
	}
}
