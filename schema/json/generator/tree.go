package generator

import (
	"errors"
	"fmt"
	"github.com/jinzhu/inflection"
	"mokapi/schema/json/parser"
)

var NoMatchFound = errors.New("no match found")
var NotSupported = errors.New("not supported")

type fakeFunc func() (any, error)

type Node struct {
	Name       string                        `json:"name"`
	Attributes []string                      `json:"attributes,omitempty"`
	Weight     float64                       `json:"weight,omitempty"`
	DependsOn  []string                      `json:"dependsOn,omitempty"`
	Children   []*Node                       `json:"children,omitempty"`
	Custom     bool                          `json:"custom,omitempty"`
	Fake       func(r *Request) (any, error) `json:"-"`
}

func NewNode(name string) *Node {
	return &Node{Name: name}
}

func FindByName(name string) *Node {
	if name == "" || name == "root" {
		return g.root
	}
	return g.root.findByName(name)
}

func (n *Node) findByName(name string) *Node {
	for _, child := range n.Children {
		if child.Name == name {
			return child
		}
		if found := child.findByName(name); found != nil {
			return found
		}
	}
	return nil
}

func (n *Node) Append(child *Node) {
	n.Children = append(n.Children, child)
}

func (n *Node) Prepend(child *Node) {
	n.Children = append([]*Node{child}, n.Children...)
}

func (n *Node) RemoveAt(index int) error {
	if index < 0 {
		return fmt.Errorf("index must be positive: %v", index)
	}
	if index >= len(n.Children) {
		return fmt.Errorf("index outside of array: %v", index)
	}
	n.Children = append(n.Children[:index], n.Children[index+1:]...)
	return nil
}

func (n *Node) Remove(name string) error {
	index := -1
	for i, n := range n.Children {
		if n.Name == name {
			index = i
			break
		}
	}
	if index == -1 {
		return fmt.Errorf("name %v not found", name)
	}
	return n.RemoveAt(index)
}

func isPlural(word string) bool {
	return word == inflection.Plural(word) && word != inflection.Singular(word)
}

func validate(v any, r *Request) (any, error) {
	if r.Schema == nil {
		return v, nil
	}
	p := parser.Parser{Schema: r.Schema, ConvertStringToNumber: true, ConvertStringToBoolean: true}
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
	r.Children = append(r.Children, dates()...)
	r.Children = append(r.Children, ictNodes()...)
	r.Children = append(r.Children, textNodes()...)
	r.Children = append(r.Children, personal...)
	r.Children = append(r.Children, addresses()...)
	r.Children = append(r.Children, locations()...)
	r.Children = append(r.Children, financials()...)
	r.Children = append(r.Children, colors()...)
	r.Children = append(r.Children, languages()...)
	r.Children = append(r.Children, pets()...)
	r.Children = append(r.Children, metadata()...)
	r.Children = append(r.Children, products()...)
	r.Children = append(r.Children, files()...)
	r.Children = append(r.Children, companyNodes()...)

	return r
}
