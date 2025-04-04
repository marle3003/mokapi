package v2

import (
	"errors"
	"github.com/jinzhu/inflection"
	"mokapi/schema/json/parser"
	"mokapi/schema/json/schema"
)

var NoMatchFound = errors.New("no match found")
var NotSupported = errors.New("not supported")

type fakeFunc func() (interface{}, error)

type Node struct {
	Name     string
	Children []*Node
	Fake     func(r *Request) (interface{}, error)
}

func NewNode(name string) *Node {
	return &Node{Name: name}
}

func fakeWalk(root *Node, r *Request) (interface{}, error) {
	for {
		if len(r.Path) == 0 {
			return nil, NoMatchFound
		}
		if v, err := root.fake(r); err == nil {
			return v()
		}
		r = r.shift()
	}
}

func (n *Node) fake(r *Request) (fakeFunc, error) {
	if len(r.Path) == 0 {
		if v, ok := applyConstraints(r); ok {
			return v, nil
		}
	}

	token := r.NextToken()
	if token == "" {
		if n.Fake != nil {
			return func() (interface{}, error) {
				v, err := n.Fake(r)
				if err != nil {
					return nil, err
				}
				return validate(v, r)
			}, nil
		}
		return nil, NotSupported
	}

	if isArrayExpected(token, r.Schema) {
		items := requestForItems(token, r)
		fakeItem, err := n.fake(items)
		if err != nil {
			return nil, err
		}
		return func() (interface{}, error) { return fakeArray(r, fakeItem) }, nil
	}

	var fake fakeFunc
	var err error
	for _, child := range n.Children {
		if child.Name == token {
			fake, err = child.fake(r.shift())
			if err == nil {
				break
			}
		}
	}
	if fake == nil {
		return nil, NoMatchFound
	}
	return fake, nil
}

func isArrayExpected(token string, s *schema.Schema) bool {
	return s.IsArray() || (s == nil && isPlural(token))
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
		newNumberNode(),
		newEmailNode(),
		newUrlNode(),
		newUriNode(),
		newDescriptionNode(),
	}

	r.Children = append(r.Children, newItNodes()...)

	return r
}
