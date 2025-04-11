package generator

import (
	"crypto/sha1"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"strings"
	"time"
)

func ictNodes() []*Node {
	return []*Node{
		newErrorNode(),
		newHashNode(),
		{
			Name:      "username",
			DependsOn: []string{"firstname", "lastname"},
			Fake:      fakeUsername,
		},
		{
			Name: "user",
			Fake: fakeUser,
			Children: []*Node{
				{
					Name: "name",
					Fake: fakeUsername,
				},
			},
		},
		{
			Name: "website",
			Fake: fakeWebsite,
		},
		{
			Name: "role",
			Fake: fakeRole,
		},
		{
			Name: "permission",
			Fake: fakePermission,
		},
		{
			Name: "last",
			Children: []*Node{
				{
					Name: "login",
					Fake: fakeLastLogin,
				},
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

func fakeUsername(r *Request) (interface{}, error) {
	var err error

	var first string
	if v, ok := r.ctx.store["firstname"]; ok {
		first = v.(string)
	} else {
		v, err = fakeFirstname(r)
		if err != nil {
			return nil, err
		}
		first = v.(string)
	}

	var last string
	if v, ok := r.ctx.store["lastname"]; ok {
		last = v.(string)
	} else {
		v, err = fakeLastname(r)
		if err != nil {
			return nil, err
		}
		last = v.(string)
	}

	first = strings.ToLower(first)
	last = strings.ToLower(last)

	return fmt.Sprintf("%c%s", first[0], last), nil
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

func fakeRole(_ *Request) (interface{}, error) {
	index := gofakeit.Number(0, len(roles)-1)
	return roles[index], nil
}

func fakePermission(_ *Request) (interface{}, error) {
	index := gofakeit.Number(0, len(permissions)-1)
	return permissions[index], nil
}

func fakeLastLogin(r *Request) (interface{}, error) {
	year := time.Now().Year()
	return fakeDateInPastWithMinYear(r, year-1)
}

var roles = []string{
	"admin", "user", "guest", "owner", "editor", "viewer",
}

var permissions = []string{
	"read", "create", "update", "delete",
}
