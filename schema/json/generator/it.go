package generator

import (
	"crypto/sha1"
	"fmt"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v7"
)

func ictNodes() []*Node {
	return []*Node{
		newErrorNode(),
		newHashNode(),
		{
			Name:       "username",
			Attributes: []string{"username"},
			DependsOn:  []string{"firstname", "lastname"},
			Fake:       fakeUsername,
		},
		{
			Name:       "user",
			Attributes: []string{"user"},
			DependsOn:  []string{"firstname", "lastname"},
			Fake:       fakeUser,
		},
		{
			Name:       "website",
			Attributes: []string{"website"},
			Fake:       fakeWebsite,
		},
		{
			Name:       "role",
			Attributes: []string{"role"},
			Fake:       fakeRole,
		},
		{
			Name:       "permission",
			Attributes: []string{"permission"},
			Fake:       fakePermission,
		},
		{
			Name:       "lastlogin",
			Attributes: []string{"lastLogin"},
			Fake:       fakeLastLogin,
		},
		{
			Name:       "password",
			Attributes: []string{"password"},
			Fake:       fakePassword,
		},
	}
}

func newErrorNode() *Node {
	return &Node{
		Name:       "error",
		Attributes: []string{"error"},
		Fake:       fakeError,
	}
}

func fakeError(_ *Request) (any, error) {
	return gofakeit.Error().Error(), nil
}

func newHashNode() *Node {
	return &Node{
		Name:       "hash",
		Attributes: []string{"hash"},
		Fake:       fakeHash,
	}
}

func fakeHash(_ *Request) (any, error) {
	hash := sha1.New()
	s := gofakeit.Sentence()
	b := hash.Sum([]byte(s))
	return fmt.Sprintf("%x", b), nil
}

func fakeUsername(r *Request) (any, error) {
	var err error

	var first string
	if v, ok := r.Context.Values["firstname"]; ok {
		first = v.(string)
	} else {
		v, err = fakeFirstname(r)
		if err != nil {
			return nil, err
		}
		first = v.(string)
	}

	var last string
	if v, ok := r.Context.Values["lastname"]; ok {
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

func fakeUser(r *Request) (any, error) {
	s := r.Schema
	if s.IsString() {
		return fakeUsername(r)
	}
	firstname := gofakeit.FirstName()
	lastname := gofakeit.LastName()
	first := strings.ToLower(firstname)
	last := strings.ToLower(lastname)
	return map[string]any{
		"firstname": firstname,
		"lastname":  lastname,
		"gender":    gofakeit.Gender(),
		"email":     fmt.Sprintf("%s.%s@%s", first, last, gofakeit.DomainName()),
		"username":  fmt.Sprintf("%c%s", first[0], last),
	}, nil
}

func fakeRole(_ *Request) (any, error) {
	index := gofakeit.Number(0, len(roles)-1)
	return roles[index], nil
}

func fakePermission(_ *Request) (any, error) {
	index := gofakeit.Number(0, len(permissions)-1)
	return permissions[index], nil
}

func fakeLastLogin(r *Request) (any, error) {
	year := time.Now().Year()
	return fakeDateInPastWithMinYear(r, year-1)
}

func fakePassword(r *Request) (any, error) {
	return gofakeit.Password(true, true, true, true, false, 11), nil
}

var roles = []string{
	"admin", "user", "guest", "owner", "editor", "viewer",
}

var permissions = []string{
	"read", "create", "update", "delete",
}
