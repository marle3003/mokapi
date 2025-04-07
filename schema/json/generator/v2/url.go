package v2

import "github.com/brianvoe/gofakeit/v6"

func newUrlNode() *Node {
	return &Node{Name: "url", Fake: fakeUrl}
}

func newUriNode() *Node {
	return &Node{Name: "uri", Fake: fakeUrl}
}

func fakeUrl(_ *Request) (interface{}, error) {
	return gofakeit.URL(), nil
}
