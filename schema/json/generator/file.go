package generator

import (
	"fmt"
	"strings"

	"github.com/brianvoe/gofakeit/v7"
)

func files() []*Node {
	return []*Node{
		{
			Name:       "file",
			Attributes: []string{"file"},
			Children: []*Node{
				{
					Name:       "name",
					Attributes: []string{"name"},
					Fake:       fakeFileName,
				},
				{
					Name:       "type",
					Attributes: []string{"type"},
					Fake:       fakeFileType,
				},
				{
					Name:       "size",
					Attributes: []string{"size"},
					Fake:       fakeFileSize,
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
