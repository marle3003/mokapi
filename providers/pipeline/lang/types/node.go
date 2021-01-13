package types

import (
	"fmt"
	"github.com/pkg/errors"
	"mokapi/providers/encoding"
	"reflect"
	"strings"
)

type Node struct {
	ObjectImpl
	attributes *Expando
	children   *Array
	name       string
	content    string
}

func NewNode(name string) *Node {
	return &Node{name: name, attributes: NewExpando(), children: NewArray()}
}

func ParseNode(node *encoding.XmlNode) *Node {
	n := NewNode(node.Name)
	n.content = node.Content
	for k, v := range node.Attributes {
		n.attributes.value[k] = NewString(v)
	}

	for _, c := range node.Children {
		n.children.Add(ParseNode(c))
	}
	return n
}

func (n *Node) GetField(name string) (Object, error) {
	if strings.HasPrefix(name, "@") {
		name = name[1:]
		return n.attributes.GetField(name)
	} else {
		for _, i := range n.children.value {
			if c, ok := i.(*Node); ok && c.name == name {
				return c, nil
			} else if !ok {
				return nil, errors.Errorf("unexpected type of child node: %v", reflect.TypeOf(i))
			}
		}
	}

	switch strings.ToLower(name) {
	case "name":
		return n.Name(), nil
	case "text":
		return n.Text(), nil
	case "children":
		return n.children, nil
	case "attributes":
		return n.children, nil
	}

	return getField(n, name)
}

func (n *Node) Index(index int) (Object, error) {
	return n.children.Index(index)
}

func (n *Node) InvokeFunc(name string, args map[string]Object) (Object, error) {
	return invokeFunc(n, name, args)
}

func (n *Node) String() string {
	sb := strings.Builder{}
	hasName := len(n.name) > 0
	if hasName {
		sb.WriteString(fmt.Sprintf("<%v", n.name))
	}
	for k, v := range n.attributes.value {
		sb.WriteString(fmt.Sprintf(" %v=\"%v\"", k, v))
	}
	if len(n.children.value) == 0 && len(n.content) == 0 && hasName {
		sb.WriteString(" />")
	} else {
		if hasName {
			sb.WriteString(">")
		}

		for _, c := range n.children.value {
			sb.WriteString(c.String())
		}

		sb.WriteString(n.content)

		if hasName {
			sb.WriteString(fmt.Sprintf("</%v>", n.name))
		}
	}

	return sb.String()
}

func (n *Node) Add(obj Object) {
	if node, ok := obj.(*Node); ok {
		n.children.Add(node)
	}
}

func (n *Node) Children() *Array {
	a := NewArray()
	for _, c := range n.children.value {
		a.Add(c)
	}
	return a
}

func (n *Node) Find(match Predicate) (Object, error) {
	for _, item := range n.children.value {
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
	for _, c := range n.children.value {
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

func (n *Node) Name() *String {
	return NewString(n.name)
}

func (n *Node) Text() *String {
	var sb strings.Builder
	for _, i := range n.children.value {
		if c, ok := i.(*Node); ok {
			sb.WriteString(c.Text().String())
		}
	}
	sb.WriteString(n.content)
	return NewString(sb.String())
}
