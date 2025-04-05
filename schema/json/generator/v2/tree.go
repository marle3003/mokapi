package v2

import (
	"errors"
	"github.com/jinzhu/inflection"
	"mokapi/schema/json/parser"
)

var NoMatchFound = errors.New("no match found")
var NotSupported = errors.New("not supported")

type fakeFunc func() (interface{}, error)

type Node struct {
	Name      string
	Weight    float64
	DependsOn []string
	Children  []*Node
	Fake      func(r *Request) (interface{}, error)
}

func NewNode(name string) *Node {
	return &Node{Name: name}
}

func isPlural(word string) bool {
	return word == inflection.Plural(word) && word != inflection.Singular(word)
}

func validate(v interface{}, r *Request) (interface{}, error) {
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

	r.Children = append(r.Children, numberNodes()...)
	r.Children = append(r.Children, newItNodes()...)
	r.Children = append(r.Children, newNumberNodes()...)
	r.Children = append(r.Children, newTextNodes()...)
	r.Children = append(r.Children, personal()...)

	return r
}
