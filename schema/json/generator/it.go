package generator

import (
	"crypto/sha1"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"strings"
)

func It() *Tree {
	return &Tree{
		Name: "IT",
		Nodes: []*Tree{
			StringId(),
			StringEmail(),
			StringHash(),
			Uri(),
			Username(),
			Error(),
		},
	}
}

func Username() *Tree {
	return &Tree{
		Name: "Username",
		Test: func(r *Request) bool {
			last := r.Last()
			if !last.Schema.IsString() {
				return false
			}
			return strings.ToLower(last.Name) == "username" || strings.HasSuffix(last.Name, "UserName") ||
				strings.HasSuffix(last.Name, "Username")
		},
		Fake: func(r *Request) (interface{}, error) {
			return gofakeit.Username(), nil
		},
	}
}

func Error() *Tree {
	return &Tree{
		Name: "Error",
		Test: func(r *Request) bool {
			last := r.Last()
			return strings.ToLower(last.Name) == "error" && (last.Schema.IsAnyString() || last.Schema.IsAny())
		},
		Fake: func(r *Request) (interface{}, error) {
			return fmt.Sprintf("%v", gofakeit.Error()), nil
		},
	}
}

func StringHash() *Tree {
	hash := sha1.New()
	return &Tree{
		Name: "StringHash",
		Test: func(r *Request) bool {
			last := r.Last()
			return (strings.ToLower(last.Name) == "hash" || strings.HasSuffix(last.Name, "Hash")) &&
				last.Schema.IsAnyString()
		},
		Fake: func(r *Request) (interface{}, error) {
			s := gofakeit.SentenceSimple()
			b := hash.Sum([]byte(s))
			return fmt.Sprintf("%x", b), nil
		},
	}
}
