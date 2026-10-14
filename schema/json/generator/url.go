package generator

import "github.com/brianvoe/gofakeit/v7"

func newUrlNode() *Node {
	return &Node{
		Name:       "url",
		Attributes: []string{"url", "uri"},
		Fake:       fakeUrl,
	}
}

func fakeUrl(_ *Request) (any, error) {
	return gofakeit.URL(), nil
}

func fakeWebsite(_ *Request) (any, error) {
	return gofakeit.DomainName(), nil
}
