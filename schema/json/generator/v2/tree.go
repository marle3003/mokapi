package v2

import (
	"errors"
	"github.com/jinzhu/inflection"
	"mokapi/schema/json/parser"
)

var NoMatchFound = errors.New("no match found")
var NotSupported = errors.New("not supported")

type fakeFunc func() (any, error)

type Node struct {
	Name      string
	Weight    float64
	DependsOn []string
	Children  []*Node
	Fake      func(r *Request) (any, error)
}

func NewNode(name string) *Node {
	return &Node{Name: name}
}

func isPlural(word string) bool {
	return word == inflection.Plural(word) && word != inflection.Singular(word)
}

func validate(v any, r *Request) (any, error) {
	if r.Schema == nil {
		return v, nil
	}
	p := parser.Parser{Schema: r.Schema}
	return p.Parse(v)
}

func buildTree() *Node {
	r := NewNode("")
	r.Children = []*Node{
		newNameNode(),
		newIdNode(),
		newKeyNode(),
		newEmailNode(),
		newUrlNode(),
		newUriNode(),
	}

	r.Children = append(r.Children, numbers()...)
	r.Children = append(r.Children, ictNodes()...)
	r.Children = append(r.Children, textNodes()...)
	r.Children = append(r.Children, personal()...)
	r.Children = append(r.Children, addresses()...)
	r.Children = append(r.Children, locations()...)
	r.Children = append(r.Children, financials()...)

	return r
}
