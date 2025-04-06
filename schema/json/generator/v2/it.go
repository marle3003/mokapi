package v2

import (
	"crypto/sha1"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"strings"
)

func ictNodes() []*Node {
	return []*Node{
		newErrorNode(),
		newHashNode(),
		{Name: "username", Fake: fakeUsername},
		{
			Name: "user",
			Fake: fakeUser,
			Children: []*Node{
				{Name: "name", Fake: fakeUsername},
			},
		},
	}
}

func newErrorNode() *Node {
	return &Node{Name: "error", Fake: fakeError}
}

func fakeError(r *Request) (interface{}, error) {
	return gofakeit.Error().Error(), nil
}

func newHashNode() *Node {
	return &Node{Name: "hash", Fake: fakeHash}
}

func fakeHash(_ *Request) (interface{}, error) {
	hash := sha1.New()
	s := gofakeit.SentenceSimple()
	b := hash.Sum([]byte(s))
	return fmt.Sprintf("%x", b), nil
}

func fakeUsername(_ *Request) (interface{}, error) {
	return gofakeit.Username(), nil
}

func fakeUser(r *Request) (interface{}, error) {
	s := r.Schema
	if s.IsString() {
		return gofakeit.Username(), nil
	}
	firstname := gofakeit.FirstName()
	lastname := gofakeit.LastName()
	first := strings.ToLower(firstname)
	last := strings.ToLower(lastname)
	return map[string]interface{}{
		"firstname": firstname,
		"lastname":  lastname,
		"gender":    gofakeit.Gender(),
		"email":     fmt.Sprintf("%s.%s@%s", first, last, gofakeit.DomainName()),
		"username":  fmt.Sprintf("%c%s", first[0], last),
	}, nil
}
