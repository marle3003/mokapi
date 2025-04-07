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
	first := r.ctx.store["firstname"]
	if first != nil {
		choosePersonEmail = true
	}
	last := r.ctx.store["lastname"]
	if last != nil {
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
