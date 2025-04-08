package generator

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"strings"
	"time"
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
				{
					Name: "at",
					Fake: fakeFileDate,
				},
				{
					Name: "created",
					Fake: fakeFileDate,
				},
				{
					Name: "modified",
					Fake: fakeFileDate,
				},
				{
					Name: "deleted",
					Fake: fakeFileDate,
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

func fakeFileDate(r *Request) (any, error) {
	year := time.Now().Year()
	return fakeDateInPastWithMinYear(r, year-3)
}
