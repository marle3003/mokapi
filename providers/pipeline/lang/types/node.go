package types

import (
	"fmt"
	"reflect"
	"strings"
)

type Node struct {
	attributes map[string]string
	children   []*Node
	name       string
}

func NewNode(name string) *Node {
	return &Node{name: name}
}

func (n *Node) GetField(name string) (Object, error) {
	return getField(n, name)
}

func (n *Node) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("<%v", n.name))
	i := 0
	for k, v := range n.attributes {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%v=\"%v\"", k, v))
	}
	if len(n.children) == 0 {
		sb.WriteString(" />")
	} else {
		sb.WriteString(">")

		for _, c := range n.children {
			sb.WriteString(c.String())
		}

		sb.WriteString(fmt.Sprintf("</%v>", n.name))
	}

	return sb.String()
}

func (n *Node) Add(obj Object) {
	if node, ok := obj.(*Node); ok {
		n.children = append(n.children, node)
	}
}

func (n *Node) Children() *Array {
	a := NewArray()
	for _, c := range n.children {
		a.Add(c)
	}
	return a
}

func (n *Node) Find(match Predicate) (Object, error) {
	for _, item := range n.children {
		if matches, err := match(item); err == nil {
			if matches {
				return item, nil
			}
		} else {
			return nil, err
		}
	}
	return nil, nil
}

func (n *Node) FindAll(match Predicate) ([]Object, error) {
	result := make([]Object, 0)
	for _, c := range n.children {
		if matches, err := match(c); err == nil {
			if matches {
				result = append(result, c)
			}
		} else {
			return nil, err
		}
	}
	return result, nil
}

func (n *Node) GetType() reflect.Type {
	return reflect.TypeOf(n)
}

func (n *Node) depthFirst() chan Object {
	ch := make(chan Object)
	go func() {
		defer close(ch)

		for _, c := range n.children {
			for o := range c.depthFirst() {
				ch <- o
			}
			ch <- c
		}
	}()
	return ch
}
