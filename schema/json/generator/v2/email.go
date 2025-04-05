package v2

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
	first := r.ctx["firstname"]
	if first != nil {
		choosePersonEmail = true
	}
	last := r.ctx["lastname"]
	if last != nil {
		choosePersonEmail = true

	}

	if choosePersonEmail {
		if first == "" {
			first = gofakeit.FirstName()
		}
		if last == "" {
			last = gofakeit.LastName()
		}
		return strings.ToLower(fmt.Sprintf("%s.%s@%s", first, last, gofakeit.DomainName())), nil
	}
	return gofakeit.Email(), nil
}
