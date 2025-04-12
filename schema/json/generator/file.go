package generator

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"strings"
)

func files() []*Node {
	return []*Node{
		{
			Name: "file",
			Children: []*Node{
				{
					Name: "name",
					Fake: fakeFileName,
				},
				{
					Name: "type",
					Fake: fakeFileType,
				},
				{
					Name: "size",
					Fake: fakeFileSize,
				},
			},
		},
	}
}

func fakeFileName(r *Request) (any, error) {
	name, err := fakeName(r)
	if err != nil {
		return nil, err
	}
	return fmt.Sprintf("%s.%s", strings.ToLower(name.(string)), gofakeit.FileExtension()), nil
}

func fakeFileType(r *Request) (any, error) {
	return gofakeit.FileMimeType(), nil
}

func fakeFileSize(r *Request) (any, error) {
	return fakeIntegerWithRange(r.Schema, 0, 100000)
}
