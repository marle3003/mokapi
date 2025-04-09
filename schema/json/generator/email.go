package generator

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"strings"
)

func newEmailNode() *Node {
	return &Node{
		Name:      "email",
		DependsOn: []string{"firstname", "lastname"},
		Fake:      fakeEmail,
	}
}

func fakeEmail(r *Request) (interface{}, error) {
	choosePersonEmail := false
	first, ok := r.ctx.store["firstname"]
	if ok {
		choosePersonEmail = true
	}
	last, ok := r.ctx.store["lastname"]
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
