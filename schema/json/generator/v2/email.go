package v2

import "github.com/brianvoe/gofakeit/v6"

func newEmailNode() *Node {
	return &Node{Name: "email", Fake: fakeEmail}
}

func fakeEmail(_ *Request) (interface{}, error) {
	return gofakeit.Email(), nil
}
