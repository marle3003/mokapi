package generator

import (
	"fmt"
	"strings"

	"github.com/brianvoe/gofakeit/v7"
)

func newEmailNode() *Node {
	return &Node{
		Name:       "email",
		Attributes: []string{"email"},
		DependsOn:  []string{"firstname", "lastname"},
		Fake:       fakeEmail,
	}
}

func fakeEmail(r *Request) (any, error) {
	choosePersonEmail := false
	first, ok := r.Context.Values["firstname"]
	if ok {
		choosePersonEmail = true
	}
	last, ok := r.Context.Values["lastname"]
	if ok {
		choosePersonEmail = true

	}

	if choosePersonEmail {
		if first == nil {
			first, _ = fakeFirstname(r)
		}
		if last == nil {
			last, _ = fakeLastname(r)
		}
		return strings.ToLower(fmt.Sprintf("%s.%s@%s", first, last, gofakeit.DomainName())), nil
	}
	return gofakeit.Email(), nil
}
